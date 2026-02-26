package errors

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"testing"
)

func TestStatusMapping(t *testing.T) {
	t.Parallel()

	tests := []struct {
		code Code
		want int
	}{
		{CodeInvalidArgument, 400},
		{CodeUnauthenticated, 401},
		{CodeForbidden, 403},
		{CodeNotFound, 404},
		{CodeConflict, 409},
		{CodeAlreadyExists, 409},
		{CodeTooManyRequests, 429},
		{CodeResourceExhausted, 413},
		{CodeDeadlineExceeded, 504},
		{CodeUnavailable, 503},
		{CodeFailedPrecondition, 400},
		{CodeInternal, 500},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.code.String(), func(t *testing.T) {
			t.Parallel()
			if got := Status(tt.code); got != tt.want {
				t.Fatalf("Status(%s)=%d, want %d", tt.code, got, tt.want)
			}
		})
	}
}

func TestWrite_GuaranteesDetailsObjectAndContentType(t *testing.T) {
	t.Parallel()

	rr := httptest.NewRecorder()

	e := &Error{
		Code:    CodeInvalidArgument,
		Message: "bad",
		Details: nil,
		TraceID: "",
	}

	Write(context.Background(), rr, 400, e)

	if rr.Code != 400 {
		t.Fatalf("status=%d, want 400", rr.Code)
	}

	ct := rr.Header().Get("Content-Type")
	if ct != "application/json; charset=utf-8" {
		t.Fatalf("content-type=%q", ct)
	}

	var out map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &out); err != nil {
		t.Fatalf("invalid json: %v, body=%q", err, rr.Body.String())
	}

	if out["details"] == nil {
		t.Fatalf("details must not be null: %v", out)
	}
	if _, ok := out["details"].(map[string]any); !ok {
		t.Fatalf("details must be object, got %T", out["details"])
	}
	if out["trace_id"] == nil {
		t.Fatalf("trace_id must exist")
	}
}

func TestWrite_ClonesDetails(t *testing.T) {
	t.Parallel()

	rr := httptest.NewRecorder()

	d := map[string]any{"k": "v"}
	e := &Error{
		Code:    CodeInvalidArgument,
		Message: "bad",
		Details: d,
	}

	Write(context.Background(), rr, 400, e)

	// mutate original map after writing; output should not change retroactively
	d["k"] = "mutated"

	var out map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &out); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	gotDetails := out["details"].(map[string]any)
	if gotDetails["k"] != "v" {
		t.Fatalf("expected cloned details, got %v", gotDetails)
	}
}

func TestWriteError_NilErrorDefaultsInternal(t *testing.T) {
	t.Parallel()

	rr := httptest.NewRecorder()
	WriteError(context.Background(), rr, nil)

	if rr.Code != 500 {
		t.Fatalf("status=%d, want 500", rr.Code)
	}
}
