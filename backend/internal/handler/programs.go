package handler

import (
	"context"
	"errors"
	"net/http"

	"github.com/calisthenics-coach/calisthenics-mini-app/backend/internal/middleware"
	"github.com/calisthenics-coach/calisthenics-mini-app/backend/internal/programs"
	"github.com/go-chi/chi/v5"
)

type ProgramStore interface {
	List(context.Context, string) ([]programs.Program, error)
	Get(context.Context, string, string) (programs.Program, error)
	Start(context.Context, string, string) (programs.Progress, error)
}

func listPrograms(store ProgramStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, _ := middleware.UserID(r.Context())
		items, err := store.List(r.Context(), user)
		if err != nil {
			writeError(w, 500, "PROGRAMS_UNAVAILABLE", "Could not load programs")
			return
		}
		writeJSON(w, 200, map[string]any{"programs": items})
	}
}

func getProgram(store ProgramStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, _ := middleware.UserID(r.Context())
		item, err := store.Get(r.Context(), user, chi.URLParam(r, "id"))
		if errors.Is(err, programs.ErrNotFound) {
			writeError(w, 404, "PROGRAM_NOT_FOUND", "Program not found")
			return
		}
		if errors.Is(err, programs.ErrForbidden) {
			writeError(w, 403, "PROGRAM_ACCESS_DENIED", "Program access denied")
			return
		}
		if err != nil {
			writeError(w, 500, "PROGRAMS_UNAVAILABLE", "Could not load program")
			return
		}
		writeJSON(w, 200, map[string]any{"program": item})
	}
}

func startProgram(store ProgramStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, _ := middleware.UserID(r.Context())
		progress, err := store.Start(r.Context(), user, chi.URLParam(r, "id"))
		if errors.Is(err, programs.ErrNotFound) {
			writeError(w, 404, "PROGRAM_NOT_FOUND", "Program not found")
			return
		}
		if errors.Is(err, programs.ErrForbidden) {
			writeError(w, 403, "PROGRAM_ACCESS_DENIED", "Program access denied")
			return
		}
		if err != nil {
			writeError(w, 500, "PROGRAMS_UNAVAILABLE", "Could not start program")
			return
		}
		writeJSON(w, 200, map[string]any{"progress": progress})
	}
}
