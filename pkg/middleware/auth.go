package middleware

import (
	"net/http"
	"strings"

	"saas-ws-lib/pkg/auth"
	wsctx "saas-ws-lib/pkg/ctx"
	wserr "saas-ws-lib/pkg/errors"
)

const HeaderAuthorization = "Authorization"

type AuthConfig struct {
	Verifier *auth.Verifier

	// Require scopes for this route (all must be present).
	RequireScopes []auth.Scope

	// Optional policy check (role/permission/abac).
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
				writeAuthErr(w, r, wserr.CodeUnauthenticated, "missing bearer token", nil)
				return
			}

			claims, err := cfg.Verifier.Verify(token)
			if err != nil {
				writeAuthErr(w, r, wserr.CodeUnauthenticated, "invalid token", nil)
				return
			}

			ctx := r.Context()
			ctx = wsctx.WithSubjectID(ctx, claims.Subject)
			ctx = wsctx.WithTenantID(ctx, claims.TenantID)
			ctx = wsctx.WithScopes(ctx, claims.Scopes)

			// Scope enforcement (cheap, local)
			if len(cfg.RequireScopes) > 0 && !auth.HasAll(claims.Scopes, cfg.RequireScopes...) {
				writeAuthErr(w, r.WithContext(ctx), wserr.CodeForbidden, "missing required scope", map[string]any{
					"required": scopesToStrings(cfg.RequireScopes),
				})
				return
			}

			// Optional policy check (expensive, remote)
			if cfg.Policy != nil {
				dec, perr := cfg.Policy.Check(ctx, auth.PolicyRequest{
					SubjectID: claims.Subject,
					TenantID:  claims.TenantID,
					Scopes:    claims.Scopes,
					Action:    cfg.Action,
					Resource:  cfg.Resource,
				})
				if perr != nil {
					writeAuthErr(w, r.WithContext(ctx), wserr.CodeUnavailable, "policy check unavailable", nil)
					return
				}
				if dec != auth.DecisionAllow {
					writeAuthErr(w, r.WithContext(ctx), wserr.CodeForbidden, "policy denied", nil)
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
	tok := strings.TrimSpace(strings.TrimPrefix(v, prefix))
	return tok
}

func scopesToStrings(in []auth.Scope) []string {
	if len(in) == 0 {
		return nil
	}
	out := make([]string, len(in))
	for i := range in {
		out[i] = in[i].String()
	}
	return out
}

func writeAuthErr(w http.ResponseWriter, r *http.Request, code wserr.Code, msg string, details map[string]any) {
	err := wserr.New(code, msg, details).WithTrace(r.Context())
	wserr.WriteError(w, err)
}
