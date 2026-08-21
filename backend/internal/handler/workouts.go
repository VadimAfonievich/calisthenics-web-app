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
	List(context.Context, string) ([]workouts.CatalogItem, error)
	Today(context.Context) (workouts.Workout, error)
	Get(context.Context, string) (workouts.Workout, error)
	Start(context.Context, string, string, workouts.StartInput) (workouts.Session, error)
	GetSession(context.Context, string, string) (workouts.ActiveSession, error)
	RecordSet(context.Context, string, string, workouts.SetInput) error
	Complete(context.Context, string, string, int32) (workouts.Session, error)
}

func listWorkouts(s WorkoutStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u, _ := wu(r)
		items, err := s.List(r.Context(), u)
		if err != nil {
			writeError(w, 500, "WORKOUTS_UNAVAILABLE", "Could not load workouts")
			return
		}
		writeJSON(w, 200, map[string]any{"workouts": items})
	}
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
		workoutID := chi.URLParam(r, "id")
		if !validUUID(workoutID) {
			writeError(w, 400, "INVALID_INPUT", "Workout id must be a valid UUID")
			return
		}
		x, e := s.Get(r.Context(), workoutID)
		if e != nil {
			writeError(w, wh(e), "WORKOUT_NOT_FOUND", "Workout not found")
			return
		}
		writeJSON(w, 200, map[string]any{"workout": x})
	}
}
func start(s WorkoutStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		workoutID := chi.URLParam(r, "id")
		if !validUUID(workoutID) {
			writeError(w, 400, "INVALID_INPUT", "Workout id must be a valid UUID")
			return
		}
		u, _ := wu(r)
		var body struct {
			PlannedWorkoutID  *string `json:"planned_workout_id"`
			FollowUpWorkoutID *string `json:"follow_up_workout_id"`
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body.PlannedWorkoutID != nil && !validUUID(*body.PlannedWorkoutID) {
			writeError(w, 400, "INVALID_INPUT", "Planned workout id must be a valid UUID")
			return
		}
		if body.FollowUpWorkoutID != nil && (!validUUID(*body.FollowUpWorkoutID) || *body.FollowUpWorkoutID == workoutID) {
			writeError(w, 400, "INVALID_INPUT", "Follow-up workout id must be a different valid UUID")
			return
		}
		x, e := s.Start(r.Context(), u, workoutID, workouts.StartInput{PlannedWorkoutID: body.PlannedWorkoutID, FollowUpWorkoutID: body.FollowUpWorkoutID})
		if e != nil {
			writeError(w, wh(e), "WORKOUT_START_FAILED", "Could not start workout")
			return
		}
		writeJSON(w, 201, map[string]any{"session": x})
	}
}
func getWorkoutSession(s WorkoutStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sessionID := chi.URLParam(r, "id")
		if !validUUID(sessionID) {
			writeError(w, 400, "INVALID_INPUT", "Workout session id must be a valid UUID")
			return
		}
		u, _ := wu(r)
		x, err := s.GetSession(r.Context(), u, sessionID)
		if err != nil {
			writeError(w, wh(err), "WORKOUT_SESSION_NOT_FOUND", "Workout session not found")
			return
		}
		writeJSON(w, 200, x)
	}
}
func set(s WorkoutStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sessionID := chi.URLParam(r, "id")
		if !validUUID(sessionID) {
			writeError(w, 400, "INVALID_INPUT", "Workout session id must be a valid UUID")
			return
		}
		u, _ := wu(r)
		var x workouts.SetInput
		if e := json.NewDecoder(r.Body).Decode(&x); e != nil || x.ExerciseID == "" || x.Number < 1 || x.Number > 100 || (x.Reps == nil) == (x.Duration == nil) || (x.Reps != nil && *x.Reps < 0) || (x.Duration != nil && *x.Duration < 0) {
			writeError(w, 400, "INVALID_SET", "A reps or duration result is required")
			return
		}
		if !validUUID(x.ExerciseID) {
			writeError(w, 400, "INVALID_INPUT", "Exercise id must be a valid UUID")
			return
		}
		if e := s.RecordSet(r.Context(), u, sessionID, x); e != nil {
			writeError(w, wh(e), "SET_UPDATE_FAILED", "Could not record set")
			return
		}
		w.WriteHeader(204)
	}
}
func complete(s WorkoutStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sessionID := chi.URLParam(r, "id")
		if !validUUID(sessionID) {
			writeError(w, 400, "INVALID_INPUT", "Workout session id must be a valid UUID")
			return
		}
		u, _ := wu(r)
		d, err := strconv.ParseInt(r.URL.Query().Get("duration_seconds"), 10, 32)
		if err != nil || d < 0 || d > 12*60*60 {
			writeError(w, 400, "INVALID_DURATION", "duration_seconds must be between 0 and 43200")
			return
		}
		x, e := s.Complete(r.Context(), u, sessionID, int32(d))
		if e != nil {
			writeError(w, wh(e), "WORKOUT_COMPLETION_FAILED", "Could not complete workout")
			return
		}
		writeJSON(w, 200, map[string]any{"session": x})
	}
}
