package auth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestParseBearer(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   string
		want string
		ok   bool
	}{
		{"empty", "", "", false},
		{"no bearer", "Token abc", "", false},
		{"one part", "Bearer", "", false},
		{"two parts ok", "Bearer abc.def.ghi", "abc.def.ghi", true},
		{"case insensitive", "bearer t", "t", true},
		{"extra spaces", "  Bearer   t  ", "t", true},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := ParseBearer(tt.in)
			if tt.ok && err != nil {
				t.Fatalf("expected ok, got err=%v", err)
			}
			if !tt.ok && err == nil {
				t.Fatalf("expected err, got nil")
			}
			if got != tt.want {
				t.Fatalf("got=%q want=%q", got, tt.want)
			}
		})
	}
}

func TestVerifier_Verify_HS256(t *testing.T) {
	t.Parallel()

	secret := []byte("secret")

	keyFunc := func(t *jwt.Token) (any, error) { return secret, nil }

	makeToken := func(sub, tenant, iss string, aud []string, exp time.Time) string {
		claims := Claims{
			TenantID: tenant,
			Scopes:   []string{"a", "b"},
			RegisteredClaims: jwt.RegisteredClaims{
				Subject:   sub,
				Issuer:    iss,
				Audience:  aud,
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

	v := Verifier{KeyFunc: keyFunc}

	t.Run("valid", func(t *testing.T) {
		t.Parallel()
		token := makeToken("u1", "t1", "", nil, time.Now().Add(time.Hour))
		got, err := v.Verify(token)
		if err != nil {
			t.Fatalf("unexpected err: %v", err)
		}
		if got.Subject != "u1" || got.TenantID != "t1" {
			t.Fatalf("unexpected claims: %+v", got)
		}
	})

	t.Run("missing keyfunc", func(t *testing.T) {
		t.Parallel()
		_, err := (Verifier{}).Verify("x")
		if err != ErrMissingKeyFunc {
			t.Fatalf("expected ErrMissingKeyFunc, got %v", err)
		}
	})

	t.Run("empty token", func(t *testing.T) {
		t.Parallel()
		_, err := v.Verify("   ")
		if err != ErrInvalidToken {
			t.Fatalf("expected ErrInvalidToken, got %v", err)
		}
	})

	t.Run("expired", func(t *testing.T) {
		t.Parallel()
		token := makeToken("u1", "t1", "", nil, time.Now().Add(-time.Hour))
		_, err := v.Verify(token)
		if err != ErrInvalidToken {
			t.Fatalf("expected ErrInvalidToken, got %v", err)
		}
	})

	t.Run("missing sub", func(t *testing.T) {
		t.Parallel()
		token := makeToken("", "t1", "", nil, time.Now().Add(time.Hour))
		_, err := v.Verify(token)
		if err != ErrMissingSubject {
			t.Fatalf("expected ErrMissingSubject, got %v", err)
		}
	})

	t.Run("missing tenant", func(t *testing.T) {
		t.Parallel()
		token := makeToken("u1", "", "", nil, time.Now().Add(time.Hour))
		_, err := v.Verify(token)
		if err != ErrMissingTenantID {
			t.Fatalf("expected ErrMissingTenantID, got %v", err)
		}
	})

	t.Run("issuer required mismatch", func(t *testing.T) {
		t.Parallel()
		v2 := Verifier{
			KeyFunc: keyFunc,
			Config: VerifyConfig{
				RequireIssuer: true,
				Issuer:        "issuer-a",
			},
		}
		token := makeToken("u1", "t1", "issuer-b", nil, time.Now().Add(time.Hour))
		_, err := v2.Verify(token)
		if err != ErrInvalidToken {
			t.Fatalf("expected ErrInvalidToken, got %v", err)
		}
	})

	t.Run("audience required mismatch", func(t *testing.T) {
		t.Parallel()
		v2 := Verifier{
			KeyFunc: keyFunc,
			Config: VerifyConfig{
				RequireAudience: true,
				Audience:        "api",
			},
		}
		token := makeToken("u1", "t1", "", []string{"other"}, time.Now().Add(time.Hour))
		_, err := v2.Verify(token)
		if err != ErrInvalidToken {
			t.Fatalf("expected ErrInvalidToken, got %v", err)
		}
	})

	t.Run("audience required ok", func(t *testing.T) {
		t.Parallel()
		v2 := Verifier{
			KeyFunc: keyFunc,
			Config: VerifyConfig{
				RequireAudience: true,
				Audience:        "api",
			},
		}
		token := makeToken("u1", "t1", "", []string{"api"}, time.Now().Add(time.Hour))
		_, err := v2.Verify(token)
		if err != nil {
			t.Fatalf("unexpected err: %v", err)
		}
	})
}
