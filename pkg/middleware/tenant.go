package middleware

import (
	"net/http"
	"strings"

	wsctx "github.com/hanzy-dev/saas-ws-lib/pkg/ctx"
	wserr "github.com/hanzy-dev/saas-ws-lib/pkg/errors"
)

const HeaderTenantID = "X-Tenant-ID"

type TenantMode int

const (
	// TenantFromContext uses tenant_id already present in request context (e.g. from JWT claims).
	TenantFromContext TenantMode = iota
	// TenantAllowHeader allows tenant_id override from a header (intended for internal calls only).
	TenantAllowHeader
)

type TenantConfig struct {
	Mode     TenantMode
	Required bool
	Header   string

	// AllowHeaderWithoutAuth allows reading tenant_id from header even when subject_id is empty.
	// Default false. Keep this disabled for public edge handlers.
	AllowHeaderWithoutAuth bool
}

func Tenant(cfg TenantConfig) func(http.Handler) http.Handler {
	if cfg.Header == "" {
		cfg.Header = HeaderTenantID
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			// tenant from context (preferred)
			tid := wsctx.TenantID(ctx)

			// optional override from header (internal-only by default)
			if tid == "" && cfg.Mode == TenantAllowHeader {
				if cfg.AllowHeaderWithoutAuth || wsctx.SubjectID(ctx) != "" {
					if h := strings.TrimSpace(r.Header.Get(cfg.Header)); h != "" {
						ctx = wsctx.WithTenantID(ctx, h)
					}
				}
			}

			if cfg.Required && wsctx.TenantID(ctx) == "" {
				if wsctx.SubjectID(ctx) == "" {
					wserr.WriteError(ctx, w, wserr.Unauthenticated("missing authentication"))
					return
				}
				wserr.WriteError(ctx, w, wserr.Forbidden("missing tenant_id"))
				return
			}

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
