package handler

import (
	"context"
	"errors"
	coachsvc "github.com/calisthenics-coach/calisthenics-mini-app/backend/internal/coach"
	"github.com/calisthenics-coach/calisthenics-mini-app/backend/internal/middleware"
	"github.com/go-chi/chi/v5"
	"net/http"
	"strings"
)

type CoachStore interface {
	Role(context.Context, string) (coachsvc.Role, error)
	Dashboard(context.Context, string, coachsvc.Role) (coachsvc.Dashboard, error)
	Analytics(context.Context) (coachsvc.Analytics, error)
	List(context.Context, string, string, coachsvc.Role, string, string) ([]coachsvc.Item, error)
	SaveLesson(context.Context, string, coachsvc.Role, *string, coachsvc.LessonInput) (string, error)
	SaveBuilder(context.Context, string, string, coachsvc.Role, *string, coachsvc.BuilderInput) (string, error)
	Lifecycle(context.Context, string, string, string, coachsvc.Role, string) error
	Duplicate(context.Context, string, string, string, coachsvc.Role) (string, error)
	ListMedia(context.Context, string, coachsvc.Role) ([]coachsvc.Media, error)
	CreateExternalMedia(context.Context, string, coachsvc.MediaInput) (coachsvc.Media, error)
	DeleteMedia(context.Context, string, string, coachsvc.Role) error
}
type coachRoleKey struct{}

func requireCoach(s CoachStore, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u, ok := middleware.UserID(r.Context())
		if !ok {
			writeError(w, 401, "UNAUTHORIZED", "Authentication is required")
			return
		}
		role, e := s.Role(r.Context(), u)
		if e != nil {
			writeError(w, 500, "ROLE_UNAVAILABLE", "Could not check content role")
			return
		}
		if role != "coach" && role != "admin" && role != "super_admin" {
			writeError(w, 403, "COACH_REQUIRED", "Coach role is required")
			return
		}
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), coachRoleKey{}, role)))
	})
}
func coachIdentity(r *http.Request) (string, coachsvc.Role) {
	u, _ := middleware.UserID(r.Context())
	role, _ := r.Context().Value(coachRoleKey{}).(coachsvc.Role)
	return u, role
}
func coachErr(w http.ResponseWriter, e error) {
	switch {
	case errors.Is(e, coachsvc.ErrForbidden):
		writeError(w, 403, "FORBIDDEN", "Content belongs to another coach")
	case errors.Is(e, coachsvc.ErrNotFound):
		writeError(w, 404, "CONTENT_NOT_FOUND", "Content not found")
	case errors.Is(e, coachsvc.ErrInUse):
		writeError(w, 409, "MEDIA_IN_USE", "Media is referenced by content")
	case errors.Is(e, coachsvc.ErrInvalid):
		writeError(w, 400, "INVALID_INPUT", "Content validation failed")
	default:
		writeError(w, 500, "COACH_STUDIO_ERROR", "Coach Studio request failed")
	}
}
func coachMe(s CoachStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, role := coachIdentity(r)
		writeJSON(w, 200, map[string]any{"role": role})
	}
}
func coachDashboard(s CoachStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u, role := coachIdentity(r)
		x, e := s.Dashboard(r.Context(), u, role)
		if e != nil {
			coachErr(w, e)
			return
		}
		writeJSON(w, 200, x)
	}
}
func coachAnalytics(s CoachStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		x, e := s.Analytics(r.Context())
		if e != nil {
			coachErr(w, e)
			return
		}
		writeJSON(w, 200, x)
	}
}
func coachContent(s CoachStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		kind := chi.URLParam(r, "kind")
		u, role := coachIdentity(r)
		if r.Method == http.MethodGet {
			x, e := s.List(r.Context(), kind, u, role, strings.TrimSpace(r.URL.Query().Get("search")), r.URL.Query().Get("status"))
			if e != nil {
				coachErr(w, e)
				return
			}
			writeJSON(w, 200, map[string]any{"items": x})
			return
		}
		var id string
		var e error
		if kind == "lessons" {
			var in coachsvc.LessonInput
			if !decode(w, r, &in) {
				return
			}
			id, e = s.SaveLesson(r.Context(), u, role, nil, in)
		} else {
			var in coachsvc.BuilderInput
			if !decode(w, r, &in) {
				return
			}
			id, e = s.SaveBuilder(r.Context(), kind, u, role, nil, in)
		}
		if e != nil {
			coachErr(w, e)
			return
		}
		writeJSON(w, 201, map[string]string{"id": id})
	}
}
func coachContentID(s CoachStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		kind, id := chi.URLParam(r, "kind"), chi.URLParam(r, "id")
		if !validUUID(id) {
			invalidInput(w)
			return
		}
		u, role := coachIdentity(r)
		if r.Method == http.MethodPut {
			if kind != "lessons" {
				var in coachsvc.BuilderInput
				if !decode(w, r, &in) {
					return
				}
				out, e := s.SaveBuilder(r.Context(), kind, u, role, &id, in)
				if e != nil {
					coachErr(w, e)
					return
				}
				writeJSON(w, 200, map[string]string{"id": out})
				return
			}
			var in coachsvc.LessonInput
			if !decode(w, r, &in) {
				return
			}
			out, e := s.SaveLesson(r.Context(), u, role, &id, in)
			if e != nil {
				coachErr(w, e)
				return
			}
			writeJSON(w, 200, map[string]string{"id": out})
			return
		}
		writeError(w, 405, "METHOD_NOT_ALLOWED", "Unsupported content action")
	}
}
func coachAction(s CoachStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		kind, id, action := chi.URLParam(r, "kind"), chi.URLParam(r, "id"), chi.URLParam(r, "action")
		if !validUUID(id) {
			invalidInput(w)
			return
		}
		u, role := coachIdentity(r)
		if action == "duplicate" {
			out, e := s.Duplicate(r.Context(), kind, id, u, role)
			if e != nil {
				coachErr(w, e)
				return
			}
			writeJSON(w, 201, map[string]string{"id": out})
			return
		}
		status := map[string]string{"publish": "published", "unpublish": "draft", "archive": "archived"}[action]
		if status == "" {
			invalidInput(w)
			return
		}
		if e := s.Lifecycle(r.Context(), kind, id, u, role, status); e != nil {
			coachErr(w, e)
			return
		}
		writeJSON(w, 200, map[string]string{"id": id, "status": status})
	}
}
func coachMedia(s CoachStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u, role := coachIdentity(r)
		if r.Method == http.MethodGet {
			x, e := s.ListMedia(r.Context(), u, role)
			if e != nil {
				coachErr(w, e)
				return
			}
			writeJSON(w, 200, map[string]any{"media": x})
			return
		}
		var in coachsvc.MediaInput
		if !decode(w, r, &in) {
			return
		}
		x, e := s.CreateExternalMedia(r.Context(), u, in)
		if e != nil {
			coachErr(w, e)
			return
		}
		writeJSON(w, 201, map[string]any{"media": x})
	}
}
func coachUploadUnavailable(w http.ResponseWriter, _ *http.Request) {
	writeError(w, 503, "OBJECT_STORAGE_UNAVAILABLE", "Configure S3-compatible object storage; Render filesystem uploads are disabled")
}
func coachMediaDelete(s CoachStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		if !validUUID(id) {
			invalidInput(w)
			return
		}
		u, role := coachIdentity(r)
		if e := s.DeleteMedia(r.Context(), u, id, role); e != nil {
			coachErr(w, e)
			return
		}
		w.WriteHeader(204)
	}
}
