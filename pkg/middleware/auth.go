package middleware

import (
	"net/http"

	"github.com/hanzy-dev/saas-ws-lib/pkg/auth"
	wsctx "github.com/hanzy-dev/saas-ws-lib/pkg/ctx"
	wserr "github.com/hanzy-dev/saas-ws-lib/pkg/errors"
)

const HeaderAuthorization = "Authorization"

type AuthConfig struct {
	Verifier *auth.Verifier

	// RequireScopes enforces that the authenticated principal has all listed scopes.
	RequireScopes []auth.Scope

	// Policy optionally performs an external policy check.
	// If Policy is set, Action and Resource must be non-empty stable strings.
	Policy   auth.PolicyChecker
	Action   string
	Resource string
}

// Auth authenticates requests using JWT in Authorization header and enriches context with:
// subject_id (sub), tenant_id, and scopes.
//
// It never leaks token verification details to clients. All failures map to standardized errors.
func Auth(cfg AuthConfig) func(http.Handler) http.Handler {
	if cfg.Verifier == nil {
		panic("middleware.Auth requires non-nil Verifier")
	}
	if cfg.Policy != nil && (cfg.Action == "" || cfg.Resource == "") {
		panic("middleware.Auth requires non-empty Action and Resource when Policy is set")
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			raw := r.Header.Get(HeaderAuthorization)
			token, err := auth.ParseBearer(raw)
			if err != nil {
				wserr.WriteError(r.Context(), w, wserr.Unauthenticated("authentication required"))
				return
			}

			claims, verr := cfg.Verifier.Verify(token)
			if verr != nil {
				wserr.WriteError(r.Context(), w, wserr.Unauthenticated("authentication required"))
				return
			}

			ctx := r.Context()
			ctx = wsctx.WithSubjectID(ctx, claims.Subject)
			ctx = wsctx.WithTenantID(ctx, claims.TenantID)
			ctx = wsctx.WithScopes(ctx, claims.Scopes)

			if len(cfg.RequireScopes) > 0 && !auth.HasAll(claims.Scopes, cfg.RequireScopes...) {
				wserr.WriteError(ctx, w, wserr.Forbidden("forbidden"))
				return
			}

			if cfg.Policy != nil {
				dec, perr := cfg.Policy.Check(ctx, auth.PolicyRequest{
					SubjectID: claims.Subject,
					TenantID:  claims.TenantID,
					Scopes:    claims.Scopes,
					Action:    cfg.Action,
					Resource:  cfg.Resource,
				})
				if perr != nil {
					wserr.WriteError(ctx, w, wserr.Unavailable("authorization service unavailable"))
					return
				}
				if !dec.IsAllow() {
					wserr.WriteError(ctx, w, wserr.Forbidden("forbidden"))
					return
				}
			}

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
