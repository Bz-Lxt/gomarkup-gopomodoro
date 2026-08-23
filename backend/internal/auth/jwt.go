package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"gopomodoro/internal/model"
	"gopomodoro/internal/timeutil"
)

type Claims struct {
	UserID string `json:"uid"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

type Tokens struct {
	Token     string `json:"token"`
	ExpiresAt string `json:"expires_at"`
}

func Issue(secret string, user *model.User, ttl time.Duration) (*Tokens, error) {
	now := timeutil.Now()
	exp := now.Add(ttl)
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		UserID: user.ID.String(),
		Email:  user.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(exp),
			Subject:   user.ID.String(),
		},
	})
	signed, err := t.SignedString([]byte(secret))
	if err != nil {
		return nil, err
	}
	return &Tokens{Token: signed, ExpiresAt: timeutil.FormatDateTime(exp)}, nil
}

func Parse(secret, raw string) (*Claims, error) {
	t, err := jwt.ParseWithClaims(raw, &Claims{}, func(token *jwt.Token) (any, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}
	c, ok := t.Claims.(*Claims)
	if !ok || !t.Valid {
		return nil, fmt.Errorf("invalid claims")
	}
	return c, nil
}
