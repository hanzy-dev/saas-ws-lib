package httpx

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	wsctx "github.com/hanzy-dev/saas-ws-lib/pkg/ctx"
)

type timeoutNetErr struct{}

func (timeoutNetErr) Error() string   { return "timeout" }
func (timeoutNetErr) Timeout() bool   { return true }
func (timeoutNetErr) Temporary() bool { return true }

type rtFunc func(*http.Request) (*http.Response, error)

func (f rtFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestNewClient_NormalizesConfig(t *testing.T) {
	t.Parallel()

	c := NewClient(ClientConfig{
		Timeout:    0,
		MaxRetries: 999,
		Transport:  nil,
	})
	if c.Timeout != 10*time.Second {
		t.Fatalf("timeout=%s", c.Timeout)
	}
	// transport should be set (wrapped)
	if c.Transport == nil {
		t.Fatalf("transport must not be nil")
	}

	c2 := NewClient(ClientConfig{MaxRetries: -10})
	_ = c2
}

func TestDo_Validation(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest(http.MethodGet, "http://example.invalid", nil)

	if _, err := Do(context.Background(), nil, req, 1); err == nil {
		t.Fatalf("expected error for nil client")
	}
	if _, err := Do(context.Background(), &http.Client{}, nil, 1); err == nil {
		t.Fatalf("expected error for nil request")
	}
}

func TestDo_RetryOnRetryableStatus_Idempotent(t *testing.T) {
	t.Parallel()

	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if atomic.AddInt32(&calls, 1) == 1 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	client := NewClient(ClientConfig{Timeout: 2 * time.Second})
	req, _ := http.NewRequest(http.MethodGet, srv.URL, nil)

	resp, err := Do(context.Background(), client, req, 2)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	_ = resp.Body.Close()

	if atomic.LoadInt32(&calls) < 2 {
		t.Fatalf("expected retry on 503")
	}
}

func TestDo_NoRetryOnRetryableStatus_NonIdempotent(t *testing.T) {
	t.Parallel()

	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.WriteHeader(http.StatusBadGateway)
	}))
	defer srv.Close()

	client := NewClient(ClientConfig{Timeout: 2 * time.Second})
	req, _ := http.NewRequest(http.MethodPost, srv.URL, nil)

	resp, err := Do(context.Background(), client, req, 2)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	_ = resp.Body.Close()

	if atomic.LoadInt32(&calls) != 1 {
		t.Fatalf("POST should not retry, calls=%d", calls)
	}
}

func TestDo_RetryableNetErr_Idempotent(t *testing.T) {
	t.Parallel()

	var calls int32
	client := &http.Client{
		Transport: rtFunc(func(r *http.Request) (*http.Response, error) {
			if atomic.AddInt32(&calls, 1) == 1 {
				return nil, timeoutNetErr{}
			}
			return &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(bytes.NewBufferString("ok")),
				Header:     make(http.Header),
				Request:    r,
			}, nil
		}),
		Timeout: 2 * time.Second,
	}

	req := httptest.NewRequest(http.MethodGet, "http://example.invalid", nil)

	resp, err := Do(context.Background(), client, req, 1)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	_ = resp.Body.Close()

	if atomic.LoadInt32(&calls) < 2 {
		t.Fatalf("expected retry on timeout net.Error")
	}
}

func TestDo_NoRetry_NonRetryableError(t *testing.T) {
	t.Parallel()

	var calls int32
	client := &http.Client{
		Transport: rtFunc(func(r *http.Request) (*http.Response, error) {
			atomic.AddInt32(&calls, 1)
			return nil, errors.New("boom")
		}),
	}

	req := httptest.NewRequest(http.MethodGet, "http://example.invalid", nil)
	_, _ = Do(context.Background(), client, req, 2)

	if atomic.LoadInt32(&calls) != 1 {
		t.Fatalf("expected no retry for non-retryable error, calls=%d", calls)
	}
}

