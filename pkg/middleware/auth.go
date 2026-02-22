package middleware

import (
	"net/http"
	"strings"

	"github.com/hanzy-dev/saas-ws-lib/pkg/auth"
	wsctx "github.com/hanzy-dev/saas-ws-lib/pkg/ctx"
	wserr "github.com/hanzy-dev/saas-ws-lib/pkg/errors"
)

const HeaderAuthorization = "Authorization"

type AuthConfig struct {
	Verifier *auth.Verifier

	RequireScopes []auth.Scope

	Policy   auth.PolicyChecker
	Action   string
	Resource string
}

func Auth(cfg AuthConfig) func(http.Handler) http.Handler {
	if cfg.Verifier == nil {
		panic("middleware.Auth requires non-nil Verifier")
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := bearerToken(r.Header.Get(HeaderAuthorization))
			if token == "" {
				wserr.WriteError(r.Context(), w, wserr.Unauthenticated("authentication required"))
				return
			}

			claims, err := cfg.Verifier.Verify(token)
			if err != nil {
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
				if dec != auth.DecisionAllow {
					wserr.WriteError(ctx, w, wserr.Forbidden("forbidden"))
					return
				}
			}

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func bearerToken(v string) string {
	if v == "" {
		return ""
	}
	const prefix = "Bearer "
	if !strings.HasPrefix(v, prefix) {
		return ""
	}
	return strings.TrimSpace(strings.TrimPrefix(v, prefix))
}
