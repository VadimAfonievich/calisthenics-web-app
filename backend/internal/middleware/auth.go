package middleware

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/calisthenics-coach/calisthenics-mini-app/backend/internal/auth"
)

type contextKey string

const userIDKey contextKey = "user_id"

func RequireAuth(secret string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		header := request.Header.Get("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			unauthorized(writer)
			return
		}
		claims, err := auth.VerifyToken(strings.TrimPrefix(header, "Bearer "), secret, time.Now())
		if err != nil {
			unauthorized(writer)
			return
		}
		ctx := context.WithValue(request.Context(), userIDKey, claims.Subject)
		next.ServeHTTP(writer, request.WithContext(ctx))
	})
}

func UserID(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(userIDKey).(string)
	return id, ok
}

func unauthorized(writer http.ResponseWriter) {
	writer.Header().Set("Content-Type", "application/json; charset=utf-8")
	writer.WriteHeader(http.StatusUnauthorized)
	_, _ = writer.Write([]byte(`{"error":{"code":"UNAUTHORIZED","message":"Authentication is required"}}\n`))
}
