package auth

import (
	"errors"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	// ErrMissingKeyFunc indicates verifier misconfiguration.
	ErrMissingKeyFunc = errors.New("auth: missing jwt keyfunc")

	// ErrInvalidToken indicates token is malformed, signature invalid, or claims invalid.
	// Do not expose underlying parsing details to clients.
	ErrInvalidToken = errors.New("auth: invalid token")

	// ErrMissingSubject indicates the token has no "sub".
	ErrMissingSubject = errors.New("auth: missing subject")

	// ErrMissingTenantID indicates the token has no tenant_id.
	ErrMissingTenantID = errors.New("auth: missing tenant_id")
)

type Claims struct {
	TenantID string   `json:"tenant_id"`
	Scopes   []string `json:"scopes,omitempty"`
	jwt.RegisteredClaims
}

// VerifyConfig controls optional JWT validation rules.
type VerifyConfig struct {
	// AllowedMethods defaults to HS256, RS256, ES256 if empty.
	AllowedMethods []string

	// Leeway allows small clock skew.
	Leeway time.Duration

	// Optional issuer/audience checks.
	Issuer          string
	RequireIssuer   bool
	Audience        string
	RequireAudience bool
}

type Verifier struct {
	// KeyFunc is used to provide the verification key based on token header (kid, alg, etc).
	KeyFunc jwt.Keyfunc
	Config  VerifyConfig
}

// ParseBearer extracts the raw JWT from an Authorization header value.
// Accepts "Bearer <token>" (case-insensitive). Returns ErrInvalidToken on invalid format.
func ParseBearer(authorization string) (string, error) {
	authorization = strings.TrimSpace(authorization)
	if authorization == "" {
		return "", ErrInvalidToken
	}
	parts := strings.Fields(authorization)
	if len(parts) != 2 {
		return "", ErrInvalidToken
	}
	if !strings.EqualFold(parts[0], "Bearer") {
		return "", ErrInvalidToken
	}
	tok := strings.TrimSpace(parts[1])
	if tok == "" {
		return "", ErrInvalidToken
	}
	return tok, nil
}

func (v Verifier) Verify(token string) (*Claims, error) {
	if v.KeyFunc == nil {
		return nil, ErrMissingKeyFunc
	}

	token = strings.TrimSpace(token)
	if token == "" {
		return nil, ErrInvalidToken
	}

	methods := v.Config.AllowedMethods
	if len(methods) == 0 {
		methods = []string{
			jwt.SigningMethodHS256.Alg(),
			jwt.SigningMethodRS256.Alg(),
			jwt.SigningMethodES256.Alg(),
		}
	}

	parsed, err := jwt.ParseWithClaims(
		token,
		&Claims{},
		v.KeyFunc,
		jwt.WithLeeway(v.Config.Leeway),
		jwt.WithValidMethods(methods),
	)
	if err != nil {
		return nil, ErrInvalidToken
	}
	if !parsed.Valid {
		return nil, ErrInvalidToken
	}

	claims, ok := parsed.Claims.(*Claims)
	if !ok || claims == nil {
		return nil, ErrInvalidToken
	}

	if strings.TrimSpace(claims.Subject) == "" {
		return nil, ErrMissingSubject
	}
	if strings.TrimSpace(claims.TenantID) == "" {
		return nil, ErrMissingTenantID
	}

	// Optional issuer check
	if v.Config.RequireIssuer {
		if strings.TrimSpace(v.Config.Issuer) == "" {
			// misconfiguration: treat as invalid token, but don't leak details
			return nil, ErrInvalidToken
		}
		if claims.Issuer != v.Config.Issuer {
			return nil, ErrInvalidToken
		}
	}

	// Optional audience check
	if v.Config.RequireAudience {
		aud := strings.TrimSpace(v.Config.Audience)
		if aud == "" {
			return nil, ErrInvalidToken
		}

		ok := false
		for _, a := range claims.Audience {
			if strings.TrimSpace(a) == aud {
				ok = true
				break
			}
		}
		if !ok {
			return nil, ErrInvalidToken
		}
	}

	return claims, nil
}
