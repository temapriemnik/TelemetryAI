package middleware

import (
	"context"
	"net/http"
	"strings"

	"telemetryai/internal/usecase"
)

type contextKey string

const UserIDKey contextKey = "user_id"

type AuthMiddleware struct {
	authService *usecase.AuthService
}

func NewAuthMiddleware(authService *usecase.AuthService) *AuthMiddleware {
	return &AuthMiddleware{authService: authService}
}

func (m *AuthMiddleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		token := r.URL.Query().Get("token")
		
		if authHeader == "" && token == "" {
			http.Error(w, "missing authorization", http.StatusUnauthorized)
			return
		}

		var tokenStr string
		if authHeader != "" {
			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				http.Error(w, "invalid authorization header", http.StatusUnauthorized)
				return
			}
			tokenStr = parts[1]
		} else {
			tokenStr = token
		}

		userID, err := m.authService.ValidateToken(tokenStr)
		if err != nil {
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), UserIDKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}