func TestDo_RewindBody_ErrorWhenGetBodyMissing(t *testing.T) {
	t.Parallel()

	// transport always fails with timeout => would retry, but rewindBody should fail first on attempt>0
	var calls int32
	client := &http.Client{
		Transport: rtFunc(func(r *http.Request) (*http.Response, error) {
			atomic.AddInt32(&calls, 1)
			return nil, timeoutNetErr{}
		}),
	}

	req := httptest.NewRequest(http.MethodPut, "http://example.invalid", io.NopCloser(bytes.NewBufferString("x")))
	// No GetBody => not replayable

	_, err := Do(context.Background(), client, req, 1)
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestIsIdempotentAndRetryStatusHelpers(t *testing.T) {
	t.Parallel()

	if !isIdempotent(http.MethodGet) || !isIdempotent(http.MethodPut) {
		t.Fatalf("expected GET/PUT idempotent")
	}
	if isIdempotent(http.MethodPost) {
		t.Fatalf("POST must not be idempotent")
	}

	resp := &http.Response{StatusCode: http.StatusBadGateway}
	if !shouldRetryStatus(resp) {
		t.Fatalf("expected retry for 502")
	}
	resp.StatusCode = 418
	if shouldRetryStatus(resp) {
		t.Fatalf("did not expect retry for 418")
	}
}

func TestBackoffCaps(t *testing.T) {
	t.Parallel()

	if backoff(1) <= 0 {
		t.Fatalf("expected positive backoff")
	}
	if backoff(999) != 800*time.Millisecond {
		t.Fatalf("expected cap at 800ms, got %s", backoff(999))
	}
}

func TestIsRetryableNetErr(t *testing.T) {
	t.Parallel()

	var ne net.Error = timeoutNetErr{}
	if !isRetryableNetErr(ne) {
		t.Fatalf("expected retryable for net.Error timeout")
	}
	if !isRetryableNetErr(io.EOF) {
		t.Fatalf("expected retryable for io.EOF")
	}
	if isRetryableNetErr(errors.New("x")) {
		t.Fatalf("did not expect retryable for generic error")
	}
}

func TestRequestIDTransport_SetsHeaderWhenMissing(t *testing.T) {
	t.Parallel()

	var got string
	base := rtFunc(func(r *http.Request) (*http.Response, error) {
		got = r.Header.Get(HeaderRequestID)
		return &http.Response{
			StatusCode: 200,
			Body:       io.NopCloser(bytes.NewBufferString("ok")),
			Header:     make(http.Header),
			Request:    r,
		}, nil
	})

	rt := &requestIDTransport{base: base}

	ctx := wsctx.WithRequestID(context.Background(), "rid-1")
	req := httptest.NewRequest(http.MethodGet, "http://example.invalid", nil).WithContext(ctx)

	resp, err := rt.RoundTrip(req)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	_ = resp.Body.Close()

	if got != "rid-1" {
		t.Fatalf("got request id header %q", got)
	}
}

func TestRequestIDTransport_DoesNotOverrideExistingHeader(t *testing.T) {
	t.Parallel()

	var got string
	base := rtFunc(func(r *http.Request) (*http.Response, error) {
		got = r.Header.Get(HeaderRequestID)
		return &http.Response{
			StatusCode: 200,
			Body:       io.NopCloser(bytes.NewBufferString("ok")),
			Header:     make(http.Header),
			Request:    r,
		}, nil
	})

	rt := &requestIDTransport{base: base}

	ctx := wsctx.WithRequestID(context.Background(), "rid-ctx")
	req := httptest.NewRequest(http.MethodGet, "http://example.invalid", nil).WithContext(ctx)
	req.Header.Set(HeaderRequestID, "rid-existing")

	resp, err := rt.RoundTrip(req)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	_ = resp.Body.Close()

	if got != "rid-existing" {
		t.Fatalf("expected existing header to win, got %q", got)
	}
}
