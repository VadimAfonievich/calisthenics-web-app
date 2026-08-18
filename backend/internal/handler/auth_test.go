package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/calisthenics-coach/calisthenics-mini-app/backend/internal/auth"
	"github.com/calisthenics-coach/calisthenics-mini-app/backend/internal/users"
)

type fakeUserStore struct{ user users.User }

func (s fakeUserStore) UpsertTelegramUser(context.Context, auth.TelegramUser) (users.User, error) {
	return s.user, nil
}
func (s fakeUserStore) GetByID(context.Context, string) (users.User, error) { return s.user, nil }

func TestMeRequiresBearerToken(t *testing.T) {
	router := NewRouter(Dependencies{Auth: AuthDependencies{JWTSecret: "secret", Users: fakeUserStore{}}})
	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/v1/me", nil))
	if response.Code != http.StatusUnauthorized || !strings.Contains(response.Body.String(), "UNAUTHORIZED") {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
	}
}

func TestMeReturnsAuthenticatedUser(t *testing.T) {
	user := users.User{ID: "11111111-1111-1111-1111-111111111111", FirstName: "Анна", DisplayName: "Анна", Level: 1, Timezone: "UTC"}
	token, err := auth.IssueToken(user.ID, "secret", time.Now(), time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	router := NewRouter(Dependencies{Auth: AuthDependencies{JWTSecret: "secret", Users: fakeUserStore{user: user}}})
	request := httptest.NewRequest(http.MethodGet, "/api/v1/me", nil)
	request.Header.Set("Authorization", "Bearer "+token)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), "Анна") {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
	}
}
