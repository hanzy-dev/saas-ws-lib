package httpx

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"

	wsctx "github.com/hanzy-dev/saas-ws-lib/pkg/ctx"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

const HeaderRequestID = "X-Request-ID"

const maxRetryCap = 3

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
	if cfg.MaxRetries > maxRetryCap {
		cfg.MaxRetries = maxRetryCap
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

func Do(ctx context.Context, c *http.Client, req *http.Request, retries int) (*http.Response, error) {
	if c == nil {
		return nil, errors.New("httpx: nil client")
	}
	if req == nil {
		return nil, errors.New("httpx: nil request")
	}

	if retries < 0 {
		retries = 0
	}
	if retries > maxRetryCap {
		retries = maxRetryCap
	}

	req = req.WithContext(ctx)

	var lastErr error

	for attempt := 0; attempt <= retries; attempt++ {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		if attempt > 0 {
			if err := rewindBody(req); err != nil {
				return nil, err
			}
			sleep(ctx, backoff(attempt))
		}

		resp, err := c.Do(req)
		if err == nil {
			if shouldRetryStatus(resp) && attempt < retries && isIdempotent(req.Method) {
				lastErr = fmt.Errorf("httpx: upstream %d", resp.StatusCode)
				_ = resp.Body.Close()
				continue
			}
			return resp, nil
		}

		lastErr = err

		if attempt == retries {
			break
		}
		if !isIdempotent(req.Method) {
			break
		}
		if !isRetryableNetErr(err) {
			break
		}
	}

	return nil, fmt.Errorf("httpx: request failed after %d attempts: %w", retries+1, lastErr)
}

func sleep(ctx context.Context, d time.Duration) {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
	case <-t.C:
	}
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

func shouldRetryStatus(resp *http.Response) bool {
	switch resp.StatusCode {
	case http.StatusBadGateway,
		http.StatusServiceUnavailable,
		http.StatusGatewayTimeout:
		return true
	default:
		return false
	}
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
