package testkit

import (
	"encoding/json"
	"io"
	"net/http"
	"testing"

	wserr "saas-ws-lib/pkg/errors"
)

func ReadBody(t *testing.T, r io.Reader) []byte {
	t.Helper()
	b, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	return b
}

func DecodeJSON[T any](t *testing.T, b []byte) T {
	t.Helper()
	var v T
	if err := json.Unmarshal(b, &v); err != nil {
		t.Fatalf("decode json: %v; body=%s", err, string(b))
	}
	return v
}

func AssertHTTPStatus(t *testing.T, got, want int, body []byte) {
	t.Helper()
	if got != want {
		t.Fatalf("status mismatch: got=%d want=%d body=%s", got, want, string(body))
	}
}

func AssertError(t *testing.T, resp *http.Response, wantCode wserr.Code) wserr.Error {
	t.Helper()

	if resp == nil {
		t.Fatalf("nil response")
	}

	body := ReadBody(t, resp.Body)

	if resp.StatusCode != wserr.Status(wantCode) {
		t.Fatalf("status mismatch: got=%d want=%d body=%s", resp.StatusCode, wserr.Status(wantCode), string(body))
	}

	errObj := DecodeJSON[wserr.Error](t, body)

	if errObj.Code != wantCode {
		t.Fatalf("error code mismatch: got=%s want=%s body=%s", errObj.Code, wantCode, string(body))
	}

	if errObj.Message == "" {
		t.Fatalf("error message is empty body=%s", string(body))
	}

	// trace_id can be empty only if service doesn't run OTel + RequestID middleware;
	// in our stack it should exist. Keep strict to enforce discipline.
	if errObj.TraceID == "" {
		t.Fatalf("trace_id is empty body=%s", string(body))
	}

	return errObj
}
