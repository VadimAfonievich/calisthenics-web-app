package handler

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/calisthenics-coach/calisthenics-mini-app/backend/internal/middleware"
	"github.com/calisthenics-coach/calisthenics-mini-app/backend/internal/workouts"
	"github.com/go-chi/chi/v5"
	"net/http"
	"strconv"
)

type WorkoutStore interface {
	Today(context.Context) (workouts.Workout, error)
	Get(context.Context, string) (workouts.Workout, error)
	Start(context.Context, string, string) (workouts.Session, error)
	RecordSet(context.Context, string, string, workouts.SetInput) error
	Complete(context.Context, string, string, int32) (workouts.Session, error)
}

func wu(r *http.Request) (string, bool) { return middleware.UserID(r.Context()) }
func wh(e error) int {
	if errors.Is(e, workouts.ErrNotFound) {
		return 404
	}
	if errors.Is(e, workouts.ErrForbidden) {
		return 403
	}
	return 500
}
func today(s WorkoutStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		x, e := s.Today(r.Context())
		if e != nil {
			writeError(w, wh(e), "WORKOUT_NOT_FOUND", "Workout not found")
			return
		}
		writeJSON(w, 200, map[string]any{"workout": x})
	}
}
func getWorkout(s WorkoutStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		x, e := s.Get(r.Context(), chi.URLParam(r, "id"))
		if e != nil {
			writeError(w, wh(e), "WORKOUT_NOT_FOUND", "Workout not found")
			return
		}
		writeJSON(w, 200, map[string]any{"workout": x})
	}
}
func start(s WorkoutStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u, _ := wu(r)
		x, e := s.Start(r.Context(), u, chi.URLParam(r, "id"))
		if e != nil {
			writeError(w, wh(e), "WORKOUT_START_FAILED", "Could not start workout")
			return
		}
		writeJSON(w, 201, map[string]any{"session": x})
	}
}
func set(s WorkoutStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u, _ := wu(r)
		var x workouts.SetInput
		if e := json.NewDecoder(r.Body).Decode(&x); e != nil || x.ExerciseID == "" || x.Number < 1 || (x.Reps == nil && x.Duration == nil) {
			writeError(w, 400, "INVALID_SET", "A reps or duration result is required")
			return
		}
		if e := s.RecordSet(r.Context(), u, chi.URLParam(r, "id"), x); e != nil {
			writeError(w, wh(e), "SET_UPDATE_FAILED", "Could not record set")
			return
		}
		w.WriteHeader(204)
	}
}
func complete(s WorkoutStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u, _ := wu(r)
		d, _ := strconv.ParseInt(r.URL.Query().Get("duration_seconds"), 10, 32)
		x, e := s.Complete(r.Context(), u, chi.URLParam(r, "id"), int32(d))
		if e != nil {
			writeError(w, wh(e), "WORKOUT_COMPLETION_FAILED", "Could not complete workout")
			return
		}
		writeJSON(w, 200, map[string]any{"session": x})
	}
}
