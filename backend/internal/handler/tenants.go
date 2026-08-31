package handler

import (
	"context"
	"github.com/calisthenics-coach/calisthenics-mini-app/backend/internal/middleware"
	"github.com/calisthenics-coach/calisthenics-mini-app/backend/internal/users"
	"github.com/go-chi/chi/v5"
	"net/http"
	"strings"
)

type tenantAdminStore interface {
	ListTenants(context.Context) ([]users.Tenant, error)
	GetTenant(context.Context, string) (users.Tenant, error)
	UpdateOwnTenant(context.Context, string, string, string, string) (users.Tenant, error)
}
type tenantSlugStore interface {
	UpdateOwnTenantSlug(context.Context, string, string, string) (users.Tenant, error)
}

func coachSpaceSlug(store UserStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s, ok := store.(tenantSlugStore)
		user, _ := middleware.UserID(r.Context())
		tenant, hasTenant := middleware.TenantID(r.Context())
		role, _ := middleware.TenantRole(r.Context())
		if !ok || !hasTenant || role != "coach" {
			writeError(w, 403, "COACH_TENANT_REQUIRED", "Coach tenant required")
			return
		}
		var body struct {
			Slug string `json:"slug"`
		}
		if !decode(w, r, &body) {
			return
		}
		x, err := s.UpdateOwnTenantSlug(r.Context(), user, tenant, body.Slug)
		if err != nil {
			writeError(w, 409, "TENANT_SLUG_REJECTED", err.Error())
			return
		}
		writeJSON(w, 200, map[string]any{"tenant": x})
	}
}

func adminTenants(store UserStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s, ok := store.(tenantAdminStore)
		role, _ := middleware.PlatformRole(r.Context())
		if !ok || role != "super_admin" {
			writeError(w, 403, "SUPER_ADMIN_REQUIRED", "Super administrator role required")
			return
		}
		items, e := s.ListTenants(r.Context())
		if e != nil {
			writeError(w, 500, "TENANTS_UNAVAILABLE", "Could not list tenants")
			return
		}
		writeJSON(w, 200, map[string]any{"tenants": items})
	}
}
func adminTenant(store UserStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s, ok := store.(tenantAdminStore)
		role, _ := middleware.PlatformRole(r.Context())
		if !ok || role != "super_admin" {
			writeError(w, 403, "SUPER_ADMIN_REQUIRED", "Super administrator role required")
			return
		}
		id := chi.URLParam(r, "id")
		if !validUUID(id) {
			invalidInput(w)
			return
		}
		x, e := s.GetTenant(r.Context(), id)
		if e != nil {
			writeError(w, 404, "TENANT_NOT_FOUND", "Tenant not found")
			return
		}
		writeJSON(w, 200, map[string]any{"tenant": x})
	}
}
func coachSpace(store UserStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s, ok := store.(tenantAdminStore)
		user, _ := middleware.UserID(r.Context())
		tenant, hasTenant := middleware.TenantID(r.Context())
		role, _ := middleware.TenantRole(r.Context())
		if !ok || !hasTenant || role != "coach" {
			writeError(w, 403, "COACH_TENANT_REQUIRED", "Coach tenant required")
			return
		}
		if r.Method == http.MethodGet {
			x, e := s.GetTenant(r.Context(), tenant)
			if e != nil {
				writeError(w, 404, "TENANT_NOT_FOUND", "Tenant not found")
				return
			}
			writeJSON(w, 200, map[string]any{"tenant": x})
			return
		}
		var body struct {
			Name        string `json:"name"`
			Description string `json:"description"`
		}
		if !decode(w, r, &body) {
			return
		}
		x, e := s.UpdateOwnTenant(r.Context(), user, tenant, strings.TrimSpace(body.Name), strings.TrimSpace(body.Description))
		if e != nil {
			writeError(w, 400, "TENANT_UPDATE_REJECTED", e.Error())
			return
		}
		writeJSON(w, 200, map[string]any{"tenant": x})
	}
}
