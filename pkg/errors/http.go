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

func Write(ctx context.Context, w http.ResponseWriter, status int, err *Error) {
	if err == nil {
		err = New(CodeInternal, "internal error", nil)
	}

	if err.Details == nil {
		err.Details = map[string]any{}
	}

	if err.TraceID == "" {
		err.TraceID = TraceID(ctx)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(err)
}

func WriteError(ctx context.Context, w http.ResponseWriter, err *Error) {
	if err == nil {
		err = New(CodeInternal, "internal error", nil)
	}
	Write(ctx, w, Status(err.Code), err)
}
