package handler

import (
	"context"
	"github.com/calisthenics-coach/calisthenics-mini-app/backend/internal/middleware"
	"github.com/calisthenics-coach/calisthenics-mini-app/backend/internal/progress"
	"net/http"
)

type ProgressStore interface {
	Summary(context.Context, string) (progress.Summary, error)
	Stats(context.Context, string) (progress.Stats, error)
	History(context.Context, string) ([]progress.History, error)
	Achievements(context.Context, string) ([]progress.Achievement, error)
}

func progressHandler(load func(context.Context, string) (any, error), code string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u, _ := middleware.UserID(r.Context())
		value, err := load(r.Context(), u)
		if err != nil {
			writeError(w, http.StatusInternalServerError, code, "Could not load progress")
			return
		}
		writeJSON(w, http.StatusOK, value)
	}
}
func summary(s ProgressStore) http.HandlerFunc {
	return progressHandler(func(c context.Context, u string) (any, error) { return s.Summary(c, u) }, "PROGRESS_UNAVAILABLE")
}
func stats(s ProgressStore) http.HandlerFunc {
	return progressHandler(func(c context.Context, u string) (any, error) { return s.Stats(c, u) }, "STATS_UNAVAILABLE")
}
func history(s ProgressStore) http.HandlerFunc {
	return progressHandler(func(c context.Context, u string) (any, error) { return s.History(c, u) }, "HISTORY_UNAVAILABLE")
}
func achievements(s ProgressStore) http.HandlerFunc {
	return progressHandler(func(c context.Context, u string) (any, error) { return s.Achievements(c, u) }, "ACHIEVEMENTS_UNAVAILABLE")
}
