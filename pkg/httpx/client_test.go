package httpx

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	wsctx "github.com/hanzy-dev/saas-ws-lib/pkg/ctx"
)

func TestDo_RetryOnTimeout_Idempotent(t *testing.T) {
	var calls int32

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if atomic.AddInt32(&calls, 1) == 1 {
			time.Sleep(50 * time.Millisecond)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	client := NewClient(ClientConfig{
		Timeout:    10 * time.Millisecond,
		MaxRetries: 1,
	})

	req, _ := http.NewRequest(http.MethodGet, srv.URL, nil)

	resp, err := Do(context.Background(), client, req, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200")
	}
	if atomic.LoadInt32(&calls) < 2 {
		t.Fatalf("expected retry")
	}
}

func TestDo_NoRetryOnPost(t *testing.T) {
	var calls int32

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		time.Sleep(50 * time.Millisecond)
	}))
	defer srv.Close()

	client := NewClient(ClientConfig{
		Timeout:    10 * time.Millisecond,
		MaxRetries: 3,
	})

	req, _ := http.NewRequest(http.MethodPost, srv.URL, nil)

	_, _ = Do(context.Background(), client, req, 3)

	if atomic.LoadInt32(&calls) != 1 {
		t.Fatalf("post should not retry")
	}
}

func TestRequestIDPropagation(t *testing.T) {
	var got string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = r.Header.Get(HeaderRequestID)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	client := NewClient(ClientConfig{})

	ctx := wsctx.WithRequestID(context.Background(), "rid-xyz")
	req, _ := http.NewRequest(http.MethodGet, srv.URL, nil)

	resp, err := Do(ctx, client, req, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	resp.Body.Close()

	if got != "rid-xyz" {
		t.Fatalf("request id not propagated")
	}
}
