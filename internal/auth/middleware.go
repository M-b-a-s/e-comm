package auth

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"

	repo "github/M-b-a-s/e-comm/internal/adapters/postgresql/sqlc"

	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

const userIDContextKey contextKey = "authenticated-user-id"
const userRoleContextKey contextKey = "authenticated-user-role"

var (
	ErrMissingToken = errors.New("authentication required")
	ErrInvalidToken = errors.New("invalid authentication token")
)

func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenString := strings.TrimSpace(strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer "))
		if tokenString != "" {
			if claims, err := parseToken(tokenString); err == nil {
				requestContext := WithUserID(r.Context(), claims.SubjectID)
				requestContext = WithRole(requestContext, claims.Role)
				r = r.WithContext(requestContext)
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

func WithRole(ctx context.Context, role string) context.Context {
	return context.WithValue(ctx, userRoleContextKey, role)
}

func Role(ctx context.Context) (string, bool) {
	role, ok := ctx.Value(userRoleContextKey).(string)
	return role, ok
}

func RequireRole(ctx context.Context, requiredRole string) error {
	role, ok := Role(ctx)
	if !ok {
		return ErrMissingToken
	}
	if role != requiredRole {
		return fmt.Errorf("role %q is not authorized", role)
	}
	return nil
}

func RequireProductReader(ctx context.Context) error {
	role, ok := Role(ctx)
	if !ok {
		return ErrMissingToken
	}
	if role != string(repo.UserRoleCustomer) && role != string(repo.UserRoleAdmin) {
		return fmt.Errorf("role %q is not authorized to read products", role)
	}
	return nil
}

type parsedClaims struct {
	SubjectID int64
	Role      string
}

func parseToken(tokenString string) (parsedClaims, error) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return parsedClaims{}, fmt.Errorf("%w: JWT_SECRET is not configured", ErrInvalidToken)
	}

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (any, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method: %s", token.Method.Alg())
		}
		return []byte(secret), nil
	})
	if err != nil || !token.Valid {
		return parsedClaims{}, ErrInvalidToken
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || claims.Subject == "" || claims.Role == "" {
		return parsedClaims{}, ErrInvalidToken
	}

	var userID int64
	if _, err := fmt.Sscan(claims.Subject, &userID); err != nil || userID <= 0 {
		return parsedClaims{}, ErrInvalidToken
	}
	return parsedClaims{SubjectID: userID, Role: claims.Role}, nil
}
