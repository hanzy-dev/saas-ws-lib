package httpx

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRequireJSON_PassesForGET(t *testing.T) {
	t.Parallel()

	h := RequireJSON(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(204)
	}))

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	h.ServeHTTP(rr, req)

	if rr.Code != 204 {
		t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
	}
}

func TestRequireJSON_RejectsNonJSON(t *testing.T) {
	t.Parallel()

	h := RequireJSON(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(204)
	}))

	tests := []struct {
		name string
		ct   string
	}{
		{"empty", ""},
		{"text", "text/plain"},
		{"xml", "application/xml"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			rr := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/", nil)
			if tt.ct != "" {
				req.Header.Set("Content-Type", tt.ct)
			}
			h.ServeHTTP(rr, req)

			if rr.Code != 400 {
				t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
			}

			var out map[string]any
			if err := json.Unmarshal(rr.Body.Bytes(), &out); err != nil {
				t.Fatalf("invalid json: %v body=%q", err, rr.Body.String())
			}
			if out["code"] != "INVALID_ARGUMENT" {
				t.Fatalf("code=%v", out["code"])
			}
		})
	}
}

func TestRequireJSON_AllowsJSONWithCharset(t *testing.T) {
	t.Parallel()

	h := RequireJSON(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(204)
	}))

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPatch, "/", nil)
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	h.ServeHTTP(rr, req)

	if rr.Code != 204 {
		t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
	}
}
