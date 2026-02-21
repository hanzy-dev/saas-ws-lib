package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	TenantID string   `json:"tenant_id"`
	Scopes   []string `json:"scopes,omitempty"`
	jwt.RegisteredClaims
}

type Verifier struct {
	KeyFunc jwt.Keyfunc
	Leeway  time.Duration
}

func (v Verifier) Verify(token string) (*Claims, error) {
	if v.KeyFunc == nil {
		return nil, errors.New("jwt: missing keyfunc")
	}

	parsed, err := jwt.ParseWithClaims(
		token,
		&Claims{},
		v.KeyFunc,
		jwt.WithLeeway(v.Leeway),
		jwt.WithValidMethods([]string{
			jwt.SigningMethodHS256.Alg(),
			jwt.SigningMethodRS256.Alg(),
			jwt.SigningMethodES256.Alg(),
		}),
	)
	if err != nil {
		return nil, err
	}

	if !parsed.Valid {
		return nil, errors.New("jwt: invalid token")
	}

	claims, ok := parsed.Claims.(*Claims)
	if !ok || claims == nil {
		return nil, errors.New("jwt: invalid claims")
	}

	// RegisteredClaims.Subject is the canonical "sub"
	if claims.Subject == "" {
		return nil, errors.New("jwt: missing sub")
	}
	if claims.TenantID == "" {
		return nil, errors.New("jwt: missing tenant_id")
	}

	return claims, nil
}
