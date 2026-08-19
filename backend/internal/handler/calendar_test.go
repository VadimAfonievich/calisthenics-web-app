package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	cal "github.com/calisthenics-coach/calisthenics-mini-app/backend/internal/calendar"
)

type calendarStub struct{}

func (calendarStub) Calendar(context.Context, string, string, string) ([]cal.Occurrence, error) {
	return nil, nil
}
func (calendarStub) Today(context.Context, string) ([]cal.Occurrence, error) { return nil, nil }
func (calendarStub) CreateSchedule(context.Context, string, cal.ScheduleInput) (cal.Schedule, error) {
	return cal.Schedule{}, nil
}
func (calendarStub) ListSchedules(context.Context, string) ([]cal.Schedule, error) { return nil, nil }
func (calendarStub) UpdateSchedule(context.Context, string, string, cal.ScheduleInput) (cal.Schedule, error) {
	return cal.Schedule{}, nil
}
func (calendarStub) DeleteSchedule(context.Context, string, string) error { return nil }
func (calendarStub) CreatePlanned(context.Context, string, cal.PlannedInput) (cal.PlannedWorkout, error) {
	return cal.PlannedWorkout{}, nil
}
func (calendarStub) GetPlanned(context.Context, string, string) (cal.PlannedWorkout, error) {
	return cal.PlannedWorkout{}, nil
}
func (calendarStub) UpdatePlanned(context.Context, string, string, cal.PlannedInput) (cal.PlannedWorkout, error) {
	return cal.PlannedWorkout{}, nil
}
func (calendarStub) DeletePlanned(context.Context, string, string) error { return nil }
func (calendarStub) SkipPlanned(context.Context, string, string) error   { return nil }

func TestCalendarPathUUIDValidation(t *testing.T) {
	for _, h := range []http.HandlerFunc{scheduleByID(calendarStub{}), plannedByID(calendarStub{}), skipPlanned(calendarStub{})} {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodDelete, "/undefined", nil)
		ctx := context.WithValue(r.Context(), struct{}{}, "")
		r = r.WithContext(ctx)
		h.ServeHTTP(w, r)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("status=%d", w.Code)
		}
	}
}
