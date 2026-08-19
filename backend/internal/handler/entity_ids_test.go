package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/calisthenics-coach/calisthenics-mini-app/backend/internal/auth"
	"github.com/calisthenics-coach/calisthenics-mini-app/backend/internal/lessons"
	"github.com/calisthenics-coach/calisthenics-mini-app/backend/internal/middleware"
	"github.com/calisthenics-coach/calisthenics-mini-app/backend/internal/workouts"
	"github.com/go-chi/chi/v5"
)

const validEntityID = "20000000-0000-0000-0000-000000000001"

type lessonIDStoreStub struct{ called bool }

func (s *lessonIDStoreStub) List(context.Context, string) ([]lessons.Lesson, error) {
	return []lessons.Lesson{{ID: validEntityID, CategoryID: "10000000-0000-0000-0000-000000000001", Title: "Lesson", DurationMinutes: 12}}, nil
}
func (s *lessonIDStoreStub) Get(context.Context, string, string) (lessons.Lesson, error) {
	s.called = true
	return lessons.Lesson{}, nil
}
func (s *lessonIDStoreStub) Complete(context.Context, string, string) (lessons.Completion, error) {
	s.called = true
	return lessons.Completion{}, nil
}

type workoutIDStoreStub struct {
	called bool
	err    error
}

func (s *workoutIDStoreStub) List(context.Context, string) ([]workouts.CatalogItem, error) {
	return []workouts.CatalogItem{{ID: validEntityID, Title: "Workout", Minutes: 30, Difficulty: "beginner", ExerciseCount: 8, ProgramID: validEntityID, ProgramName: "Program"}}, nil
}
func (s *workoutIDStoreStub) Today(context.Context) (workouts.Workout, error) {
	return workouts.Workout{ID: validEntityID, Title: "Workout", Minutes: 30, Exercises: []workouts.Exercise{{ID: validEntityID, Name: "Exercise", Sets: 1}}}, nil
}
func (s *workoutIDStoreStub) GetSession(context.Context, string, string) (workouts.ActiveSession, error) {
	s.called = true
	return workouts.ActiveSession{}, s.err
}
func (s *workoutIDStoreStub) Get(context.Context, string) (workouts.Workout, error) {
	s.called = true
	return workouts.Workout{}, nil
}
func (s *workoutIDStoreStub) Start(context.Context, string, string) (workouts.Session, error) {
	s.called = true
	return workouts.Session{}, nil
}
func (s *workoutIDStoreStub) RecordSet(context.Context, string, string, workouts.SetInput) error {
	s.called = true
	return nil
}
func (s *workoutIDStoreStub) Complete(context.Context, string, string, int32) (workouts.Session, error) {
	s.called = true
	return workouts.Session{}, nil
}

func TestLessonListItemUsesSnakeCaseID(t *testing.T) {
	token, err := auth.IssueToken(validEntityID, "secret", time.Now(), time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	request := httptest.NewRequest(http.MethodGet, "/lessons", nil)
	request.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	middleware.RequireAuth("secret", listLessons(&lessonIDStoreStub{})).ServeHTTP(w, request)
	body := w.Body.String()
	if w.Code != http.StatusOK || !strings.Contains(body, `"id":"`+validEntityID+`"`) || !strings.Contains(body, `"category_id":"`) || !strings.Contains(body, `"duration_minutes":12`) || strings.Contains(body, `"ID"`) {
		t.Fatalf("unexpected lesson JSON: %s", body)
	}
}

func TestWorkoutUsesSnakeCaseIDs(t *testing.T) {
	w := httptest.NewRecorder()
	today(&workoutIDStoreStub{}).ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/workouts/today", nil))
	body := w.Body.String()
	if !strings.Contains(body, `"id":"`+validEntityID+`"`) || !strings.Contains(body, `"estimated_minutes":30`) || strings.Contains(body, `"ID"`) {
		t.Fatalf("unexpected workout JSON: %s", body)
	}
}

func TestInvalidLessonUUIDReturns400WithoutStoreCall(t *testing.T) {
	store := &lessonIDStoreStub{}
	router := chi.NewRouter()
	router.Get("/lessons/{id}", getLesson(store))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/lessons/undefined", nil))
	if w.Code != http.StatusBadRequest || store.called || !strings.Contains(w.Body.String(), `"code":"INVALID_INPUT"`) {
		t.Fatalf("status=%d called=%v body=%s", w.Code, store.called, w.Body.String())
	}
}

func TestInvalidWorkoutUUIDReturns400WithoutStoreCall(t *testing.T) {
	store := &workoutIDStoreStub{}
	router := chi.NewRouter()
	router.Get("/workouts/{id}", getWorkout(store))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/workouts/undefined", nil))
	if w.Code != http.StatusBadRequest || store.called || !strings.Contains(w.Body.String(), `"code":"INVALID_INPUT"`) {
		t.Fatalf("status=%d called=%v body=%s", w.Code, store.called, w.Body.String())
	}
}

func TestWorkoutCatalogContract(t *testing.T) {
	w := httptest.NewRecorder()
	listWorkouts(&workoutIDStoreStub{}).ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/workouts", nil))
	body := w.Body.String()
	for _, field := range []string{`"id"`, `"estimated_minutes":30`, `"difficulty":"beginner"`, `"exercise_count":8`, `"program_id"`, `"program_name":"Program"`} {
		if !strings.Contains(body, field) {
			t.Fatalf("missing %s in %s", field, body)
		}
	}
}

func TestWorkoutSessionOwnershipFailure(t *testing.T) {
	store := &workoutIDStoreStub{err: workouts.ErrForbidden}
	router := chi.NewRouter()
	router.Get("/workout-sessions/{id}", getWorkoutSession(store))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/workout-sessions/20000000-0000-0000-0000-000000000001", nil))
	if w.Code != http.StatusForbidden || !store.called || !strings.Contains(w.Body.String(), "WORKOUT_SESSION_NOT_FOUND") {
		t.Fatalf("status=%d called=%v body=%s", w.Code, store.called, w.Body.String())
	}
}
