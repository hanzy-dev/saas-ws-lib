package middleware

import (
	"net/http"

	wsctx "github.com/hanzy-dev/saas-ws-lib/pkg/ctx"

	"github.com/google/uuid"
)

const HeaderRequestID = "X-Request-ID"

func RequestID() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rid := r.Header.Get(HeaderRequestID)
			if rid == "" {
				rid = uuid.NewString()
			}

			w.Header().Set(HeaderRequestID, rid)

			ctx := wsctx.WithRequestID(r.Context(), rid)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
