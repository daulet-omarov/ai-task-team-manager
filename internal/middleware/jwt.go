package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/daulet-omarov/ai-task-team-manager/internal/response"
	"github.com/daulet-omarov/ai-task-team-manager/pkg/jwt"
)

type contextKey string

const userContextKey = contextKey("userID")

func JWTMiddleware(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		authHeader := r.Header.Get("Authorization")

		if authHeader == "" {
			response.Error(w, http.StatusUnauthorized, "missing token")
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		userID, err := jwt.ParseToken(tokenString)
		if err != nil {
			response.Error(w, http.StatusUnauthorized, "invalid token")
			return
		}

		ctx := context.WithValue(r.Context(), userContextKey, userID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetUserID(r *http.Request) int64 {
	return r.Context().Value(userContextKey).(int64)
}
