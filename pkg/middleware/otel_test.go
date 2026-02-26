package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestOTel_UsesSpanNameFormatter(t *testing.T) {
	t.Parallel()

	called := false
	mw := OTel(OTelConfig{
		SpanNameFormatter: func(operation string, r *http.Request) string {
			called = true
			return "custom"
		},
	})

	h := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(204)
	}))

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	h.ServeHTTP(rr, req)

	if rr.Code != 204 {
		t.Fatalf("status=%d", rr.Code)
	}
	if !called {
		t.Fatalf("expected span name formatter to be called")
	}
}
