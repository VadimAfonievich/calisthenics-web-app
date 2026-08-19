package calendar

import (
	"errors"
	"testing"
)

func TestScheduleValidation(t *testing.T) {
	tm := "19:00"
	for name, in := range map[string]ScheduleInput{
		"duplicate weekday": {StartDate: "2026-08-19", Weekdays: []int16{1, 1}, PreferredTime: &tm},
		"empty weekdays":    {StartDate: "2026-08-19"},
		"end before start":  {StartDate: "2026-08-19", EndDate: ptr("2026-08-18"), Weekdays: []int16{1}},
		"invalid time":      {StartDate: "2026-08-19", Weekdays: []int16{1}, PreferredTime: ptr("25:00")},
	} {
		t.Run(name, func(t *testing.T) {
			if _, e := parseInput(in, "UTC"); !errors.Is(e, ErrInvalid) {
				t.Fatalf("error=%v", e)
			}
		})
	}
}
func TestScheduleAcceptsMultipleWeekdaysAndTimezone(t *testing.T) {
	tm := "07:30"
	tz, e := parseInput(ScheduleInput{StartDate: "2026-08-19", Weekdays: []int16{1, 3, 5}, PreferredTime: &tm}, "Asia/Novosibirsk")
	if e != nil || tz != "Asia/Novosibirsk" {
		t.Fatalf("timezone=%q error=%v", tz, e)
	}
}
func TestPlannedValidation(t *testing.T) {
	if !errors.Is(validatePlanned(PlannedInput{ScheduledDate: "not-a-date"}), ErrInvalid) {
		t.Fatal("invalid date accepted")
	}
	if e := validatePlanned(PlannedInput{ScheduledDate: "2026-08-19", ScheduledTime: ptr("19:00")}); e != nil {
		t.Fatal(e)
	}
}
func ptr[T any](x T) *T { return &x }
