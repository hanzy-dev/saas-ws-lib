package middleware

import (
	"context"
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
				writeAuthErr(r.Context(), w, wserr.CodeUnauthenticated, "missing bearer token", nil)
				return
			}

			claims, err := cfg.Verifier.Verify(token)
			if err != nil {
				writeAuthErr(r.Context(), w, wserr.CodeUnauthenticated, "invalid token", nil)
				return
			}

			ctx := r.Context()
			ctx = wsctx.WithSubjectID(ctx, claims.Subject)
			ctx = wsctx.WithTenantID(ctx, claims.TenantID)
			ctx = wsctx.WithScopes(ctx, claims.Scopes)

			if len(cfg.RequireScopes) > 0 && !auth.HasAll(claims.Scopes, cfg.RequireScopes...) {
				writeAuthErr(ctx, w, wserr.CodeForbidden, "missing required scope", map[string]any{
					"required": scopesToStrings(cfg.RequireScopes),
				})
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
					writeAuthErr(ctx, w, wserr.CodeUnavailable, "policy check unavailable", nil)
					return
				}
				if dec != auth.DecisionAllow {
					writeAuthErr(ctx, w, wserr.CodeForbidden, "policy denied", nil)
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

func writeAuthErr(ctx context.Context, w http.ResponseWriter, code wserr.Code, msg string, details map[string]any) {
	err := wserr.New(code, msg, details)
	wserr.WriteError(ctx, w, err)
}
