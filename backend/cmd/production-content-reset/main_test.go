package main

import (
	"strings"
	"testing"
)

func TestDestructiveResetNeverDeletesIdentityTables(t *testing.T) {
	for _, target := range deleteTargets {
		name := strings.ToLower(target.Name)
		if name == "users" || name == "profiles" || name == "admin_users" || name == "achievements" || name == "levels" || name == "lesson_categories" {
			t.Fatalf("protected table scheduled for delete: %s", name)
		}
		if !strings.HasPrefix(strings.ToUpper(target.SQL), "DELETE FROM ") {
			t.Fatalf("unexpected statement: %s", target.SQL)
		}
	}
}

func TestHistoryPrecedesContentAndMediaIsLast(t *testing.T) {
	positions := map[string]int{}
	for i, x := range deleteTargets {
		positions[x.Name] = i
	}
	pairs := [][2]string{{"exercise_sets", "workout_sessions"}, {"workout_sessions", "workouts"}, {"user_lesson_progress", "lessons"}, {"workout_exercises", "workouts"}, {"workout_exercises", "exercises"}, {"skill_criteria", "skills"}, {"skill_levels", "skills"}, {"workouts", "programs"}, {"programs", "media_assets"}, {"exercises", "media_assets"}}
	for _, pair := range pairs {
		if positions[pair[0]] >= positions[pair[1]] {
			t.Fatalf("unsafe order: %s must precede %s", pair[0], pair[1])
		}
	}
}

func TestProgressResetUsesCleanDefaults(t *testing.T) {
	joined := ""
	for _, x := range resetTargets {
		joined += x.SQL + "\n"
	}
	for _, required := range []string{"level=1", "xp=0", "current_streak=0", "longest_streak=0", "total_workouts=0", "total_completed_exercises=0", "total_training_seconds=0"} {
		if !strings.Contains(joined, required) {
			t.Fatalf("missing reset %s", required)
		}
	}
}
