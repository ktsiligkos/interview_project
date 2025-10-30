package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrMissingSecret = errors.New("jwt secret is required")
	ErrInvalidToken  = errors.New("invalid token")
)

// Claims represents the payload embedded in JWTs.
type Claims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

// GenerateJWT creates a signed token for the specified user
func GenerateJWT(userID string, secret []byte, ttl time.Duration) (string, error) {
	if len(secret) == 0 {
		return "", ErrMissingSecret
	}

	now := time.Now()

	claims := Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secret)
}

// ParseJWT validates an incoming token string and returns the claims.
func ParseJWT(tokenString string, secret []byte) (*Claims, error) {
	if len(secret) == 0 {
		return nil, ErrMissingSecret
	}

	claims := &Claims{}

	parsedToken, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %s", t.Header["alg"])
		}

		return secret, nil
	})
	if err != nil {
		return nil, err
	}

	if !parsedToken.Valid {
		return nil, ErrInvalidToken
	}

	return claims, nil
}
