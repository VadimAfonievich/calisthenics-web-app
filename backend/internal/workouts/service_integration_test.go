package workouts

import (
	"context"
	"os"
	"testing"

	"github.com/calisthenics-coach/calisthenics-mini-app/backend/internal/coach"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestWarmupContinuationPostgres(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL is not set")
	}
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	const user = "91000000-0000-0000-0000-000000000001"
	const warmup = "59000000-0000-0000-0000-000000000001"
	const main = "54000000-0000-0000-0000-000000000001"
	for _, query := range []string{
		`INSERT INTO users(id,telegram_id,first_name) VALUES($1,910000001,'Flow Test') ON CONFLICT(id) DO NOTHING`,
		`INSERT INTO profiles(user_id,display_name) VALUES($1,'Flow Test') ON CONFLICT(user_id) DO NOTHING`,
		`INSERT INTO user_progress(user_id) VALUES($1) ON CONFLICT(user_id) DO NOTHING`,
	} {
		if _, err = pool.Exec(ctx, query, user); err != nil {
			t.Fatal(err)
		}
	}
	defer func() {
		for _, query := range []string{`DELETE FROM workout_sessions WHERE user_id=$1`, `DELETE FROM user_progress WHERE user_id=$1`, `DELETE FROM profiles WHERE user_id=$1`, `DELETE FROM users WHERE id=$1`} {
			_, _ = pool.Exec(ctx, query, user)
		}
	}()
	svc := &Service{pool: pool}
	flow, err := svc.Start(ctx, user, warmup, StartInput{FollowUpWorkoutID: ptr(main)})
	if err != nil {
		t.Fatal(err)
	}
	resumed, err := svc.GetSession(ctx, user, flow.ID)
	if err != nil || resumed.Session.FollowUpWorkoutID == nil || *resumed.Session.FollowUpWorkoutID != main {
		t.Fatalf("resume lost follow-up: %+v %v", resumed.Session, err)
	}
	first, err := svc.Complete(ctx, user, flow.ID, 60)
	if err != nil || first.NextSession == nil {
		t.Fatalf("first completion: %+v %v", first, err)
	}
	second, err := svc.Complete(ctx, user, flow.ID, 60)
	if err != nil || second.NextSession == nil || second.NextSession.ID != first.NextSession.ID {
		t.Fatalf("completion was not idempotent: first=%+v second=%+v err=%v", first.NextSession, second.NextSession, err)
	}
	standalone, err := svc.Start(ctx, user, warmup, StartInput{})
	if err != nil {
		t.Fatal(err)
	}
	standaloneDone, err := svc.Complete(ctx, user, standalone.ID, 30)
	if err != nil || standaloneDone.NextSession != nil {
		t.Fatalf("standalone warmup started continuation: %+v %v", standaloneDone, err)
	}
}

func ptr(value string) *string { return &value }

