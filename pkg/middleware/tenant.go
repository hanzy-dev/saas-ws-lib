package middleware

import (
	"net/http"

	wsctx "saas-ws-lib/pkg/ctx"
	wserr "saas-ws-lib/pkg/errors"
)

const HeaderTenantID = "X-Tenant-ID"

type TenantMode int

const (
	// Tenant is expected to be present in context (typically set by Auth middleware from JWT claims).
	TenantFromContext TenantMode = iota

	// Tenant can be read from X-Tenant-ID header (only for trusted/internal traffic).
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
					tid = h
					ctx = wsctx.WithTenantID(ctx, tid)
				}
			}

			if cfg.Required && wsctx.TenantID(ctx) == "" {
				err := wserr.New(
					wserr.CodeInvalidArgument,
					"missing tenant_id",
					nil,
				).WithTrace(ctx)

				wserr.WriteError(w, err)
				return
			}

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
