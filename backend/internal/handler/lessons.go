package handler

import (
	"context"
	"errors"
	"github.com/calisthenics-coach/calisthenics-mini-app/backend/internal/lessons"
	"github.com/calisthenics-coach/calisthenics-mini-app/backend/internal/middleware"
	"github.com/go-chi/chi/v5"
	"net/http"
)

type LessonStore interface {
	List(context.Context, string) ([]lessons.Lesson, error)
	Get(context.Context, string, string) (lessons.Lesson, error)
	Complete(context.Context, string, string) (lessons.Completion, error)
}

func lessonUserID(request *http.Request) (string, bool) { return middleware.UserID(request.Context()) }
func listLessons(store LessonStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := lessonUserID(r)
		if !ok {
			writeError(w, 401, "UNAUTHORIZED", "Authentication is required")
			return
		}
		items, err := store.List(r.Context(), id)
		if err != nil {
			writeError(w, 500, "LESSONS_UNAVAILABLE", "Could not load lessons")
			return
		}
		writeJSON(w, 200, map[string]any{"lessons": items})
	}
}
func getLesson(store LessonStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := lessonUserID(r)
		if !ok {
			writeError(w, 401, "UNAUTHORIZED", "Authentication is required")
			return
		}
		item, err := store.Get(r.Context(), id, chi.URLParam(r, "id"))
		if errors.Is(err, lessons.ErrNotFound) {
			writeError(w, 404, "LESSON_NOT_FOUND", "Lesson not found")
			return
		}
		if err != nil {
			writeError(w, 500, "LESSONS_UNAVAILABLE", "Could not load lesson")
			return
		}
		writeJSON(w, 200, map[string]any{"lesson": item})
	}
}
func completeLesson(store LessonStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := lessonUserID(r)
		if !ok {
			writeError(w, 401, "UNAUTHORIZED", "Authentication is required")
			return
		}
		result, err := store.Complete(r.Context(), id, chi.URLParam(r, "id"))
		if errors.Is(err, lessons.ErrNotFound) {
			writeError(w, 404, "LESSON_NOT_FOUND", "Lesson not found")
			return
		}
		if err != nil {
			writeError(w, 500, "LESSON_COMPLETION_FAILED", "Could not complete lesson")
			return
		}
		writeJSON(w, 200, result)
	}
}
