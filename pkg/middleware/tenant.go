package middleware

import (
	"net/http"

	wsctx "github.com/hanzy-dev/saas-ws-lib/pkg/ctx"
	wserr "github.com/hanzy-dev/saas-ws-lib/pkg/errors"
)

const HeaderTenantID = "X-Tenant-ID"

type TenantMode int

const (
	TenantFromContext TenantMode = iota
	TenantAllowHeader
)

type TenantConfig struct {
	Mode     TenantMode
	Required bool
	Header   string
}

func Tenant(cfg TenantConfig) func(http.Handler) http.Handler {
	if cfg.Header == "" {
		cfg.Header = HeaderTenantID
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			tid := wsctx.TenantID(ctx)

			if tid == "" && cfg.Mode == TenantAllowHeader {
				if h := r.Header.Get(cfg.Header); h != "" {
					ctx = wsctx.WithTenantID(ctx, h)
				}
			}

			if cfg.Required && wsctx.TenantID(ctx) == "" {
				if wsctx.SubjectID(ctx) == "" {
					wserr.WriteError(ctx, w, wserr.New(wserr.CodeUnauthenticated, "missing authentication", nil))
					return
				}
				wserr.WriteError(ctx, w, wserr.New(wserr.CodeForbidden, "missing tenant_id", nil))
				return
			}

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