func TestCoachBuilderToStudentPlayerPostgres(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL is not set")
	}
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	const coachID = "92000000-0000-0000-0000-000000000001"
	const studentID = "92000000-0000-0000-0000-000000000002"
	for _, row := range []struct {
		id       string
		telegram int64
		name     string
	}{{coachID, 920000001, "Builder Coach"}, {studentID, 920000002, "Builder Student"}} {
		if _, err = pool.Exec(ctx, `INSERT INTO users(id,telegram_id,first_name) VALUES($1,$2,$3) ON CONFLICT(id) DO NOTHING`, row.id, row.telegram, row.name); err != nil {
			t.Fatal(err)
		}
		if _, err = pool.Exec(ctx, `INSERT INTO profiles(user_id,display_name) VALUES($1,$2) ON CONFLICT(user_id) DO NOTHING`, row.id, row.name); err != nil {
			t.Fatal(err)
		}
		if _, err = pool.Exec(ctx, `INSERT INTO user_progress(user_id) VALUES($1) ON CONFLICT(user_id) DO NOTHING`, row.id); err != nil {
			t.Fatal(err)
		}
	}
	if _, err = pool.Exec(ctx, `INSERT INTO admin_users(user_id,role) VALUES($1,'coach') ON CONFLICT(user_id) DO UPDATE SET role='coach'`, coachID); err != nil {
		t.Fatal(err)
	}
	var repsID string
	if err = pool.QueryRow(ctx, `SELECT id::text FROM exercises WHERE status='published' AND movement_type<>'duration' ORDER BY id LIMIT 1`).Scan(&repsID); err != nil {
		t.Fatal(err)
	}
	reps, seconds := 10, 30
	builder := coach.NewService(pool)
	durationID, err := builder.SaveBuilder(ctx, "exercises", coachID, coach.Role("coach"), nil, coach.BuilderInput{Name: "E2E Планка", Description: "Удержание корпуса", Difficulty: "beginner", MovementType: "duration", Instructions: "Держите корпус ровно", CommonMistakes: "Не прогибайтесь", CoachTips: "Дышите спокойно", MuscleGroups: []string{"core"}, Equipment: []string{"floor"}, Tags: []string{"core"}})
	if err != nil {
		t.Fatal(err)
	}
	if err = builder.Lifecycle(ctx, "exercises", durationID, coachID, coach.Role("coach"), "published"); err != nil {
		t.Fatal(err)
	}
	workoutID, err := builder.SaveBuilder(ctx, "workouts", coachID, coach.Role("coach"), nil, coach.BuilderInput{Title: "Тест конструктора", Description: "Coach to student acceptance", Difficulty: "beginner", Category: "strength", EstimatedMinutes: 12, Exercises: []coach.BuilderExercise{{ExerciseID: repsID, Sets: 3, TargetReps: &reps, RestSeconds: 45, Notes: ptr("Контроль техники"), SortOrder: 0}, {ExerciseID: durationID, Sets: 3, TargetDurationSeconds: &seconds, RestSeconds: 45, SortOrder: 1}}})
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		for _, q := range []string{`DELETE FROM exercise_sets WHERE session_id IN (SELECT id FROM workout_sessions WHERE user_id=$1)`, `DELETE FROM user_exercise_stats WHERE user_id=$1`, `DELETE FROM user_achievements WHERE user_id=$1`, `DELETE FROM workout_sessions WHERE user_id=$1`, `DELETE FROM workouts WHERE id=$2`, `DELETE FROM exercises WHERE id=$4`, `DELETE FROM admin_users WHERE user_id=$3`, `DELETE FROM user_progress WHERE user_id IN ($1,$3)`, `DELETE FROM profiles WHERE user_id IN ($1,$3)`, `DELETE FROM users WHERE id IN ($1,$3)`} {
			_, _ = pool.Exec(ctx, q, studentID, workoutID, coachID, durationID)
		}
	}()
	if err = builder.Lifecycle(ctx, "workouts", workoutID, coachID, coach.Role("coach"), "published"); err != nil {
		t.Fatal(err)
	}
	player := NewService(pool)
	catalog, err := player.List(ctx, studentID)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, item := range catalog {
		if item.ID == workoutID {
			found = true
		}
	}
	if !found {
		t.Fatal("published coach workout missing from student catalog")
	}
	detail, err := player.Get(ctx, workoutID)
	if err != nil || len(detail.Exercises) != 2 || detail.Exercises[0].Notes != "Контроль техники" || detail.Exercises[1].TargetDuration == nil {
		t.Fatalf("player contract mismatch: %+v %v", detail, err)
	}
	session, err := player.Start(ctx, studentID, workoutID, StartInput{})
	if err != nil {
		t.Fatal(err)
	}
	r := int16(10)
	d := int32(30)
	if err = player.RecordSet(ctx, studentID, session.ID, SetInput{ExerciseID: repsID, Number: 1, Reps: &r, Completed: true}); err != nil {
		t.Fatal(err)
	}
	if err = player.RecordSet(ctx, studentID, session.ID, SetInput{ExerciseID: durationID, Number: 1, Duration: &d, Completed: true}); err != nil {
		t.Fatal(err)
	}
	done, err := player.Complete(ctx, studentID, session.ID, 180)
	if err != nil || done.Status != "completed" {
		t.Fatalf("completion failed: %+v %v", done, err)
	}
}
