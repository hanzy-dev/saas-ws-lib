package errors

import (
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
	case CodeConflict:
		return http.StatusConflict
	case CodeTooManyRequests:
		return http.StatusTooManyRequests
	case CodeUnavailable:
		return http.StatusServiceUnavailable
	default:
		return http.StatusInternalServerError
	}
}

func Write(w http.ResponseWriter, status int, err *Error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(err)
}

// WriteError maps error code to HTTP status automatically.
// If err is nil, it writes INTERNAL.
func WriteError(w http.ResponseWriter, err *Error) {
	if err == nil {
		Write(w, http.StatusInternalServerError, New(CodeInternal, "internal error", nil))
		return
	}
	Write(w, Status(err.Code), err)
}
