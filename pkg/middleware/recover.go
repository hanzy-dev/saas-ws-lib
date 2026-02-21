package middleware

import (
	"fmt"
	"net/http"
	"runtime/debug"

	wserr "github.com/hanzy-dev/saas-ws-lib/pkg/errors"
	wslog "github.com/hanzy-dev/saas-ws-lib/pkg/log"
)

func Recover(logger *wslog.Logger) func(http.Handler) http.Handler {
	if logger == nil {
		panic("middleware.Recover requires non-nil logger")
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					stack := debug.Stack()

					logger.With(r.Context()).Error(
						"panic recovered",
						"panic", fmt.Sprint(rec),
						"stack", string(stack),
					)

					err := wserr.New(
						wserr.CodeInternal,
						"internal error",
						nil,
					)

					wserr.WriteError(r.Context(), w, err)
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}
