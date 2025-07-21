package middleware

import (
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

func generateValidToken(t *testing.T, userID string) string {
	t.Helper()
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	secret := os.Getenv("SECRETKEY")
	if secret == "" {
		t.Fatal("SECRETKEY env variable is not set")
	}

	signedToken, err := token.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("Failed to sign token: %v", err)
	}
	return signedToken
}

func TestJWTAuthMiddleware_ValidToken(t *testing.T) {
	validToken := generateValidToken(t, "42")

	req := httptest.NewRequest(http.MethodGet, "/some-path", nil)
	req.Header.Set("Authorization", "Bearer "+validToken)
	rr := httptest.NewRecorder()

	var userID string
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctxVal := r.Context().Value(UserIDKey)
		userID, _ = ctxVal.(string)
		w.WriteHeader(http.StatusOK)
	})

	JWTAuthMiddleware(handler).ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "42", userID)
}

func TestJWTAuthMiddleware_SkipOnLogin(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/users/login", nil)
	rr := httptest.NewRecorder()

	var called bool
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})

	JWTAuthMiddleware(handler).ServeHTTP(rr, req)

	assert.True(t, called)
	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestJWTAuthMiddleware_NoAuthHeader(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called")
	})

	JWTAuthMiddleware(handler).ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestJWTAuthMiddleware_InvalidToken(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer invalid.token.here")
	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called")
	})

	JWTAuthMiddleware(handler).ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}
