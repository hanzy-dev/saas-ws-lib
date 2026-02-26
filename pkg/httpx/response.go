package httpx

import (
	"encoding/json"
	"net/http"
)

// JSON writes v as JSON with the given status code.
// It sets Content-Type to application/json; charset=utf-8.
// Encoding errors are ignored (best-effort) and must not panic.
func JSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// NoContent writes HTTP 204 with no body.
func NoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}
