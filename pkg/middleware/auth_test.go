package middleware

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/hanzy-dev/saas-ws-lib/pkg/auth"
	wsctx "github.com/hanzy-dev/saas-ws-lib/pkg/ctx"
)

type fakePolicy struct {
	dec auth.Decision
	err error
	got auth.PolicyRequest
}

func (f *fakePolicy) Check(ctx context.Context, req auth.PolicyRequest) (auth.Decision, error) {
	f.got = req
	return f.dec, f.err
}

func TestAuthMiddleware(t *testing.T) {
	t.Parallel()

	secret := []byte("secret")
	keyFunc := func(t *jwt.Token) (any, error) { return secret, nil }
	verifier := &auth.Verifier{KeyFunc: keyFunc}

	sign := func(sub, tenant string, scopes []string, exp time.Time) string {
		claims := auth.Claims{
			TenantID: tenant,
			Scopes:   scopes,
			RegisteredClaims: jwt.RegisteredClaims{
				Subject:   sub,
				ExpiresAt: jwt.NewNumericDate(exp),
				IssuedAt:  jwt.NewNumericDate(time.Now().Add(-time.Minute)),
			},
		}
		tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		s, err := tok.SignedString(secret)
		if err != nil {
			panic(err)
		}
		return s
	}

	okHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Assert ctx enriched
		if wsctx.SubjectID(r.Context()) == "" || wsctx.TenantID(r.Context()) == "" {
			t.Fatalf("context not enriched")
		}
		w.WriteHeader(204)
	})

	tests := []struct {
		name       string
		cfg        AuthConfig
		authHeader string
		wantStatus int
	}{
		{"missing auth header", AuthConfig{Verifier: verifier}, "", 401},
		{"bad auth header", AuthConfig{Verifier: verifier}, "Token abc", 401},
		{"invalid token", AuthConfig{Verifier: verifier}, "Bearer invalid", 401},
		{"valid token ok", AuthConfig{Verifier: verifier}, "Bearer " + sign("u1", "t1", []string{"a"}, time.Now().Add(time.Hour)), 204},
		{"missing required scope => 403", AuthConfig{Verifier: verifier, RequireScopes: []auth.Scope{"admin"}}, "Bearer " + sign("u1", "t1", []string{"user"}, time.Now().Add(time.Hour)), 403},
		{"has required scopes => 204", AuthConfig{Verifier: verifier, RequireScopes: []auth.Scope{"user"}}, "Bearer " + sign("u1", "t1", []string{"user"}, time.Now().Add(time.Hour)), 204},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			r := httptest.NewRequest(http.MethodGet, "/", nil)
			if tt.authHeader != "" {
				r.Header.Set(HeaderAuthorization, tt.authHeader)
			}
			rr := httptest.NewRecorder()

			h := Auth(tt.cfg)(okHandler)
			h.ServeHTTP(rr, r)

			if rr.Code != tt.wantStatus {
				t.Fatalf("status=%d want=%d body=%s", rr.Code, tt.wantStatus, rr.Body.String())
			}
		})
	}
}

func TestAuthMiddleware_Policy(t *testing.T) {
	t.Parallel()

	secret := []byte("secret")
	keyFunc := func(t *jwt.Token) (any, error) { return secret, nil }
	verifier := &auth.Verifier{KeyFunc: keyFunc}

	claims := auth.Claims{
		TenantID: "t1",
		Scopes:   []string{"s1"},
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "u1",
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-time.Minute)),
		},
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, _ := tok.SignedString(secret)

	okHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(204) })

	t.Run("policy error => 503", func(t *testing.T) {
		p := &fakePolicy{dec: auth.DecisionDeny, err: errors.New("down")}
		cfg := AuthConfig{Verifier: verifier, Policy: p, Action: "a", Resource: "r"}

		r := httptest.NewRequest(http.MethodGet, "/", nil)
		r.Header.Set(HeaderAuthorization, "Bearer "+signed)
		rr := httptest.NewRecorder()

		Auth(cfg)(okHandler).ServeHTTP(rr, r)

		if rr.Code != 503 {
			t.Fatalf("status=%d want=503", rr.Code)
		}
	})

	t.Run("policy deny => 403", func(t *testing.T) {
		p := &fakePolicy{dec: auth.DecisionDeny, err: nil}
		cfg := AuthConfig{Verifier: verifier, Policy: p, Action: "a", Resource: "r"}

		r := httptest.NewRequest(http.MethodGet, "/", nil)
		r.Header.Set(HeaderAuthorization, "Bearer "+signed)
		rr := httptest.NewRecorder()

		Auth(cfg)(okHandler).ServeHTTP(rr, r)

		if rr.Code != 403 {
			t.Fatalf("status=%d want=403", rr.Code)
		}
	})

	t.Run("policy allow => 204 and request filled", func(t *testing.T) {
		p := &fakePolicy{dec: auth.DecisionAllow, err: nil}
		cfg := AuthConfig{Verifier: verifier, Policy: p, Action: "tenant.members.invite", Resource: "tenant:t1"}

		r := httptest.NewRequest(http.MethodGet, "/", nil)
		r.Header.Set(HeaderAuthorization, "Bearer "+signed)
		rr := httptest.NewRecorder()

		Auth(cfg)(okHandler).ServeHTTP(rr, r)

		if rr.Code != 204 {
			t.Fatalf("status=%d want=204", rr.Code)
		}
		if p.got.SubjectID != "u1" || p.got.TenantID != "t1" {
			t.Fatalf("unexpected policy req: %+v", p.got)
		}
	})
}
