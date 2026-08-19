package handler

import (
	"context"
	"github.com/calisthenics-coach/calisthenics-mini-app/backend/internal/auth"
	c "github.com/calisthenics-coach/calisthenics-mini-app/backend/internal/coach"
	"github.com/calisthenics-coach/calisthenics-mini-app/backend/internal/middleware"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type coachStub struct{ role c.Role }

func (s coachStub) Role(context.Context, string) (c.Role, error) { return s.role, nil }
func (coachStub) Dashboard(context.Context, string, c.Role) (c.Dashboard, error) {
	return c.Dashboard{}, nil
}
func (coachStub) Analytics(context.Context) (c.Analytics, error) { return c.Analytics{}, nil }
func (coachStub) List(context.Context, string, string, c.Role, string, string) ([]c.Item, error) {
	return nil, nil
}
func (coachStub) Get(context.Context, string, string, string, c.Role) (map[string]any, error) {
	return map[string]any{"id": "10000000-0000-0000-0000-000000000001"}, nil
}
func (coachStub) Options(context.Context, string, c.Role) (c.Options, error) { return c.Options{}, nil }
func (coachStub) SaveLesson(context.Context, string, c.Role, *string, c.LessonInput) (string, error) {
	return "", nil
}
func (coachStub) SaveBuilder(context.Context, string, string, c.Role, *string, c.BuilderInput) (string, error) {
	return "", nil
}
func (coachStub) Lifecycle(context.Context, string, string, string, c.Role, string) error { return nil }
func (coachStub) Duplicate(context.Context, string, string, string, c.Role) (string, error) {
	return "", nil
}
func (coachStub) ListMedia(context.Context, string, c.Role) ([]c.Media, error) { return nil, nil }
func (coachStub) CreateExternalMedia(context.Context, string, c.MediaInput) (c.Media, error) {
	return c.Media{}, nil
}
func (coachStub) DeleteMedia(context.Context, string, string, c.Role) error { return nil }
func TestCoachAuthorizationRoles(t *testing.T) {
	const secret = "coach-test-secret-which-is-long-enough"
	token, e := auth.IssueToken("10000000-0000-0000-0000-000000000001", secret, time.Now(), time.Hour)
	if e != nil {
		t.Fatal(e)
	}
	for _, tc := range []struct {
		role c.Role
		want int
	}{{"user", 403}, {"coach", 200}, {"admin", 200}, {"super_admin", 200}} {
		next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(200) })
		h := middleware.RequireAuth(secret, requireCoach(coachStub{tc.role}, next))
		r := httptest.NewRequest("GET", "/api/v1/coach/dashboard", nil)
		r.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		if w.Code != tc.want {
			t.Fatalf("role=%s status=%d", tc.role, w.Code)
		}
	}
}
