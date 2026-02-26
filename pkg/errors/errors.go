package errors

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"go.opentelemetry.io/otel/trace"
)

type Error struct {
	Code    Code           `json:"code"`
	Message string         `json:"message"`
	Details map[string]any `json:"details"`
	TraceID string         `json:"trace_id"`
}

func New(code Code, message string, details map[string]any) *Error {
	return &Error{
		Code:    code,
		Message: message,
		Details: cloneDetails(details),
	}
}

func Internal(message string) *Error           { return New(CodeInternal, message, nil) }
func InvalidArgument(message string) *Error    { return New(CodeInvalidArgument, message, nil) }
func NotFound(message string) *Error           { return New(CodeNotFound, message, nil) }
func Forbidden(message string) *Error          { return New(CodeForbidden, message, nil) }
func Unauthenticated(message string) *Error    { return New(CodeUnauthenticated, message, nil) }
func Conflict(message string) *Error           { return New(CodeConflict, message, nil) }
func ResourceExhausted(message string) *Error  { return New(CodeResourceExhausted, message, nil) }
func Unavailable(message string) *Error        { return New(CodeUnavailable, message, nil) }
func DeadlineExceeded(message string) *Error   { return New(CodeDeadlineExceeded, message, nil) }
func AlreadyExists(message string) *Error      { return New(CodeAlreadyExists, message, nil) }
func FailedPrecondition(message string) *Error { return New(CodeFailedPrecondition, message, nil) }

func (e *Error) Error() string {
	if e == nil {
		return ""
	}
	if e.Message == "" {
		return e.Code.String()
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// WithTrace returns a shallow copy of e with TraceID derived from OTel span context.
// Details map is cloned to keep the error immutable from caller mutations.
func (e *Error) WithTrace(ctx context.Context) *Error {
	if e == nil {
		return nil
	}
	cp := *e
	cp.Details = cloneDetails(e.Details)
	cp.TraceID = TraceID(ctx)
	return &cp
}

// TraceID returns trace_id from OpenTelemetry span context.
// If there's no valid span context, it returns empty string.
// NOTE: request_id is NOT a trace_id and must not be used as fallback.
func TraceID(ctx context.Context) string {
	sc := trace.SpanContextFromContext(ctx)
	if sc.IsValid() {
		return sc.TraceID().String()
	}
	return ""
}

func As(err error) (*Error, bool) {
	if err == nil {
		return nil, false
	}
	var e *Error
	if errors.As(err, &e) && e != nil {
		return e, true
	}
	return nil, false
}

func (e *Error) MarshalJSON() ([]byte, error) {
	if e == nil {
		return []byte("null"), nil
	}

	type out struct {
		Code    string         `json:"code"`
		Message string         `json:"message"`
		Details map[string]any `json:"details"`
		TraceID string         `json:"trace_id"`
	}

	return json.Marshal(out{
		Code:    e.Code.String(),
		Message: e.Message,
		Details: cloneDetails(e.Details),
		TraceID: e.TraceID,
	})
}

func cloneDetails(in map[string]any) map[string]any {
	if len(in) == 0 {
		return map[string]any{}
	}
	out := make(map[string]any, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}
