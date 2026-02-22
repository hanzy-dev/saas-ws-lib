package errors

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	wsctx "github.com/hanzy-dev/saas-ws-lib/pkg/ctx"
)

func TestNew_EnforcesNonNilDetails(t *testing.T) {
	e := New(CodeInvalidArgument, "validation failed", nil)
	if e == nil {
		t.Fatalf("expected error")
	}
	if e.Details == nil {
		t.Fatalf("expected non-nil details")
	}
}

func TestWithTrace_UsesRequestIDFallback(t *testing.T) {
	ctx := wsctx.WithRequestID(context.Background(), "rid-123")

	e := New(CodeInternal, "internal error", nil).WithTrace(ctx)
	if e.TraceID != "rid-123" {
		t.Fatalf("trace_id mismatch: got=%q want=%q", e.TraceID, "rid-123")
	}
	if e.Details == nil {
		t.Fatalf("expected non-nil details")
	}
}

func TestMarshalJSON_StableSchema(t *testing.T) {
	e := New(CodeNotFound, "not found", nil)
	b, err := json.Marshal(e)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if m["code"] != "NOT_FOUND" {
		t.Fatalf("code mismatch: got=%v", m["code"])
	}

	if _, ok := m["message"]; !ok {
		t.Fatalf("missing message field")
	}
	if _, ok := m["details"]; !ok {
		t.Fatalf("missing details field")
	}
	if _, ok := m["trace_id"]; !ok {
		t.Fatalf("missing trace_id field")
	}
}

func TestAs_Unwraps(t *testing.T) {
	base := New(CodeConflict, "conflict", nil)
	wrapped := errors.New("wrap: " + base.Error())

	if _, ok := As(wrapped); ok {
		t.Fatalf("expected false for non-wrapped *Error")
	}

	err := errors.Join(errors.New("x"), base)

	got, ok := As(err)
	if !ok {
		t.Fatalf("expected unwrap ok")
	}
	if got.Code != CodeConflict {
		t.Fatalf("code mismatch: got=%s want=%s", got.Code, CodeConflict)
	}
}
