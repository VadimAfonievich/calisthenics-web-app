package coach

import "testing"

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

func TestBodyEntityIDsAreValidatedBeforeSQL(t *testing.T) {
	if validID("undefined") || validID("not-a-uuid") {
		t.Fatal("invalid relation IDs must not reach PostgreSQL casts")
	}
	if !validID("10000000-0000-0000-0000-000000000001") {
		t.Fatal("canonical UUID rejected")
	}
}
