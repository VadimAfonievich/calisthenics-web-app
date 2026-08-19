package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	cal "github.com/calisthenics-coach/calisthenics-mini-app/backend/internal/calendar"
	"github.com/go-chi/chi/v5"
)

type CalendarStore interface {
	Calendar(context.Context, string, string, string) ([]cal.Occurrence, error)
	Today(context.Context, string) ([]cal.Occurrence, error)
	CreateSchedule(context.Context, string, cal.ScheduleInput) (cal.Schedule, error)
	ListSchedules(context.Context, string) ([]cal.Schedule, error)
	UpdateSchedule(context.Context, string, string, cal.ScheduleInput) (cal.Schedule, error)
	DeleteSchedule(context.Context, string, string) error
	CreatePlanned(context.Context, string, cal.PlannedInput) (cal.PlannedWorkout, error)
	GetPlanned(context.Context, string, string) (cal.PlannedWorkout, error)
	UpdatePlanned(context.Context, string, string, cal.PlannedInput) (cal.PlannedWorkout, error)
	DeletePlanned(context.Context, string, string) error
	SkipPlanned(context.Context, string, string) error
}

func calendarError(w http.ResponseWriter, e error, notFound string) {
	switch {
	case errors.Is(e, cal.ErrInvalid):
		writeError(w, 400, "INVALID_INPUT", "Calendar input is invalid")
	case errors.Is(e, cal.ErrNotFound):
		writeError(w, 404, notFound, "Calendar item not found")
	case errors.Is(e, cal.ErrForbidden):
		writeError(w, 403, "FORBIDDEN", "Calendar item belongs to another user")
	default:
		writeError(w, 500, "CALENDAR_UNAVAILABLE", "Could not update calendar")
	}
}
func calendarRange(s CalendarStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u, _ := wu(r)
		x, e := s.Calendar(r.Context(), u, r.URL.Query().Get("from"), r.URL.Query().Get("to"))
		if e != nil {
			calendarError(w, e, "CALENDAR_NOT_FOUND")
			return
		}
		writeJSON(w, 200, map[string]any{"occurrences": x})
	}
}
func calendarToday(s CalendarStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u, _ := wu(r)
		x, e := s.Today(r.Context(), u)
		if e != nil {
			calendarError(w, e, "CALENDAR_NOT_FOUND")
			return
		}
		writeJSON(w, 200, map[string]any{"occurrences": x})
	}
}
func schedules(s CalendarStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u, _ := wu(r)
		if r.Method == http.MethodGet {
			x, e := s.ListSchedules(r.Context(), u)
			if e != nil {
				calendarError(w, e, "SCHEDULE_NOT_FOUND")
				return
			}
			writeJSON(w, 200, map[string]any{"schedules": x})
			return
		}
		var in cal.ScheduleInput
		if json.NewDecoder(r.Body).Decode(&in) != nil || !validUUID(in.WorkoutID) {
			writeError(w, 400, "INVALID_INPUT", "Schedule input is invalid")
			return
		}
		x, e := s.CreateSchedule(r.Context(), u, in)
		if e != nil {
			calendarError(w, e, "WORKOUT_NOT_FOUND")
			return
		}
		writeJSON(w, 201, map[string]any{"schedule": x})
	}
}
func scheduleByID(s CalendarStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		if !validUUID(id) {
			writeError(w, 400, "INVALID_INPUT", "Schedule id must be a valid UUID")
			return
		}
		u, _ := wu(r)
		if r.Method == http.MethodDelete {
			if e := s.DeleteSchedule(r.Context(), u, id); e != nil {
				calendarError(w, e, "SCHEDULE_NOT_FOUND")
				return
			}
			w.WriteHeader(204)
			return
		}
		var in cal.ScheduleInput
		if json.NewDecoder(r.Body).Decode(&in) != nil || !validUUID(in.WorkoutID) {
			writeError(w, 400, "INVALID_INPUT", "Schedule input is invalid")
			return
		}
		x, e := s.UpdateSchedule(r.Context(), u, id, in)
		if e != nil {
			calendarError(w, e, "SCHEDULE_NOT_FOUND")
			return
		}
		writeJSON(w, 200, map[string]any{"schedule": x})
	}
}
func planned(s CalendarStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u, _ := wu(r)
		var in cal.PlannedInput
		if json.NewDecoder(r.Body).Decode(&in) != nil || !validUUID(in.WorkoutID) || (in.SourceScheduleID != nil && !validUUID(*in.SourceScheduleID)) {
			writeError(w, 400, "INVALID_INPUT", "Planned workout input is invalid")
			return
		}
		x, e := s.CreatePlanned(r.Context(), u, in)
		if e != nil {
			calendarError(w, e, "WORKOUT_NOT_FOUND")
			return
		}
		writeJSON(w, 201, map[string]any{"planned_workout": x})
	}
}
func plannedByID(s CalendarStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		if !validUUID(id) {
			writeError(w, 400, "INVALID_INPUT", "Planned workout id must be a valid UUID")
			return
		}
		u, _ := wu(r)
		switch r.Method {
		case http.MethodGet:
			x, e := s.GetPlanned(r.Context(), u, id)
			if e != nil {
				calendarError(w, e, "PLANNED_WORKOUT_NOT_FOUND")
				return
			}
			writeJSON(w, 200, map[string]any{"planned_workout": x})
		case http.MethodDelete:
			if e := s.DeletePlanned(r.Context(), u, id); e != nil {
				calendarError(w, e, "PLANNED_WORKOUT_NOT_FOUND")
				return
			}
			w.WriteHeader(204)
		default:
			var in cal.PlannedInput
			if json.NewDecoder(r.Body).Decode(&in) != nil || !validUUID(in.WorkoutID) {
				writeError(w, 400, "INVALID_INPUT", "Planned workout input is invalid")
				return
			}
			x, e := s.UpdatePlanned(r.Context(), u, id, in)
			if e != nil {
				calendarError(w, e, "PLANNED_WORKOUT_NOT_FOUND")
				return
			}
			writeJSON(w, 200, map[string]any{"planned_workout": x})
		}
	}
}
func skipPlanned(s CalendarStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		if !validUUID(id) {
			writeError(w, 400, "INVALID_INPUT", "Planned workout id must be a valid UUID")
			return
		}
		u, _ := wu(r)
		if e := s.SkipPlanned(r.Context(), u, id); e != nil {
			calendarError(w, e, "PLANNED_WORKOUT_NOT_FOUND")
			return
		}
		w.WriteHeader(204)
	}
}
