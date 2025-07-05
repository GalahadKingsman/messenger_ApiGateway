package middleware

import (
	"context"
	"log"
	"net/http"
	"strings"

	"messenger_frontend/internal/jwt"
)

type contextKey string

const UserIDKey = contextKey("userID")

func JWTAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if r.Method == http.MethodPost && (r.URL.Path == "/users/login" || r.URL.Path == "/users/register") {
			next.ServeHTTP(w, r)
			return
		}

		auth := r.Header.Get("Authorization")
		if !strings.HasPrefix(auth, "Bearer ") {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		tokenStr := strings.TrimPrefix(auth, "Bearer ")

		userID, err := jwt.ValidateToken(tokenStr)
		if err != nil {
			log.Printf("JWT validation failed: %v", err)
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), UserIDKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
