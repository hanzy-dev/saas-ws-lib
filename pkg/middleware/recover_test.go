package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	wslog "github.com/hanzy-dev/saas-ws-lib/pkg/log"
)

func TestRecover_RecoversPanic(t *testing.T) {
	logger := wslog.NewJSON(wslog.Options{})

	h := Recover(logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("boom")
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rr.Code)
	}

	if rr.Body.Len() == 0 {
		t.Fatalf("expected error body")
	}
}
