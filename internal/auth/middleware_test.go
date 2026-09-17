package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestMiddlewareAddsAuthenticatedUserID(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-jwt-secret")
	token := testToken(t, time.Hour)
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, ok := UserID(r.Context())
		if !ok || userID != 42 {
			t.Fatalf("expected user ID 42, got %d (present=%v)", userID, ok)
		}
		w.WriteHeader(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodPost, "/query", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	res := httptest.NewRecorder()
	Middleware(next).ServeHTTP(res, req)

	if res.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, res.Code)
	}
}

func TestMiddlewareRejectsInvalidToken(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-jwt-secret")
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("next handler should not be called")
	})

	req := httptest.NewRequest(http.MethodPost, "/query", nil)
	req.Header.Set("Authorization", "Bearer invalid")
	res := httptest.NewRecorder()
	Middleware(next).ServeHTTP(res, req)

	if res.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, res.Code)
	}
}

func TestMiddlewareAllowsAnonymousRequestForPublicOperations(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, ok := UserID(r.Context()); ok {
			t.Fatal("anonymous request should not have a user ID")
		}
		w.WriteHeader(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodPost, "/query", nil)
	res := httptest.NewRecorder()
	Middleware(next).ServeHTTP(res, req)

	if res.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, res.Code)
	}
}

func testToken(t *testing.T, validFor time.Duration) string {
	t.Helper()
	claims := jwt.RegisteredClaims{
		Subject:   "42",
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(validFor)),
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte("test-jwt-secret"))
	if err != nil {
		t.Fatalf("sign test token: %v", err)
	}
	return token
}

func TestRequireUserIDRejectsMissingContext(t *testing.T) {
	if _, err := RequireUserID(context.Background()); err != ErrMissingToken {
		t.Fatalf("expected missing token error, got %v", err)
	}
}
