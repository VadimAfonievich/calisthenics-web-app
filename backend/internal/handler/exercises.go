package handler

import (
	"context"
	"errors"
	"github.com/calisthenics-coach/calisthenics-mini-app/backend/internal/exercises"
	"github.com/go-chi/chi/v5"
	"net/http"
)

type ExerciseStore interface {
	List(context.Context, string, string) ([]exercises.Exercise, error)
	Get(context.Context, string) (exercises.Exercise, error)
}

func listExercises(store ExerciseStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		items, e := store.List(r.Context(), r.URL.Query().Get("difficulty"), r.URL.Query().Get("muscle_group"))
		if e != nil {
			writeError(w, 500, "EXERCISES_UNAVAILABLE", "Could not load exercises")
			return
		}
		writeJSON(w, 200, map[string]any{"exercises": items})
	}
}
func getExercise(store ExerciseStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		item, e := store.Get(r.Context(), chi.URLParam(r, "id"))
		if errors.Is(e, exercises.ErrNotFound) {
			writeError(w, 404, "EXERCISE_NOT_FOUND", "Exercise not found")
			return
		}
		if e != nil {
			writeError(w, 500, "EXERCISES_UNAVAILABLE", "Could not load exercise")
			return
		}
		writeJSON(w, 200, map[string]any{"exercise": item})
	}
}
