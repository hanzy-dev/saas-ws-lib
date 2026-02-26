package httpx

import (
	"net/http"
	"strings"

	wserr "github.com/hanzy-dev/saas-ws-lib/pkg/errors"
)

// RequireJSON enforces Content-Type: application/json for methods that typically carry a body
// (POST, PUT, PATCH). Other methods pass through.
func RequireJSON(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost, http.MethodPut, http.MethodPatch:
			ct := r.Header.Get("Content-Type")
			if ct == "" || !strings.HasPrefix(strings.ToLower(ct), "application/json") {
				err := wserr.New(
					wserr.CodeInvalidArgument,
					"content-type must be application/json",
					map[string]any{"content_type": ct},
				)
				wserr.WriteError(r.Context(), w, err)
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}
