package httpx

import (
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"time"

	wsctx "saas-ws-lib/pkg/ctx"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

type ClientConfig struct {
	Timeout time.Duration

	// MaxRetries is applied only for idempotent methods and network/timeout errors.
	// Default: 1
	MaxRetries int

	// Transport is optional. If nil, a sane default transport is used.
	Transport http.RoundTripper

	// PropagateRequestID copies X-Request-ID from ctx into outbound request header when missing.
	// Default: true
	PropagateRequestID bool
}

func NewClient(cfg ClientConfig) *http.Client {
	if cfg.Timeout == 0 {
		cfg.Timeout = 10 * time.Second
	}
	if cfg.MaxRetries < 0 {
		cfg.MaxRetries = 0
	}
	if cfg.MaxRetries == 0 {
		cfg.MaxRetries = 1
	}
	if cfg.Transport == nil {
		cfg.Transport = defaultTransport()
	}
	if !cfg.PropagateRequestID {
		// leave it false explicitly
	} else {
		cfg.PropagateRequestID = true
	}

	rt := cfg.Transport
	if cfg.PropagateRequestID {
		rt = &requestIDTransport{base: rt}
	}
	// OTel propagation + spans for outbound HTTP calls.
	rt = otelhttp.NewTransport(rt)

	return &http.Client{
		Timeout:   cfg.Timeout,
		Transport: rt,
	}
}

// Do sends an HTTP request with safe retry semantics:
// - retries only on network/timeout errors
// - retries only for idempotent methods (GET/HEAD/PUT/DELETE) and when body is replayable
func Do(ctx context.Context, c *http.Client, req *http.Request, maxRetries int) (*http.Response, error) {
	if c == nil {
		return nil, errors.New("httpx: nil client")
	}
	if req == nil {
		return nil, errors.New("httpx: nil request")
	}

	req = req.WithContext(ctx)

	var lastErr error
	for attempt := 0; attempt <= maxRetries; attempt++ {
		// ensure body can be re-sent if we retry
		if attempt > 0 {
			if err := rewindBody(req); err != nil {
				return nil, err
			}
			time.Sleep(backoff(attempt))
		}

		resp, err := c.Do(req)
		if err == nil {
			return resp, nil
		}
		lastErr = err

		if attempt == maxRetries {
			break
		}
		if !isIdempotent(req.Method) {
			break
		}
		if !isRetryableNetErr(err) {
			break
		}
	}

	return nil, lastErr
}

func defaultTransport() *http.Transport {
	return &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   5 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   5 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		ForceAttemptHTTP2:     true,
	}
}

func isIdempotent(method string) bool {
	switch method {
	case http.MethodGet, http.MethodHead, http.MethodPut, http.MethodDelete:
		return true
	default:
		return false
	}
}

func isRetryableNetErr(err error) bool {
	var ne net.Error
	if errors.As(err, &ne) && (ne.Timeout() || ne.Temporary()) {
		return true
	}
	// common network error cases
	if errors.Is(err, io.EOF) {
		return true
	}
	return false
}

func backoff(attempt int) time.Duration {
	// simple bounded linear backoff, avoids noise
	d := time.Duration(attempt) * 150 * time.Millisecond
	if d > 800*time.Millisecond {
		d = 800 * time.Millisecond
	}
	return d
}

func rewindBody(req *http.Request) error {
	// No body or body already replayable via GetBody
	if req.Body == nil || req.Body == http.NoBody {
		return nil
	}
	if req.GetBody == nil {
		return errors.New("httpx: request body is not replayable for retry (missing GetBody)")
	}
	b, err := req.GetBody()
	if err != nil {
		return err
	}
	req.Body = b
	return nil
}

type requestIDTransport struct {
	base http.RoundTripper
}

func (t *requestIDTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	if t.base == nil {
		t.base = http.DefaultTransport
	}
	if r.Header.Get(HeaderRequestID) == "" {
		if rid := wsctx.RequestID(r.Context()); rid != "" {
			r2 := r.Clone(r.Context())
			r2.Header = cloneHeader(r.Header)
			r2.Header.Set(HeaderRequestID, rid)
			return t.base.RoundTrip(r2)
		}
	}
	return t.base.RoundTrip(r)
}

func cloneHeader(h http.Header) http.Header {
	cp := make(http.Header, len(h))
	for k, v := range h {
		vv := make([]string, len(v))
		copy(vv, v)
		cp[k] = vv
	}
	return cp
}
