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
		err = Internal("internal error")
	}

	out := err
	if out.TraceID == "" || out.Details == nil {
		cp := *out
		if cp.Details == nil {
			cp.Details = map[string]any{}
		}
		if cp.TraceID == "" {
			cp.TraceID = TraceID(ctx)
		}
		out = &cp
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(out)
}

func WriteError(ctx context.Context, w http.ResponseWriter, err *Error) {
	if err == nil {
		err = Internal("internal error")
	}
	Write(ctx, w, Status(err.Code), err)
}
