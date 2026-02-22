package httpx

import (
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"time"

	wsctx "github.com/hanzy-dev/saas-ws-lib/pkg/ctx"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

const HeaderRequestID = "X-Request-ID"

type ClientConfig struct {
	Timeout time.Duration

	MaxRetries int

	Transport http.RoundTripper

	DisableRequestIDPropagation bool
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

	rt := cfg.Transport
	if !cfg.DisableRequestIDPropagation {
		rt = &requestIDTransport{base: rt}
	}
	rt = otelhttp.NewTransport(rt)

	return &http.Client{
		Timeout:   cfg.Timeout,
		Transport: rt,
	}
}

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
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		if attempt > 0 {
			if err := rewindBody(req); err != nil {
				return nil, err
			}
			d := backoff(attempt)
			t := time.NewTimer(d)
			select {
			case <-ctx.Done():
				t.Stop()
				return nil, ctx.Err()
			case <-t.C:
			}
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
	if errors.As(err, &ne) && ne.Timeout() {
		return true
	}
	if errors.Is(err, io.EOF) {
		return true
	}
	return false
}

func backoff(attempt int) time.Duration {
	d := time.Duration(attempt) * 150 * time.Millisecond
	if d > 800*time.Millisecond {
		d = 800 * time.Millisecond
	}
	return d
}

func rewindBody(req *http.Request) error {
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
