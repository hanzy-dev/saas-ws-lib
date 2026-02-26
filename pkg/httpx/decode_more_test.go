package httpx

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	wserr "github.com/hanzy-dev/saas-ws-lib/pkg/errors"
)

type payloadStrict struct {
	Name string `json:"name"`
}

func TestDecodeJSON_NilRequestOrDst(t *testing.T) {
	t.Parallel()

	var p payloadStrict
	if err := DecodeJSON(nil, &p); err == nil || err.Code != wserr.CodeInvalidArgument {
		t.Fatalf("expected invalid argument for nil request")
	}
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(`{"name":"x"}`))
	if err := DecodeJSON(req, nil); err == nil || err.Code != wserr.CodeInvalidArgument {
		t.Fatalf("expected invalid argument for nil dst")
	}
}

func TestDecodeJSON_EmptyBody(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Body = http.NoBody

	var p payloadStrict
	err := DecodeJSON(req, &p)
	if err == nil || err.Code != wserr.CodeInvalidArgument {
		t.Fatalf("expected INVALID_ARGUMENT for empty body, got %+v", err)
	}
}

func TestDecodeJSON_UnknownFieldsRejected(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(`{"name":"x","extra":1}`))

	var p payloadStrict
	err := DecodeJSON(req, &p)
	if err == nil || err.Code != wserr.CodeInvalidArgument {
		t.Fatalf("expected INVALID_ARGUMENT for unknown field, got %+v", err)
	}
}

func TestDecodeJSON_TrailingGarbageRejected(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(`{"name":"x"} {"name":"y"}`))

	var p payloadStrict
	err := DecodeJSON(req, &p)
	if err == nil || err.Code != wserr.CodeInvalidArgument {
		t.Fatalf("expected INVALID_ARGUMENT for trailing tokens, got %+v", err)
	}
}

func TestDecodeJSON_ExtraBytesAfterObjectRejected(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(`{"name":"x"} 123`))

	var p payloadStrict
	err := DecodeJSON(req, &p)
	if err == nil || err.Code != wserr.CodeInvalidArgument {
		t.Fatalf("expected INVALID_ARGUMENT for extra bytes, got %+v", err)
	}
}
