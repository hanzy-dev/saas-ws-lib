package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRequestID_GeneratesIfMissing(t *testing.T) {
	h := RequestID()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Context().Value(interface{}(nil))
		_ = id
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	rid := rr.Header().Get(HeaderRequestID)
	if rid == "" {
		t.Fatalf("expected generated request id")
	}
}

func TestRequestID_RespectsIncomingHeader(t *testing.T) {
	h := RequestID()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(HeaderRequestID, "custom-id")
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	if rr.Header().Get(HeaderRequestID) != "custom-id" {
		t.Fatalf("expected same request id")
	}
}

func TestRequestID_RejectsTooLong(t *testing.T) {
	h := RequestID()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

	longID := strings.Repeat("a", 200)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(HeaderRequestID, longID)

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	rid := rr.Header().Get(HeaderRequestID)
	if rid == longID {
		t.Fatalf("expected regenerated id when too long")
	}
	if rid == "" {
		t.Fatalf("expected non-empty id")
	}
}
