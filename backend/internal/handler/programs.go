package handler

import (
	"context"
	"errors"
	"net/http"

	"github.com/calisthenics-coach/calisthenics-mini-app/backend/internal/programs"
	"github.com/go-chi/chi/v5"
)

type ProgramStore interface {
	List(context.Context) ([]programs.Program, error)
	Get(context.Context, string) (programs.Program, error)
}

func listPrograms(store ProgramStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		items, err := store.List(r.Context())
		if err != nil {
			writeError(w, 500, "PROGRAMS_UNAVAILABLE", "Could not load programs")
			return
		}
		writeJSON(w, 200, map[string]any{"programs": items})
	}
}

func getProgram(store ProgramStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		item, err := store.Get(r.Context(), chi.URLParam(r, "id"))
		if errors.Is(err, programs.ErrNotFound) {
			writeError(w, 404, "PROGRAM_NOT_FOUND", "Program not found")
			return
		}
		if err != nil {
			writeError(w, 500, "PROGRAMS_UNAVAILABLE", "Could not load program")
			return
		}
		writeJSON(w, 200, map[string]any{"program": item})
	}
}
