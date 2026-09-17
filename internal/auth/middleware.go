package auth

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

const userIDContextKey contextKey = "authenticated-user-id"

var (
	ErrMissingToken = errors.New("authentication required")
	ErrInvalidToken = errors.New("invalid authentication token")
)

func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenString := strings.TrimSpace(strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer "))
		if tokenString != "" {
			if userID, err := parseToken(tokenString); err == nil {
				r = r.WithContext(WithUserID(r.Context(), userID))
			} else {
				http.Error(w, ErrInvalidToken.Error(), http.StatusUnauthorized)
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}

func WithUserID(ctx context.Context, userID int64) context.Context {
	return context.WithValue(ctx, userIDContextKey, userID)
}

func UserID(ctx context.Context) (int64, bool) {
	userID, ok := ctx.Value(userIDContextKey).(int64)
	return userID, ok
}

func RequireUserID(ctx context.Context) (int64, error) {
	userID, ok := UserID(ctx)
	if !ok {
		return 0, ErrMissingToken
	}
	return userID, nil
}

func parseToken(tokenString string) (int64, error) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return 0, fmt.Errorf("%w: JWT_SECRET is not configured", ErrInvalidToken)
	}

	token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(token *jwt.Token) (any, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method: %s", token.Method.Alg())
		}
		return []byte(secret), nil
	})
	if err != nil || !token.Valid {
		return 0, ErrInvalidToken
	}

	claims, ok := token.Claims.(*jwt.RegisteredClaims)
	if !ok || claims.Subject == "" {
		return 0, ErrInvalidToken
	}

	var userID int64
	if _, err := fmt.Sscan(claims.Subject, &userID); err != nil || userID <= 0 {
		return 0, ErrInvalidToken
	}
	return userID, nil
}
