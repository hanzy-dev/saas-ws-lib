package errors

import (
	"context"
	stdErrors "errors"
	"testing"
)

func TestNew_EnsuresDetailsObject(t *testing.T) {
	t.Parallel()

	e := New(CodeInternal, "x", nil)
	if e == nil {
		t.Fatal("expected non-nil")
	}
	if e.Details == nil {
		t.Fatal("Details must not be nil")
	}
	if len(e.Details) != 0 {
		t.Fatalf("expected empty details, got %v", e.Details)
	}
}

func TestNew_ClonesDetails(t *testing.T) {
	t.Parallel()

	in := map[string]any{"a": 1}
	e := New(CodeInternal, "x", in)

	in["a"] = 2
	if e.Details["a"].(int) != 1 {
		t.Fatalf("expected cloned details, got %v", e.Details)
	}
}

func TestErrorString(t *testing.T) {
	t.Parallel()

	var e *Error
	if e.Error() != "" {
		t.Fatalf("nil error string should be empty")
	}

	e = &Error{Code: CodeInvalidArgument, Message: ""}
	if e.Error() != CodeInvalidArgument.String() {
		t.Fatalf("got %q", e.Error())
	}

	e = &Error{Code: CodeInvalidArgument, Message: "bad"}
	if e.Error() != "INVALID_ARGUMENT: bad" {
		t.Fatalf("got %q", e.Error())
	}
}

func TestWithTrace_ClonesDetails_AndSetsTraceIDEmptyWithoutSpan(t *testing.T) {
	t.Parallel()

	base := New(CodeInternal, "x", map[string]any{"k": "v"})
	out := base.WithTrace(context.Background())

	if out == nil {
		t.Fatal("expected non-nil")
	}
	if out == base {
		t.Fatal("expected copy")
	}
	if out.TraceID != "" {
		t.Fatalf("expected empty trace_id without span, got %q", out.TraceID)
	}

	// mutate base details after WithTrace; out must not change
	base.Details["k"] = "mutated"
	if out.Details["k"] != "v" {
		t.Fatalf("expected cloned details, got %v", out.Details)
	}
}

func TestAs(t *testing.T) {
	t.Parallel()

	if _, ok := As(nil); ok {
		t.Fatal("expected ok=false")
	}

	e := Internal("x")
	wrapped := stdErrors.New("wrap: " + e.Error())
	_ = wrapped // ensure no false positive

	w2 := stdErrors.Join(stdErrors.New("x"), e)
	got, ok := As(w2)
	if !ok || got == nil {
		t.Fatal("expected ok=true")
	}
	if got.Code != CodeInternal {
		t.Fatalf("expected CodeInternal, got %v", got.Code)
	}
}

func TestMarshalJSON_NilAndEmptyDetails(t *testing.T) {
	t.Parallel()

	var e *Error
	b, err := e.MarshalJSON()
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if string(b) != "null" {
		t.Fatalf("expected null, got %s", string(b))
	}

	e2 := &Error{
		Code:    CodeInternal,
		Message: "x",
		Details: nil,
		TraceID: "",
	}
	b2, err := e2.MarshalJSON()
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	// Must include "details":{} not null.
	if string(b2) == "" || !contains(string(b2), `"details":{}`) {
		t.Fatalf("expected details object, got %s", string(b2))
	}
}

func contains(s, sub string) bool {
	return len(sub) == 0 || (len(s) >= len(sub) && indexOf(s, sub) >= 0)
}

func indexOf(s, sub string) int {
	// tiny helper to avoid importing strings (keep test lean)
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
