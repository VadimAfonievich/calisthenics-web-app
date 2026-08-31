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
const tenantIDKey contextKey = "tenant_id"
const tenantRoleKey contextKey = "tenant_role"
const platformRoleKey contextKey = "platform_role"

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

func WithTenant(ctx context.Context, id, role string) context.Context {
	return context.WithValue(context.WithValue(ctx, tenantIDKey, id), tenantRoleKey, role)
}
func TenantID(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(tenantIDKey).(string)
	return id, ok
}
func TenantRole(ctx context.Context) (string, bool) {
	role, ok := ctx.Value(tenantRoleKey).(string)
	return role, ok
}
func WithPlatformRole(ctx context.Context, role string) context.Context {
	return context.WithValue(ctx, platformRoleKey, role)
}
func PlatformRole(ctx context.Context) (string, bool) {
	role, ok := ctx.Value(platformRoleKey).(string)
	return role, ok
}

func unauthorized(writer http.ResponseWriter) {
	writer.Header().Set("Content-Type", "application/json; charset=utf-8")
	writer.WriteHeader(http.StatusUnauthorized)
	_, _ = writer.Write([]byte(`{"error":{"code":"UNAUTHORIZED","message":"Authentication is required"}}\n`))
}
