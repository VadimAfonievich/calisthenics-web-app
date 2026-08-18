package handler

import (
	"context"
	"errors"
	"github.com/calisthenics-coach/calisthenics-mini-app/backend/internal/programs"
	"github.com/go-chi/chi/v5"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type programStoreStub struct {
	items []programs.Program
	item  programs.Program
	err   error
}

func (s programStoreStub) List(context.Context) ([]programs.Program, error) { return s.items, s.err }
func (s programStoreStub) Get(context.Context, string) (programs.Program, error) {
	return s.item, s.err
}
func TestListProgramsContract(t *testing.T) {
	store := programStoreStub{items: []programs.Program{{ID: "1", Name: "Published", Slug: "published", Difficulty: "beginner", DurationWeeks: 4}}}
	w := httptest.NewRecorder()
	listPrograms(store).ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/programs", nil))
	if w.Code != 200 || !strings.Contains(w.Body.String(), `"programs"`) || !strings.Contains(w.Body.String(), `"duration_weeks":4`) {
		t.Fatalf("got %d: %s", w.Code, w.Body.String())
	}
}
func TestGetProgramNotFound(t *testing.T) {
	store := programStoreStub{err: programs.ErrNotFound}
	r := httptest.NewRequest(http.MethodGet, "/programs/missing", nil)
	route := chi.NewRouter()
	route.Get("/programs/{id}", getProgram(store))
	w := httptest.NewRecorder()
	route.ServeHTTP(w, r)
	if w.Code != 404 || !strings.Contains(w.Body.String(), "PROGRAM_NOT_FOUND") {
		t.Fatalf("got %d: %s", w.Code, w.Body.String())
	}
}
func TestListProgramsFailure(t *testing.T) {
	w := httptest.NewRecorder()
	listPrograms(programStoreStub{err: errors.New("database")}).ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/programs", nil))
	if w.Code != 500 {
		t.Fatalf("got %d", w.Code)
	}
}
func TestProgramsRouteRequiresAuth(t *testing.T) {
	router := NewRouter(Dependencies{Auth: AuthDependencies{JWTSecret: "secret"}})
	w := httptest.NewRecorder()
	router.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/programs", nil))
	if w.Code != 401 {
		t.Fatalf("got %d: %s", w.Code, w.Body.String())
	}
}
