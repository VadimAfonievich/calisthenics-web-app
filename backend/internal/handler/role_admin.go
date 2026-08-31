package handler

import (
	"context"
	"encoding/json"
	"github.com/calisthenics-coach/calisthenics-mini-app/backend/internal/middleware"
	"github.com/calisthenics-coach/calisthenics-mini-app/backend/internal/users"
	"github.com/go-chi/chi/v5"
	"net/http"
	"strings"
)

type roleUserStore interface {
	SearchRoleUsers(context.Context, string) ([]users.RoleUser, error)
	SetCoachRole(context.Context, string, string, string) error
	GetByID(context.Context, string) (users.User, error)
}
type coachSpaceStore interface {
	SetCoachSpace(context.Context, string, string, string, string, string) error
}

func superAdminUsers(store UserStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s, ok := store.(roleUserStore)
		if !ok {
			writeError(w, 503, "ADMIN_UNAVAILABLE", "Role management unavailable")
			return
		}
		actor, _ := middleware.UserID(r.Context())
		me, err := s.GetByID(r.Context(), actor)
		if err != nil || me.Role != "super_admin" {
			writeError(w, 403, "SUPER_ADMIN_REQUIRED", "Super administrator role required")
			return
		}
		items, err := s.SearchRoleUsers(r.Context(), strings.TrimSpace(r.URL.Query().Get("search")))
		if err != nil {
			writeError(w, 500, "USERS_UNAVAILABLE", "Could not search users")
			return
		}
		writeJSON(w, 200, map[string]any{"users": items})
	}
}

func superAdminUserRole(store UserStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s, ok := store.(roleUserStore)
		if !ok {
			writeError(w, 503, "ADMIN_UNAVAILABLE", "Role management unavailable")
			return
		}
		actor, _ := middleware.UserID(r.Context())
		me, err := s.GetByID(r.Context(), actor)
		if err != nil || me.Role != "super_admin" {
			writeError(w, 403, "SUPER_ADMIN_REQUIRED", "Super administrator role required")
			return
		}
		var body struct {
			Role string `json:"role"`
			Name string `json:"name"`
			Slug string `json:"slug"`
		}
		if json.NewDecoder(r.Body).Decode(&body) != nil || (body.Role != "user" && body.Role != "coach") {
			writeError(w, 400, "INVALID_ROLE", "Role must be user or coach")
			return
		}
		body.Name = strings.TrimSpace(body.Name)
		body.Slug = strings.TrimSpace(body.Slug)
		if body.Role == "coach" && (body.Name == "" || body.Slug == "") {
			writeError(w, 400, "TENANT_DETAILS_REQUIRED", "Tenant name and slug are required")
			return
		}
		if spaces, ok := store.(coachSpaceStore); ok {
			err = spaces.SetCoachSpace(r.Context(), actor, chiURLParam(r, "id"), body.Role, body.Name, body.Slug)
		} else {
			err = s.SetCoachRole(r.Context(), actor, chiURLParam(r, "id"), body.Role)
		}
		if err != nil {
			writeError(w, 409, "ROLE_CHANGE_REJECTED", err.Error())
			return
		}
		writeJSON(w, 200, map[string]string{"role": body.Role, "slug": body.Slug})
	}
}

func chiURLParam(r *http.Request, key string) string { return chi.URLParam(r, key) }
