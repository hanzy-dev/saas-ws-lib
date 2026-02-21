package middleware

import (
	"errors"
	"net/http"

	wserr "github.com/hanzy-dev/saas-ws-lib/pkg/errors"
)

func BodyLimit(maxBytes int64) func(http.Handler) http.Handler {
	if maxBytes <= 0 {
		panic("middleware.BodyLimit requires positive maxBytes")
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.Body = http.MaxBytesReader(w, r.Body, maxBytes)

			next.ServeHTTP(w, r)

			// If body exceeded limit, Go will return an error on read.
			// We catch it here via the ResponseWriter state.
			if r.Body != nil {
				if err := r.Body.Close(); err != nil {
					var maxErr *http.MaxBytesError
					if errors.As(err, &maxErr) {
						e := wserr.New(
							wserr.CodeInvalidArgument,
							"request body too large",
							map[string]any{"limit_bytes": maxBytes},
						).WithTrace(r.Context())

						wserr.WriteError(w, e)
					}
				}
			}
		})
	}
}
