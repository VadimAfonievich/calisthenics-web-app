package coach

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

func TestLessonBlockValidation(t *testing.T) {
	id := "10000000-0000-0000-0000-000000000001"
	valid := []Block{{Type: "heading", Text: "Title"}, {Type: "image", MediaID: &id}, {Type: "checklist", Items: []string{"Ready"}}, {Type: "divider"}}
	if !validBlocks(valid) {
		t.Fatal("valid blocks rejected")
	}
	for _, blocks := range [][]Block{{{Type: "script", Text: "x"}}, {{Type: "image"}}, {{Type: "checklist"}}} {
		if validBlocks(blocks) {
			t.Fatalf("invalid blocks accepted: %#v", blocks)
		}
	}
}
func TestMediaValidationAndFilenameSafety(t *testing.T) {
	if !validMime("image", "image/webp") || validMime("image", "text/html") || !validMime("video", "video/mp4") {
		t.Fatal("mime validation failed")
	}
	if got := safeName("../../secret.png"); got != ".._.._secret.png" {
		t.Fatalf("unsafe name %q", got)
	}
}
func TestRoles(t *testing.T) {
	if !Role("admin").CanManageAll() || !Role("super_admin").CanManageAll() || Role("coach").CanManageAll() {
		t.Fatal("role hierarchy invalid")
	}
}

func TestWorkoutExercisesRejectDuplicates(t *testing.T) {
	id := "10000000-0000-0000-0000-000000000001"
	reps := 10
	items := []BuilderExercise{{ExerciseID: id, Sets: 3, TargetReps: &reps, RestSeconds: 60, SortOrder: 0}, {ExerciseID: id, Sets: 2, TargetReps: &reps, RestSeconds: 30, SortOrder: 1}}
	if validExercises(items) {
		t.Fatal("duplicate exercise must be rejected before database insert")
	}
}

func TestBuilderValidationMatchesDatabaseEnums(t *testing.T) {
	base := BuilderInput{Difficulty: "beginner", MovementType: "reps"}
	if !validBuilderEnums("exercises", base) {
		t.Fatal("valid exercise enums rejected")
	}
	base.Difficulty = "expert"
	if validBuilderEnums("exercises", base) {
		t.Fatal("difficulty outside the DB check must be rejected before SQL")
	}
	base.Difficulty, base.MovementType = "beginner", "strength"
	if validBuilderEnums("exercises", base) {
		t.Fatal("movement type outside the DB check must be rejected before SQL")
	}
}

func TestWorkoutCategoriesMatchPublicTaxonomy(t *testing.T) {
	for _, category := range []string{"warmup", "morning", "strength", "skill"} {
		if !validBuilderEnums("workouts", BuilderInput{Difficulty: "beginner", Category: category}) {
			t.Fatalf("public workout category %q rejected", category)
		}
	}
	for _, legacy := range []string{"WARMUP", "MORNING_ROUTINE", "BASE_STRENGTH", "mobility", "other"} {
		if validBuilderEnums("workouts", BuilderInput{Difficulty: "beginner", Category: legacy}) {
			t.Fatalf("legacy category %q accepted for a new write", legacy)
		}
	}
}

func TestWorkoutValidationReturnsSafeFieldInformation(t *testing.T) {
	in := BuilderInput{Title: "Сила", Description: "План", Difficulty: "beginner", Category: "strength", EstimatedMinutes: 30, DayNumber: 1, ProgramID: "10000000-0000-0000-0000-000000000001", Exercises: []BuilderExercise{{ExerciseID: "10000000-0000-0000-0000-000000000002", Sets: 3, RestSeconds: 60}}}
	err := validateWorkoutInput(in)
	var validation *ValidationError
	if !errors.As(err, &validation) || !strings.Contains(validation.Message, "повторения или длительность") {
		t.Fatalf("expected safe exercise field error, got %v", err)
	}
}

func TestBodyEntityIDsAreValidatedBeforeSQL(t *testing.T) {
	if validID("undefined") || validID("not-a-uuid") {
		t.Fatal("invalid relation IDs must not reach PostgreSQL casts")
	}
	if !validID("10000000-0000-0000-0000-000000000001") {
		t.Fatal("canonical UUID rejected")
	}
}

func TestWorkoutListUsesStandaloneDifficulty(t *testing.T) {
	workouts := tables["workouts"]
	if workouts.difficulty != "difficulty" || workouts.listFrom != "workouts" {
		t.Fatalf("workout list must include standalone rows: %#v", workouts)
	}
}

func TestStandaloneWorkoutValidation(t *testing.T) {
	reps := 10
	in := BuilderInput{Title: "Standalone", Description: "Plan", Difficulty: "beginner", Category: "strength", EstimatedMinutes: 20, Exercises: []BuilderExercise{{ExerciseID: "10000000-0000-0000-0000-000000000002", Sets: 3, TargetReps: &reps, RestSeconds: 60, SortOrder: 0}}}
	if err := validateWorkoutInput(in); err != nil {
		t.Fatalf("standalone workout rejected: %v", err)
	}
}

func TestCoachWorkoutListContractSerializesSystemAndOwnedRows(t *testing.T) {
	owner := "10000000-0000-0000-0000-000000000099"
	rows := []Item{{ID: "10000000-0000-0000-0000-000000000001", Name: "System", Status: "published", Difficulty: "beginner", OwnerUserID: nil}, {ID: "10000000-0000-0000-0000-000000000002", Name: "Owned", Status: "draft", Difficulty: "intermediate", OwnerUserID: &owner}}
	data, err := json.Marshal(rows)
	if err != nil || !strings.Contains(string(data), `"owner_user_id":null`) || !strings.Contains(string(data), `"difficulty":"beginner"`) {
		t.Fatalf("workout list DTO cannot represent system rows: %s %v", data, err)
	}
}

func TestProgramWorkoutRelationsValidateIDsOrderAndDuplicates(t *testing.T) {
	a := "10000000-0000-0000-0000-000000000001"
	b := "10000000-0000-0000-0000-000000000002"
	if !validProgramWorkouts([]ProgramWorkout{{WorkoutID: a, SortOrder: 0}, {WorkoutID: b, SortOrder: 1}}) {
		t.Fatal("valid ordered workouts rejected")
	}
	if validProgramWorkouts([]ProgramWorkout{{WorkoutID: a, SortOrder: 0}, {WorkoutID: a, SortOrder: 1}}) {
		t.Fatal("duplicate workout accepted")
	}
}
