package errors

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"go.opentelemetry.io/otel/trace"
)

func TestConstructors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		fn   func(string) *Error
		code Code
	}{
		{"Internal", Internal, CodeInternal},
		{"InvalidArgument", InvalidArgument, CodeInvalidArgument},
		{"NotFound", NotFound, CodeNotFound},
		{"Forbidden", Forbidden, CodeForbidden},
		{"Unauthenticated", Unauthenticated, CodeUnauthenticated},
		{"Conflict", Conflict, CodeConflict},
		{"ResourceExhausted", ResourceExhausted, CodeResourceExhausted},
		{"Unavailable", Unavailable, CodeUnavailable},
		{"DeadlineExceeded", DeadlineExceeded, CodeDeadlineExceeded},
		{"AlreadyExists", AlreadyExists, CodeAlreadyExists},
		{"FailedPrecondition", FailedPrecondition, CodeFailedPrecondition},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			e := tt.fn("m")
			if e == nil {
				t.Fatal("expected non-nil error")
			}
			if e.Code != tt.code {
				t.Fatalf("code=%v want=%v", e.Code, tt.code)
			}
			if e.Message != "m" {
				t.Fatalf("message=%q want=%q", e.Message, "m")
			}
			if e.Details == nil {
				t.Fatalf("details must not be nil")
			}
		})
	}
}

func TestAs_NoMatch(t *testing.T) {
	t.Parallel()

	_, ok := As(errors.New("plain"))
	if ok {
		t.Fatalf("expected ok=false")
	}
}

func TestTraceID_FromSpanContext(t *testing.T) {
	t.Parallel()

	// Create a valid span context and inject into context
	tid := trace.TraceID{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}
	sid := trace.SpanID{9, 9, 9, 9, 9, 9, 9, 9}

	sc := trace.NewSpanContext(trace.SpanContextConfig{
		TraceID:    tid,
		SpanID:     sid,
		TraceFlags: trace.FlagsSampled,
		Remote:     true,
	})
	ctx := trace.ContextWithSpanContext(context.Background(), sc)

	got := TraceID(ctx)
	if got == "" {
		t.Fatalf("expected non-empty trace_id")
	}
	if got != tid.String() {
		t.Fatalf("trace_id=%q want=%q", got, tid.String())
	}
}

func TestStatus_DefaultInternal(t *testing.T) {
	t.Parallel()

	// Use a code that should hit default branch (unknown)
	if got := Status(Code("SOME_UNKNOWN_CODE")); got != http.StatusInternalServerError {
		t.Fatalf("status=%d want=%d", got, http.StatusInternalServerError)
	}
}

func TestWrite_NilErrDefaultsInternal(t *testing.T) {
	t.Parallel()

	rr := httptest.NewRecorder()
	Write(context.Background(), rr, 500, nil)

	if rr.Code != 500 {
		t.Fatalf("status=%d want=500", rr.Code)
	}

	var out map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &out); err != nil {
		t.Fatalf("invalid json: %v body=%q", err, rr.Body.String())
	}
	if out["code"] == nil || out["message"] == nil || out["details"] == nil || out["trace_id"] == nil {
		t.Fatalf("missing fields: %v", out)
	}
}
