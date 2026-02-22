package errors

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	wsctx "github.com/hanzy-dev/saas-ws-lib/pkg/ctx"
)

func TestWriteError_EnforcesTraceIDAndDetails(t *testing.T) {
	ctx := wsctx.WithRequestID(context.Background(), "rid-abc")

	rr := httptest.NewRecorder()
	WriteError(ctx, rr, New(CodeInvalidArgument, "bad", nil))

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status mismatch: got=%d want=%d", rr.Code, http.StatusBadRequest)
	}

	var e Error
	if err := json.Unmarshal(rr.Body.Bytes(), &e); err != nil {
		t.Fatalf("decode: %v", err)
	}

	if e.TraceID != "rid-abc" {
		t.Fatalf("trace_id mismatch: got=%q want=%q", e.TraceID, "rid-abc")
	}
	if e.Details == nil {
		t.Fatalf("details must be non-nil")
	}
	if e.Code != CodeInvalidArgument {
		t.Fatalf("code mismatch: got=%s want=%s", e.Code, CodeInvalidArgument)
	}
}

func TestStatus_ResourceExhausted_Is413(t *testing.T) {
	if Status(CodeResourceExhausted) != http.StatusRequestEntityTooLarge {
		t.Fatalf("expected 413")
	}
}
