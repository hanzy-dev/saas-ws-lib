package middleware

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestBodyLimit_PanicsOnNonPositive(t *testing.T) {
	t.Parallel()

	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("expected panic")
		}
	}()

	_ = BodyLimit(0)
}

func TestBodyLimit_EnforcesLimit(t *testing.T) {
	t.Parallel()

	// Handler reads full body; should error when exceeding limit
	h := BodyLimit(5)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := io.ReadAll(r.Body)
		if err == nil {
			t.Fatalf("expected read error due to body limit")
		}
		// Error string from MaxBytesReader is stable enough for a contains check.
		if !strings.Contains(err.Error(), "request body too large") {
			t.Fatalf("unexpected error: %v", err)
		}
		w.WriteHeader(413)
	}))

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("0123456789"))
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	if rr.Code != 413 {
		t.Fatalf("status=%d want=413", rr.Code)
	}
}
