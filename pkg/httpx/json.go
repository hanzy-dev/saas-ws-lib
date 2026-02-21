package httpx

import (
	"net/http"
	"strings"

	wserr "github.com/hanzy-dev/saas-ws-lib/pkg/errors"
)

func RequireJSON(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost, http.MethodPut, http.MethodPatch:
			ct := r.Header.Get("Content-Type")
			// allow "application/json; charset=utf-8"
			if ct == "" || !strings.HasPrefix(strings.ToLower(ct), "application/json") {
				err := wserr.New(
					wserr.CodeInvalidArgument,
					"content-type must be application/json",
					map[string]any{"content_type": ct},
				).WithTrace(r.Context())

				wserr.WriteError(w, err)
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}
