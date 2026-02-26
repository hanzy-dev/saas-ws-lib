package middleware

import "net/http"

// BodyLimit limits the size of request body read by downstream handlers.
// It wraps r.Body with http.MaxBytesReader.
//
// Panics if maxBytes <= 0.
func BodyLimit(maxBytes int64) func(http.Handler) http.Handler {
	if maxBytes <= 0 {
		panic("middleware.BodyLimit requires positive maxBytes")
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
			next.ServeHTTP(w, r)
		})
	}
}
