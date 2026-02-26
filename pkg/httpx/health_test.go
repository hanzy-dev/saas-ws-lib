package httpx

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewHealth_DefaultTimeout(t *testing.T) {
	t.Parallel()

	h := NewHealth(0)
	if h.Timeout != 2*time.Second {
		t.Fatalf("timeout=%s", h.Timeout)
	}
}

func TestHealthz(t *testing.T) {
	t.Parallel()

	h := NewHealth(100 * time.Millisecond)
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)

	h.Healthz(rr, req)

	if rr.Code != 200 {
		t.Fatalf("status=%d", rr.Code)
	}

	var out map[string]string
	if err := json.Unmarshal(rr.Body.Bytes(), &out); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if out["status"] != "ok" {
		t.Fatalf("out=%v", out)
	}
}

func TestReadyz_AllChecksPass(t *testing.T) {
	t.Parallel()

	var sawDeadline bool
	check := func(ctx context.Context) error {
		_, ok := ctx.Deadline()
		sawDeadline = ok
		return nil
	}

	h := NewHealth(50*time.Millisecond, check)
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)

	h.Readyz(rr, req)

	if rr.Code != 200 {
		t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
	}
	if !sawDeadline {
		t.Fatalf("expected ctx deadline to be set")
	}

	var out map[string]string
	_ = json.Unmarshal(rr.Body.Bytes(), &out)
	if out["status"] != "ready" {
		t.Fatalf("out=%v", out)
	}
}

func TestReadyz_CheckFails(t *testing.T) {
	t.Parallel()

	h := NewHealth(50*time.Millisecond, func(ctx context.Context) error {
		return errors.New("down")
	})

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)

	h.Readyz(rr, req)

	if rr.Code != 503 {
		t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
	}

	var out map[string]string
	_ = json.Unmarshal(rr.Body.Bytes(), &out)
	if out["status"] != "not_ready" {
		t.Fatalf("out=%v", out)
	}
}
