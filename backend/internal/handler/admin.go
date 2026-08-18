package handler

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/calisthenics-coach/calisthenics-mini-app/backend/internal/admin"
	"github.com/calisthenics-coach/calisthenics-mini-app/backend/internal/middleware"
	"github.com/go-chi/chi/v5"
	"net/http"
)

type AdminAuthorizer interface {
	IsAdmin(context.Context, string) (bool, error)
}
type AdminStore interface {
	IsAdmin(context.Context, string) (bool, error)
	CreateLesson(context.Context, admin.Lesson) (string, error)
	UpdateLesson(context.Context, string, admin.Lesson) (string, error)
	PublishLesson(context.Context, string, bool) (string, error)
	DeleteLesson(context.Context, string) error
	CreateExercise(context.Context, admin.Exercise) (string, error)
	UpdateExercise(context.Context, string, admin.Exercise) (string, error)
	DeleteExercise(context.Context, string) error
	CreateProgram(context.Context, admin.Program) (string, error)
	UpdateProgram(context.Context, string, admin.Program) (string, error)
	PublishProgram(context.Context, string, bool) (string, error)
	DeleteProgram(context.Context, string) error
	CreateWorkout(context.Context, admin.Workout) (string, error)
	UpdateWorkout(context.Context, string, admin.Workout) (string, error)
	DeleteWorkout(context.Context, string) error
	UpsertWorkoutExercise(context.Context, string, admin.WorkoutExercise) (string, error)
	DeleteWorkoutExercise(context.Context, string, string) error
}

func requireAdmin(s AdminAuthorizer, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u, ok := middleware.UserID(r.Context())
		if !ok {
			writeError(w, 401, "UNAUTHORIZED", "Authentication is required")
			return
		}
		yes, err := s.IsAdmin(r.Context(), u)
		if err != nil {
			writeError(w, 500, "ADMIN_UNAVAILABLE", "Could not check administrator role")
			return
		}
		if !yes {
			writeError(w, 403, "ADMIN_REQUIRED", "Administrator role is required")
			return
		}
		next.ServeHTTP(w, r)
	})
}
func decode(w http.ResponseWriter, r *http.Request, v any) bool {
	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		writeError(w, 400, "INVALID_INPUT", "Request body is invalid")
		return false
	}
	return true
}
func adminError(w http.ResponseWriter, e error) {
	if errors.Is(e, admin.ErrNotFound) {
		writeError(w, 404, "CONTENT_NOT_FOUND", "Content not found")
		return
	}
	writeError(w, 400, "CONTENT_UPDATE_FAILED", "Could not save content")
}
func createLesson(s AdminStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var x admin.Lesson
		if !decode(w, r, &x) || x.CategoryID == "" || x.Title == "" || x.Slug == "" || x.Content == "" || x.DurationMinutes < 1 {
			return
		}
		id, e := s.CreateLesson(r.Context(), x)
		if e != nil {
			adminError(w, e)
			return
		}
		writeJSON(w, 201, map[string]string{"id": id})
	}
}
func updateLesson(s AdminStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var x admin.Lesson
		if !decode(w, r, &x) {
			return
		}
		id, e := s.UpdateLesson(r.Context(), chi.URLParam(r, "id"), x)
		if e != nil {
			adminError(w, e)
			return
		}
		writeJSON(w, 200, map[string]string{"id": id})
	}
}
func publishLesson(s AdminStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var x struct {
			Published bool `json:"published"`
		}
		if !decode(w, r, &x) {
			return
		}
		id, e := s.PublishLesson(r.Context(), chi.URLParam(r, "id"), x.Published)
		if e != nil {
			adminError(w, e)
			return
		}
		writeJSON(w, 200, map[string]string{"id": id})
	}
}
func deleteLesson(s AdminStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if e := s.DeleteLesson(r.Context(), chi.URLParam(r, "id")); e != nil {
			adminError(w, e)
			return
		}
		w.WriteHeader(204)
	}
}
func createExercise(s AdminStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var x admin.Exercise
		if !decode(w, r, &x) || x.Name == "" || x.Slug == "" || x.Description == "" || len(x.MuscleGroups) == 0 {
			return
		}
		id, e := s.CreateExercise(r.Context(), x)
		if e != nil {
			adminError(w, e)
			return
		}
		writeJSON(w, 201, map[string]string{"id": id})
	}
}
func updateExercise(s AdminStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var x admin.Exercise
		if !decode(w, r, &x) {
			return
		}
		id, e := s.UpdateExercise(r.Context(), chi.URLParam(r, "id"), x)
		if e != nil {
			adminError(w, e)
			return
		}
		writeJSON(w, 200, map[string]string{"id": id})
	}
}
func deleteExercise(s AdminStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if e := s.DeleteExercise(r.Context(), chi.URLParam(r, "id")); e != nil {
			adminError(w, e)
			return
		}
		w.WriteHeader(204)
	}
}
func programInput(s AdminStore, create bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var x admin.Program
		if !decode(w, r, &x) || x.Name == "" || x.Slug == "" || x.DurationWeeks < 1 {
			return
		}
		var id string
		var e error
		if create {
			id, e = s.CreateProgram(r.Context(), x)
		} else {
			id, e = s.UpdateProgram(r.Context(), chi.URLParam(r, "id"), x)
		}
		if e != nil {
			adminError(w, e)
			return
		}
		status := 200
		if create {
			status = 201
		}
		writeJSON(w, status, map[string]string{"id": id})
	}
}
func publishProgram(s AdminStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var x struct {
			Published bool `json:"published"`
		}
		if !decode(w, r, &x) {
			return
		}
		id, e := s.PublishProgram(r.Context(), chi.URLParam(r, "id"), x.Published)
		if e != nil {
			adminError(w, e)
			return
		}
		writeJSON(w, 200, map[string]string{"id": id})
	}
}
func deleteProgram(s AdminStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if e := s.DeleteProgram(r.Context(), chi.URLParam(r, "id")); e != nil {
			adminError(w, e)
			return
		}
		w.WriteHeader(204)
	}
}
func workoutInput(s AdminStore, create bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var x admin.Workout
		if !decode(w, r, &x) || x.ProgramID == "" || x.Title == "" || x.DayNumber < 1 || x.EstimatedMinutes < 1 {
			return
		}
		var id string
		var e error
		if create {
			id, e = s.CreateWorkout(r.Context(), x)
		} else {
			id, e = s.UpdateWorkout(r.Context(), chi.URLParam(r, "id"), x)
		}
		if e != nil {
			adminError(w, e)
			return
		}
		status := 200
		if create {
			status = 201
		}
		writeJSON(w, status, map[string]string{"id": id})
	}
}
func deleteWorkout(s AdminStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if e := s.DeleteWorkout(r.Context(), chi.URLParam(r, "id")); e != nil {
			adminError(w, e)
			return
		}
		w.WriteHeader(204)
	}
}
func upsertWorkoutExercise(s AdminStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var x admin.WorkoutExercise
		if !decode(w, r, &x) || x.ExerciseID == "" || x.Sets < 1 {
			return
		}
		id, e := s.UpsertWorkoutExercise(r.Context(), chi.URLParam(r, "id"), x)
		if e != nil {
			adminError(w, e)
			return
		}
		writeJSON(w, 201, map[string]string{"id": id})
	}
}
func deleteWorkoutExercise(s AdminStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if e := s.DeleteWorkoutExercise(r.Context(), chi.URLParam(r, "id"), chi.URLParam(r, "exerciseID")); e != nil {
			adminError(w, e)
			return
		}
		w.WriteHeader(204)
	}
}
