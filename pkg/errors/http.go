package errors

import (
	"context"
	"encoding/json"
	"net/http"
)

func Status(code Code) int {
	switch code {
	case CodeInvalidArgument:
		return http.StatusBadRequest
	case CodeUnauthenticated:
		return http.StatusUnauthorized
	case CodeForbidden:
		return http.StatusForbidden
	case CodeNotFound:
		return http.StatusNotFound
	case CodeConflict, CodeAlreadyExists:
		return http.StatusConflict
	case CodeTooManyRequests:
		return http.StatusTooManyRequests
	case CodeResourceExhausted:
		return http.StatusRequestEntityTooLarge
	case CodeDeadlineExceeded:
		return http.StatusGatewayTimeout
	case CodeUnavailable:
		return http.StatusServiceUnavailable
	case CodeFailedPrecondition:
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}

// Write writes a JSON error response. It guarantees:
// - content-type is application/json; charset=utf-8
// - details is always an object (never null)
// - trace_id is injected from OTel span context if missing
func Write(ctx context.Context, w http.ResponseWriter, status int, err *Error) {
	if err == nil {
		err = Internal("internal error")
	}

	// Make output immutable and compliant.
	out := *err
	if out.Details == nil {
		out.Details = map[string]any{}
	} else {
		out.Details = cloneDetails(out.Details)
	}
	if out.TraceID == "" {
		out.TraceID = TraceID(ctx)
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)

	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(true)

	// If encoding fails, we cannot change the status code (already written),
	// but we should attempt to write a minimal valid JSON object.
	if encodeErr := enc.Encode(&out); encodeErr != nil {
		_, _ = w.Write([]byte(`{"code":"INTERNAL","message":"internal error","details":{},"trace_id":""}` + "\n"))
	}
}

func WriteError(ctx context.Context, w http.ResponseWriter, err *Error) {
	if err == nil {
		err = Internal("internal error")
	}
	Write(ctx, w, Status(err.Code), err)
}
