package httpx

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestJSON_WritesHeaderStatusAndBody(t *testing.T) {
	t.Parallel()

	rr := httptest.NewRecorder()
	JSON(rr, 201, map[string]any{"ok": true})

	if rr.Code != 201 {
		t.Fatalf("status=%d", rr.Code)
	}
	if ct := rr.Header().Get("Content-Type"); ct != "application/json; charset=utf-8" {
		t.Fatalf("content-type=%q", ct)
	}

	var out map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &out); err != nil {
		t.Fatalf("invalid json: %v body=%q", err, rr.Body.String())
	}
	if out["ok"] != true {
		t.Fatalf("out=%v", out)
	}
}

func TestNoContent(t *testing.T) {
	t.Parallel()

	rr := httptest.NewRecorder()
	NoContent(rr)

	if rr.Code != http.StatusNoContent {
		t.Fatalf("status=%d", rr.Code)
	}
	if rr.Body.Len() != 0 {
		t.Fatalf("expected empty body, got %q", rr.Body.String())
	}
}
