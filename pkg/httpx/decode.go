package httpx

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	wserr "github.com/hanzy-dev/saas-ws-lib/pkg/errors"
)

type testPayload struct {
	Name string `json:"name"`
}

func TestDecodeJSON_Valid(t *testing.T) {
	body := bytes.NewBufferString(`{"name":"ok"}`)
	req := httptest.NewRequest(http.MethodPost, "/", body)

	var p testPayload
	err := DecodeJSON(req, &p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Name != "ok" {
		t.Fatalf("decode mismatch")
	}
}

func TestDecodeJSON_InvalidJSON(t *testing.T) {
	body := bytes.NewBufferString(`{"name":`)
	req := httptest.NewRequest(http.MethodPost, "/", body)

	var p testPayload
	err := DecodeJSON(req, &p)
	if err == nil {
		t.Fatalf("expected error")
	}
	if err.Code != wserr.CodeInvalidArgument {
		t.Fatalf("code mismatch: got=%s", err.Code)
	}
}

func TestDecodeJSON_TooLarge(t *testing.T) {
	body := bytes.NewBufferString(`{"name":"ok"}`)
	req := httptest.NewRequest(http.MethodPost, "/", body)

	rr := httptest.NewRecorder()
	req.Body = http.MaxBytesReader(rr, req.Body, 5)

	var p testPayload
	err := DecodeJSON(req, &p)
	if err == nil {
		t.Fatalf("expected error")
	}
	if err.Code != wserr.CodeResourceExhausted {
		t.Fatalf("expected RESOURCE_EXHAUSTED, got=%s", err.Code)
	}
}
