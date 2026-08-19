package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/calisthenics-coach/calisthenics-mini-app/backend/internal/middleware"
	"github.com/calisthenics-coach/calisthenics-mini-app/backend/internal/skills"
	"github.com/go-chi/chi/v5"
)

type SkillsStore interface {
	List(context.Context, string) ([]skills.Skill, error)
	Get(context.Context, string, string) (skills.Detail, error)
	Map(context.Context, string) (skills.Map, error)
	CompleteLevel(context.Context, string, string, int32, int32) error
	Master(context.Context, string, string, int32) (skills.Mastery, error)
}

func skillError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, skills.ErrNotFound):
		writeError(w, 404, "SKILL_NOT_FOUND", "Skill or level not found")
	case errors.Is(err, skills.ErrLocked):
		writeError(w, 403, "SKILL_LOCKED", "Skill prerequisites are not completed")
	case errors.Is(err, skills.ErrCriterion):
		writeError(w, 409, "CRITERION_NOT_MET", "Required criterion is not met")
	default:
		writeError(w, 500, "SKILLS_UNAVAILABLE", "Could not process skill progression")
	}
}
func listSkills(s SkillsStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u, _ := middleware.UserID(r.Context())
		items, e := s.List(r.Context(), u)
		if e != nil {
			skillError(w, e)
			return
		}
		writeJSON(w, 200, map[string]any{"skills": items})
	}
}
func skillMap(s SkillsStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u, _ := middleware.UserID(r.Context())
		out, e := s.Map(r.Context(), u)
		if e != nil {
			skillError(w, e)
			return
		}
		writeJSON(w, 200, out)
	}
}
func getSkill(s SkillsStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		if !validUUID(id) {
			writeError(w, 400, "INVALID_INPUT", "Skill id must be a valid UUID")
			return
		}
		u, _ := middleware.UserID(r.Context())
		out, e := s.Get(r.Context(), u, id)
		if e != nil {
			skillError(w, e)
			return
		}
		writeJSON(w, 200, out)
	}
}
func criterionValue(w http.ResponseWriter, r *http.Request) (int32, bool) {
	var in struct {
		Value int32 `json:"value"`
	}
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if e := decoder.Decode(&in); e != nil || in.Value < 0 {
		writeError(w, 400, "INVALID_INPUT", "A non-negative criterion value is required")
		return 0, false
	}
	return in.Value, true
}
func completeSkillLevel(s SkillsStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		level, err := strconv.ParseInt(chi.URLParam(r, "level"), 10, 32)
		if !validUUID(id) || err != nil || level < 1 {
			writeError(w, 400, "INVALID_INPUT", "Skill id and level must be valid")
			return
		}
		value, ok := criterionValue(w, r)
		if !ok {
			return
		}
		u, _ := middleware.UserID(r.Context())
		if err = s.CompleteLevel(r.Context(), u, id, int32(level), value); err != nil {
			skillError(w, err)
			return
		}
		writeJSON(w, 200, map[string]any{"status": "completed", "level_number": level})
	}
}
func masterSkill(s SkillsStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		if !validUUID(id) {
			writeError(w, 400, "INVALID_INPUT", "Skill id must be a valid UUID")
			return
		}
		value, ok := criterionValue(w, r)
		if !ok {
			return
		}
		u, _ := middleware.UserID(r.Context())
		out, e := s.Master(r.Context(), u, id, value)
		if e != nil {
			skillError(w, e)
			return
		}
		writeJSON(w, 200, out)
	}
}
