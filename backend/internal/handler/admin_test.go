package handler

import (
	"context"
	"github.com/calisthenics-coach/calisthenics-mini-app/backend/internal/admin"
	"github.com/calisthenics-coach/calisthenics-mini-app/backend/internal/auth"
	"github.com/calisthenics-coach/calisthenics-mini-app/backend/internal/middleware"
	"github.com/go-chi/chi/v5"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

type adminStoreStub struct {
	allowed         bool
	called          string
	lesson          admin.Lesson
	exercise        admin.Exercise
	program         admin.Program
	workout         admin.Workout
	workoutExercise admin.WorkoutExercise
}

func (s *adminStoreStub) IsAdmin(context.Context, string) (bool, error) { return s.allowed, nil }
func (s *adminStoreStub) CreateLesson(_ context.Context, x admin.Lesson) (string, error) {
	s.called = "create_lesson"
	s.lesson = x
	return "70000000-0000-0000-0000-000000000001", nil
}
func (s *adminStoreStub) UpdateLesson(_ context.Context, _ string, x admin.Lesson) (string, error) {
	s.called = "update_lesson"
	s.lesson = x
	return "70000000-0000-0000-0000-000000000001", nil
}
func (s *adminStoreStub) PublishLesson(context.Context, string, bool) (string, error) {
	return "70000000-0000-0000-0000-000000000001", nil
}
func (s *adminStoreStub) DeleteLesson(context.Context, string) error { return nil }
func (s *adminStoreStub) CreateExercise(_ context.Context, x admin.Exercise) (string, error) {
	s.called = "create_exercise"
	s.exercise = x
	return "70000000-0000-0000-0000-000000000002", nil
}
func (s *adminStoreStub) UpdateExercise(context.Context, string, admin.Exercise) (string, error) {
	return "70000000-0000-0000-0000-000000000002", nil
}
func (s *adminStoreStub) DeleteExercise(context.Context, string) error { return nil }
func (s *adminStoreStub) CreateProgram(_ context.Context, x admin.Program) (string, error) {
	s.called = "create_program"
	s.program = x
	return "70000000-0000-0000-0000-000000000003", nil
}
func (s *adminStoreStub) UpdateProgram(context.Context, string, admin.Program) (string, error) {
	return "70000000-0000-0000-0000-000000000003", nil
}
func (s *adminStoreStub) PublishProgram(context.Context, string, bool) (string, error) {
	return "70000000-0000-0000-0000-000000000003", nil
}
func (s *adminStoreStub) DeleteProgram(context.Context, string) error { return nil }
func (s *adminStoreStub) CreateWorkout(_ context.Context, x admin.Workout) (string, error) {
	s.called = "create_workout"
	s.workout = x
	return "70000000-0000-0000-0000-000000000004", nil
}
func (s *adminStoreStub) UpdateWorkout(context.Context, string, admin.Workout) (string, error) {
	return "70000000-0000-0000-0000-000000000004", nil
}
func (s *adminStoreStub) DeleteWorkout(context.Context, string) error { return nil }
func (s *adminStoreStub) UpsertWorkoutExercise(_ context.Context, _ string, x admin.WorkoutExercise) (string, error) {
	s.called = "upsert_workout_exercise"
	s.workoutExercise = x
	return "70000000-0000-0000-0000-000000000005", nil
}
func (s *adminStoreStub) DeleteWorkoutExercise(context.Context, string, string) error { return nil }
func adminRequest(method, path, body string) *http.Request {
	return httptest.NewRequest(method, path, strings.NewReader(body))
}
func TestAdminLessonSnakeCaseBinding(t *testing.T) {
	s := &adminStoreStub{}
	body := `{"category_id":"10000000-0000-0000-0000-000000000001","title":"Title","slug":"title","short_description":"Short","content":"Body","difficulty":"beginner","duration_minutes":5,"sort_order":2,"published":false}`
	w := httptest.NewRecorder()
	createLesson(s).ServeHTTP(w, adminRequest(http.MethodPost, "/", body))
	if w.Code != 201 || s.lesson.CategoryID == "" || s.lesson.ShortDescription != "Short" || s.lesson.DurationMinutes != 5 {
		t.Fatalf("got %d %#v: %s", w.Code, s.lesson, w.Body.String())
	}
}
func TestAdminLessonUpdateSnakeCaseBinding(t *testing.T) {
	s := &adminStoreStub{}
	body := `{"category_id":"10000000-0000-0000-0000-000000000001","title":"Updated","slug":"updated","short_description":"Short","content":"Body","difficulty":"intermediate","duration_minutes":8,"sort_order":1,"published":true}`
	route := chi.NewRouter()
	route.Put("/{id}", updateLesson(s))
	w := httptest.NewRecorder()
	route.ServeHTTP(w, adminRequest(http.MethodPut, "/70000000-0000-0000-0000-000000000001", body))
	if w.Code != 200 || s.called != "update_lesson" || s.lesson.Title != "Updated" {
		t.Fatalf("got %d %#v: %s", w.Code, s.lesson, w.Body.String())
	}
}
func TestAdminInvalidLessonReturns400(t *testing.T) {
	s := &adminStoreStub{}
	w := httptest.NewRecorder()
	createLesson(s).ServeHTTP(w, adminRequest(http.MethodPost, "/", `{"category_id":"not-a-uuid","title":"x"}`))
	if w.Code != 400 || s.called != "" || !strings.Contains(w.Body.String(), "INVALID_INPUT") {
		t.Fatalf("got %d called=%s: %s", w.Code, s.called, w.Body.String())
	}
}
func TestAdminExerciseSnakeCaseBinding(t *testing.T) {
	s := &adminStoreStub{}
	body := `{"name":"Push","slug":"push","description":"D","instructions":"I","common_mistakes":"M","difficulty":"beginner","muscle_groups":["chest"],"equipment":[]}`
	w := httptest.NewRecorder()
	createExercise(s).ServeHTTP(w, adminRequest(http.MethodPost, "/", body))
	if w.Code != 201 || s.exercise.CommonMistakes != "M" || len(s.exercise.MuscleGroups) != 1 {
		t.Fatalf("got %d %#v: %s", w.Code, s.exercise, w.Body.String())
	}
}
func TestAdminProgramSnakeCaseBinding(t *testing.T) {
	s := &adminStoreStub{}
	body := `{"name":"Base","slug":"base","description":"D","difficulty":"beginner","duration_weeks":4,"published":false}`
	w := httptest.NewRecorder()
	programInput(s, true).ServeHTTP(w, adminRequest(http.MethodPost, "/", body))
	if w.Code != 201 || s.program.DurationWeeks != 4 {
		t.Fatalf("got %d %#v: %s", w.Code, s.program, w.Body.String())
	}
}
func TestAdminWorkoutSnakeCaseBinding(t *testing.T) {
	s := &adminStoreStub{}
	body := `{"program_id":"40000000-0000-0000-0000-000000000001","day_number":2,"title":"Day","description":"D","estimated_minutes":30,"sort_order":1}`
	w := httptest.NewRecorder()
	workoutInput(s, true).ServeHTTP(w, adminRequest(http.MethodPost, "/", body))
	if w.Code != 201 || s.workout.ProgramID == "" || s.workout.EstimatedMinutes != 30 {
		t.Fatalf("got %d %#v: %s", w.Code, s.workout, w.Body.String())
	}
}
func TestAdminWorkoutExerciseSnakeCaseBinding(t *testing.T) {
	s := &adminStoreStub{}
	body := `{"exercise_id":"30000000-0000-0000-0000-000000000001","sets":3,"target_reps":8,"rest_seconds":60,"sort_order":1}`
	route := chi.NewRouter()
	route.Post("/{id}/exercises", upsertWorkoutExercise(s))
	w := httptest.NewRecorder()
	route.ServeHTTP(w, adminRequest(http.MethodPost, "/50000000-0000-0000-0000-000000000001/exercises", body))
	if w.Code != 201 || s.workoutExercise.TargetReps == nil || *s.workoutExercise.TargetReps != 8 {
		t.Fatalf("got %d %#v: %s", w.Code, s.workoutExercise, w.Body.String())
	}
}
func TestRequireAdminRejectsNonAdmin(t *testing.T) {
	s := &adminStoreStub{}
	token, err := auth.IssueToken("70000000-0000-0000-0000-000000000099", "secret", time.Now(), time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	next := middleware.RequireAuth("secret", requireAdmin(s, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(204) })))
	r := adminRequest(http.MethodPost, "/admin/lessons", "")
	r.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	next.ServeHTTP(w, r)
	if w.Code != 403 || !strings.Contains(w.Body.String(), "ADMIN_REQUIRED") {
		t.Fatalf("got %d: %s", w.Code, w.Body.String())
	}
}
