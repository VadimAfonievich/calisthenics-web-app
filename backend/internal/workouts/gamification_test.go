package workouts

import (
	"testing"
	"time"
)

func TestWarmupDoesNotCountAsFullWorkout(t *testing.T) {
	if countsAsFullWorkout("warmup") || !countsAsFullWorkout("strength") {
		t.Fatal("warmup gamification classification is incorrect")
	}
}

func TestNextStreak(t *testing.T) {
	today := time.Date(2026, 8, 18, 0, 0, 0, 0, time.UTC)
	yesterday := today.AddDate(0, 0, -1)
	same := today
	old := today.AddDate(0, 0, -2)
	for _, tc := range []struct {
		name          string
		last          *time.Time
		current, want int32
	}{{"first", nil, 0, 1}, {"same day", &same, 4, 4}, {"next day", &yesterday, 4, 5}, {"missed", &old, 4, 1}} {
		t.Run(tc.name, func(t *testing.T) {
			if got := nextStreak(tc.last, tc.current, today); got != tc.want {
				t.Fatalf("got %d want %d", got, tc.want)
			}
		})
	}
}
