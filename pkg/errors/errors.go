package errors

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	wsctx "github.com/hanzy-dev/saas-ws-lib/pkg/ctx"

	"go.opentelemetry.io/otel/trace"
)

type Error struct {
	Code    Code           `json:"code"`
	Message string         `json:"message"`
	Details map[string]any `json:"details,omitempty"`
	TraceID string         `json:"trace_id,omitempty"`
}

func New(code Code, message string, details map[string]any) *Error {
	return &Error{
		Code:    code,
		Message: message,
		Details: details,
	}
}

func (e *Error) Error() string {
	if e == nil {
		return ""
	}
	if e.Message == "" {
		return e.Code.String()
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// WithTrace attaches trace_id (preferred) or request_id (fallback) from context.
func (e *Error) WithTrace(ctx context.Context) *Error {
	if e == nil {
		return nil
	}

	cp := *e
	cp.TraceID = TraceID(ctx)
	return &cp
}

func TraceID(ctx context.Context) string {
	if sc := trace.SpanContextFromContext(ctx); sc.IsValid() {
		return sc.TraceID().String()
	}
	if rid := wsctx.RequestID(ctx); rid != "" {
		return rid
	}
	return ""
}

// As tries to unwrap any error into *Error.
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

// MarshalJSON ensures Code is rendered as string.
func (e *Error) MarshalJSON() ([]byte, error) {
	type alias Error
	out := struct {
		Code string `json:"code"`
		alias
	}{
		Code:  e.Code.String(),
		alias: alias(*e),
	}
	// alias includes Code too, so we must avoid duplicate by zeroing it.
	out.alias.Code = ""
	return json.Marshal(out)
}
