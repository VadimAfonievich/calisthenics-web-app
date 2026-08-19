package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/calisthenics-coach/calisthenics-mini-app/backend/internal/skills"
	"github.com/go-chi/chi/v5"
)

type skillsStub struct {
	err     error
	mastery skills.Mastery
}

func (s skillsStub) List(context.Context, string) ([]skills.Skill, error) {
	return []skills.Skill{{ID: validEntityID, Code: "HANDSTAND", Name: "Стойка", Status: "available", TotalLevels: 5}}, s.err
}
func (s skillsStub) Get(context.Context, string, string) (skills.Detail, error) {
	return skills.Detail{Skill: skills.Skill{ID: validEntityID, Status: "in_progress"}, Levels: []skills.Level{{ID: validEntityID, LevelNumber: 1, Status: "completed"}}}, s.err
}
func (s skillsStub) Map(context.Context, string) (skills.Map, error) {
	return skills.Map{Nodes: []skills.Skill{{ID: validEntityID, Status: "locked"}}, Requirements: []skills.Requirement{{SkillID: validEntityID, RequiredSkillID: "30000000-0000-0000-0000-000000000001", Type: "skill_mastered"}}}, s.err
}
func (s skillsStub) CompleteLevel(context.Context, string, string, int32, int32) error { return s.err }
func (s skillsStub) Master(context.Context, string, string, int32) (skills.Mastery, error) {
	return s.mastery, s.err
}

func TestSkillListAndMapContracts(t *testing.T) {
	for _, handler := range []http.HandlerFunc{listSkills(skillsStub{}), skillMap(skillsStub{})} {
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/", nil))
		if w.Code != 200 || !strings.Contains(w.Body.String(), `"status"`) {
			t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
		}
	}
}
func TestInvalidSkillUUIDRejected(t *testing.T) {
	router := chi.NewRouter()
	router.Get("/skills/{id}", getSkill(skillsStub{}))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/skills/undefined", nil))
	if w.Code != 400 || !strings.Contains(w.Body.String(), "INVALID_INPUT") {
		t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
	}
}
func TestSkillLockedAndCriterionErrors(t *testing.T) {
	for _, err := range []error{skills.ErrLocked, skills.ErrCriterion} {
		w := httptest.NewRecorder()
		skillError(w, err)
		if w.Code != 403 && w.Code != 409 {
			t.Fatalf("status=%d", w.Code)
		}
	}
}
func TestMasteryIdempotencyContract(t *testing.T) {
	out := skills.Mastery{SkillID: validEntityID, Status: "mastered", AlreadyMastered: true}
	router := chi.NewRouter()
	router.Post("/skills/{id}/master", masterSkill(skillsStub{mastery: out}))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/skills/"+validEntityID+"/master", strings.NewReader(`{"value":10}`)))
	if w.Code != 200 || !strings.Contains(w.Body.String(), `"already_mastered":true`) || !strings.Contains(w.Body.String(), `"xp_earned":0`) {
		t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
	}
}
