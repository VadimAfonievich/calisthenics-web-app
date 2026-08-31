package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"regexp"
	"time"

	"github.com/calisthenics-coach/calisthenics-mini-app/backend/internal/auth"
	"github.com/calisthenics-coach/calisthenics-mini-app/backend/internal/middleware"
	"github.com/calisthenics-coach/calisthenics-mini-app/backend/internal/users"
)

type UserStore interface {
	UpsertTelegramUser(ctx context.Context, telegramUser auth.TelegramUser) (users.User, error)
	GetByID(ctx context.Context, id string) (users.User, error)
}

type AuthDependencies struct {
	BotToken  string
	JWTSecret string
	Users     UserStore
}
type telegramAuthRequest struct {
	InitData string `json:"init_data"`
}

type tenantResolver interface {
	ResolveTenant(context.Context, string, string) (*users.Tenant, error)
}
type tenantBootstrapper interface {
	BootstrapTenant(context.Context, string, string) (*users.Tenant, error)
}

var tenantSlug = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

func telegramAuthHandler(dependencies AuthDependencies) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		if dependencies.Users == nil || dependencies.BotToken == "" {
			writeError(writer, http.StatusServiceUnavailable, "AUTH_UNAVAILABLE", "Telegram authentication is not configured")
			return
		}
		var payload telegramAuthRequest
		decoder := json.NewDecoder(http.MaxBytesReader(writer, request.Body, 16<<10))
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&payload); err != nil || payload.InitData == "" {
			writeError(writer, http.StatusBadRequest, "INVALID_REQUEST", "init_data is required")
			return
		}
		telegramUser, err := auth.ValidateInitData(payload.InitData, dependencies.BotToken, time.Now(), 24*time.Hour)
		if err != nil {
			code := "INVALID_TELEGRAM_INIT_DATA"
			if errors.Is(err, auth.ErrExpiredInitData) {
				code = "EXPIRED_TELEGRAM_INIT_DATA"
			}
			writeError(writer, http.StatusUnauthorized, code, "Telegram authentication data is invalid")
			return
		}
		user, err := dependencies.Users.UpsertTelegramUser(request.Context(), telegramUser)
		if err != nil {
			writeError(writer, http.StatusInternalServerError, "USER_PROVISIONING_FAILED", "Could not provision user")
			return
		}
		values, _ := url.ParseQuery(payload.InitData)
		startParam := values.Get("start_param") // covered by Telegram's validated signature
		if startParam != "" && (!tenantSlug.MatchString(startParam) || len(startParam) > 63) {
			writeError(writer, http.StatusBadRequest, "INVALID_TENANT_SLUG", "Coach space link is invalid")
			return
		}
		if resolver, ok := dependencies.Users.(tenantBootstrapper); ok {
			tenant, resolveErr := resolver.BootstrapTenant(request.Context(), user.ID, startParam)
			if resolveErr != nil {
				writeError(writer, http.StatusNotFound, "TENANT_NOT_FOUND", "Coach space is unavailable")
				return
			}
			user.CurrentTenant = tenant
		}
		token, err := auth.IssueToken(user.ID, dependencies.JWTSecret, time.Now(), 24*time.Hour)
		if err != nil {
			writeError(writer, http.StatusInternalServerError, "TOKEN_ISSUANCE_FAILED", "Could not issue access token")
			return
		}
		writeJSON(writer, http.StatusOK, map[string]any{"access_token": token, "token_type": "Bearer", "expires_in": 86400, "user": user})
	}
}

func meHandler(dependencies AuthDependencies) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		userID, ok := middleware.UserID(request.Context())
		if !ok {
			writeError(writer, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication is required")
			return
		}
		user, err := dependencies.Users.GetByID(request.Context(), userID)
		if err != nil {
			writeError(writer, http.StatusNotFound, "USER_NOT_FOUND", "User not found")
			return
		}
		writeJSON(writer, http.StatusOK, map[string]any{"user": user})
	}
}
