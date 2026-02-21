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
	Details map[string]any `json:"details"`
	TraceID string         `json:"trace_id"`
}

func New(code Code, message string, details map[string]any) *Error {
	if details == nil {
		details = map[string]any{}
	}
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

func (e *Error) WithTrace(ctx context.Context) *Error {
	if e == nil {
		return nil
	}

	cp := *e
	cp.TraceID = TraceID(ctx)
	if cp.Details == nil {
		cp.Details = map[string]any{}
	}
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
	if e.Details == nil {
		e = &Error{
			Code:    e.Code,
			Message: e.Message,
			Details: map[string]any{},
			TraceID: e.TraceID,
		}
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
		Details: e.Details,
		TraceID: e.TraceID,
	})
}
