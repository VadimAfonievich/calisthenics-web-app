package workouts

import (
	"context"
	"os"
	"testing"

	"github.com/calisthenics-coach/calisthenics-mini-app/backend/internal/coach"
	"github.com/calisthenics-coach/calisthenics-mini-app/backend/internal/middleware"
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
	const warmup = "91000000-0000-0000-0000-000000000011"
	const main = "91000000-0000-0000-0000-000000000012"
	const tenant = "91000000-0000-0000-0000-000000000013"
	for _, query := range []string{
		`INSERT INTO users(id,telegram_id,first_name) VALUES($1,910000001,'Flow Test') ON CONFLICT(id) DO NOTHING`,
		`INSERT INTO profiles(user_id,display_name) VALUES($1,'Flow Test') ON CONFLICT(user_id) DO NOTHING`,
		`INSERT INTO user_progress(user_id) VALUES($1) ON CONFLICT(user_id) DO NOTHING`,
	} {
		if _, err = pool.Exec(ctx, query, user); err != nil {
			t.Fatal(err)
		}
	}
	if _, err = pool.Exec(ctx, `INSERT INTO tenants(id,slug,name,owner_user_id) VALUES($1,'flow-e2e','Flow E2E',$2)`, tenant, user); err != nil {
		t.Fatal(err)
	}
	if _, err = pool.Exec(ctx, `INSERT INTO tenant_memberships(tenant_id,user_id,role) VALUES($1,$2,'coach')`, tenant, user); err != nil {
		t.Fatal(err)
	}
	if _, err = pool.Exec(ctx, `INSERT INTO workouts(id,title,description,estimated_minutes,category,warmup_enabled,owner_user_id,tenant_id,status) VALUES($3,'Warmup','test',5,'warmup',false,$2,$1,'published'),($4,'Main','test',10,'strength',true,$2,$1,'published')`, tenant, user, warmup, main); err != nil {
		t.Fatal(err)
	}
	if _, err = pool.Exec(ctx, `UPDATE workouts SET warmup_workout_id=$1 WHERE id=$2`, warmup, main); err != nil {
		t.Fatal(err)
	}
	ctx = middleware.WithTenant(ctx, tenant, "coach")
	defer func() {
		_, _ = pool.Exec(ctx, `DELETE FROM workout_sessions WHERE user_id=$1`, user)
		_, _ = pool.Exec(ctx, `DELETE FROM workouts WHERE id IN ($1,$2)`, warmup, main)
		_, _ = pool.Exec(ctx, `DELETE FROM tenant_memberships WHERE tenant_id=$1`, tenant)
		_, _ = pool.Exec(ctx, `DELETE FROM tenants WHERE id=$1`, tenant)
		_, _ = pool.Exec(ctx, `DELETE FROM user_progress WHERE user_id=$1`, user)
		_, _ = pool.Exec(ctx, `DELETE FROM profiles WHERE user_id=$1`, user)
		_, _ = pool.Exec(ctx, `DELETE FROM users WHERE id=$1`, user)
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
	const tenantID = "92000000-0000-0000-0000-000000000010"
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
	if _, err = pool.Exec(ctx, `INSERT INTO tenants(id,slug,name,owner_user_id) VALUES($1,'builder-e2e','Builder E2E',$2)`, tenantID, coachID); err != nil {
		t.Fatal(err)
	}
	if _, err = pool.Exec(ctx, `INSERT INTO tenant_memberships(tenant_id,user_id,role) VALUES($1,$2,'coach'),($1,$3,'student')`, tenantID, coachID, studentID); err != nil {
		t.Fatal(err)
	}
	ctx = middleware.WithTenant(ctx, tenantID, "coach")
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
		_, _ = pool.Exec(ctx, `DELETE FROM exercise_sets WHERE session_id IN (SELECT id FROM workout_sessions WHERE user_id=$1)`, studentID)
		_, _ = pool.Exec(ctx, `DELETE FROM user_exercise_stats WHERE user_id=$1`, studentID)
		_, _ = pool.Exec(ctx, `DELETE FROM user_achievements WHERE user_id=$1`, studentID)
		_, _ = pool.Exec(ctx, `DELETE FROM workout_sessions WHERE user_id=$1`, studentID)
		_, _ = pool.Exec(ctx, `DELETE FROM workout_exercises WHERE workout_id=$1`, workoutID)
		_, _ = pool.Exec(ctx, `DELETE FROM workouts WHERE id=$1`, workoutID)
		_, _ = pool.Exec(ctx, `DELETE FROM exercises WHERE id=$1`, durationID)
		_, _ = pool.Exec(ctx, `DELETE FROM tenant_memberships WHERE tenant_id=$1`, tenantID)
		_, _ = pool.Exec(ctx, `DELETE FROM tenants WHERE id=$1`, tenantID)
		_, _ = pool.Exec(ctx, `DELETE FROM user_progress WHERE user_id IN ($1,$2)`, studentID, coachID)
		_, _ = pool.Exec(ctx, `DELETE FROM profiles WHERE user_id IN ($1,$2)`, studentID, coachID)
		_, _ = pool.Exec(ctx, `DELETE FROM users WHERE id IN ($1,$2)`, studentID, coachID)
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

func TestSessionCannotCrossTenantBoundary(t *testing.T) {
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
	const user = "98000000-0000-0000-0000-000000000001"
	const ownerA = "98000000-0000-0000-0000-000000000002"
	const ownerB = "98000000-0000-0000-0000-000000000003"
	const tenantA = "98000000-0000-0000-0000-000000000011"
	const tenantB = "98000000-0000-0000-0000-000000000012"
	const workoutA = "98000000-0000-0000-0000-000000000021"
	const workoutB = "98000000-0000-0000-0000-000000000022"
	const exerciseB = "98000000-0000-0000-0000-000000000031"
	const sessionB = "98000000-0000-0000-0000-000000000041"
	cleanup := func() {
		_, _ = pool.Exec(ctx, `DELETE FROM exercise_sets WHERE session_id=$1`, sessionB)
		_, _ = pool.Exec(ctx, `DELETE FROM workout_sessions WHERE id=$1`, sessionB)
		_, _ = pool.Exec(ctx, `DELETE FROM workout_exercises WHERE workout_id IN ($1,$2)`, workoutA, workoutB)
		_, _ = pool.Exec(ctx, `DELETE FROM workouts WHERE id IN ($1,$2)`, workoutA, workoutB)
		_, _ = pool.Exec(ctx, `DELETE FROM exercises WHERE id=$1`, exerciseB)
		_, _ = pool.Exec(ctx, `DELETE FROM tenant_memberships WHERE tenant_id IN ($1,$2)`, tenantA, tenantB)
		_, _ = pool.Exec(ctx, `DELETE FROM tenants WHERE id IN ($1,$2)`, tenantA, tenantB)
		_, _ = pool.Exec(ctx, `DELETE FROM user_progress WHERE user_id IN ($1,$2,$3)`, user, ownerA, ownerB)
		_, _ = pool.Exec(ctx, `DELETE FROM profiles WHERE user_id IN ($1,$2,$3)`, user, ownerA, ownerB)
		_, _ = pool.Exec(ctx, `DELETE FROM users WHERE id IN ($1,$2,$3)`, user, ownerA, ownerB)
	}
	cleanup()
	defer cleanup()
	for index, id := range []string{user, ownerA, ownerB} {
		if _, err = pool.Exec(ctx, `INSERT INTO users(id,telegram_id,first_name) VALUES($1,$2,'Boundary')`, id, 980000001+index); err != nil {
			t.Fatal(err)
		}
		if _, err = pool.Exec(ctx, `INSERT INTO profiles(user_id,display_name) VALUES($1,'Boundary')`, id); err != nil {
			t.Fatal(err)
		}
		if _, err = pool.Exec(ctx, `INSERT INTO user_progress(user_id) VALUES($1)`, id); err != nil {
			t.Fatal(err)
		}
	}
	if _, err = pool.Exec(ctx, `INSERT INTO tenants(id,slug,name,owner_user_id) VALUES($1,'session-a','Session A',$3),($2,'session-b','Session B',$4)`, tenantA, tenantB, ownerA, ownerB); err != nil {
		t.Fatal(err)
	}
	if _, err = pool.Exec(ctx, `INSERT INTO tenant_memberships(tenant_id,user_id,role) VALUES($1,$3,'student'),($2,$3,'student'),($1,$4,'coach'),($2,$5,'coach')`, tenantA, tenantB, user, ownerA, ownerB); err != nil {
		t.Fatal(err)
	}
	if _, err = pool.Exec(ctx, `INSERT INTO exercises(id,name,slug,description,instructions,common_mistakes,difficulty,muscle_groups,owner_user_id,tenant_id,status) VALUES($1,'Session B Exercise','session-b-exercise','B','B','B','beginner',ARRAY['core'],$2,$3,'published')`, exerciseB, ownerB, tenantB); err != nil {
		t.Fatal(err)
	}
	if _, err = pool.Exec(ctx, `INSERT INTO workouts(id,title,description,estimated_minutes,owner_user_id,tenant_id,status,category,warmup_enabled) VALUES($1,'Session A Workout','A',10,$3,$4,'published','strength',false),($2,'Session B Workout','B',10,$5,$6,'published','strength',false)`, workoutA, workoutB, ownerA, tenantA, ownerB, tenantB); err != nil {
		t.Fatal(err)
	}
	if _, err = pool.Exec(ctx, `INSERT INTO workout_exercises(workout_id,exercise_id,sets,target_reps,rest_seconds) VALUES($1,$2,1,5,30)`, workoutB, exerciseB); err != nil {
		t.Fatal(err)
	}
	if _, err = pool.Exec(ctx, `INSERT INTO workout_sessions(id,user_id,workout_id,tenant_id) VALUES($1,$2,$3,$4)`, sessionB, user, workoutB, tenantB); err != nil {
		t.Fatal(err)
	}
	svc := NewService(pool)
	ctxA := middleware.WithTenant(ctx, tenantA, "student")
	if _, err = svc.Get(ctxA, workoutB); err == nil {
		t.Fatal("Tenant A read Tenant B workout")
	}
	if _, err = svc.Start(ctxA, user, workoutB, StartInput{}); err == nil {
		t.Fatal("Tenant A started Tenant B workout")
	}
	if _, err = svc.GetSession(ctxA, user, sessionB); err == nil {
		t.Fatal("Tenant A resumed Tenant B session")
	}
	reps := int16(5)
	if err = svc.RecordSet(ctxA, user, sessionB, SetInput{ExerciseID: exerciseB, Number: 1, Reps: &reps, Completed: true}); err == nil {
		t.Fatal("Tenant A wrote set to Tenant B session")
	}
	if _, err = svc.Complete(ctxA, user, sessionB, 60); err == nil {
		t.Fatal("Tenant A completed Tenant B session")
	}
	var status string
	var sets int
	if err = pool.QueryRow(ctx, `SELECT status FROM workout_sessions WHERE id=$1`, sessionB).Scan(&status); err != nil {
		t.Fatal(err)
	}
	if err = pool.QueryRow(ctx, `SELECT count(*) FROM exercise_sets WHERE session_id=$1`, sessionB).Scan(&sets); err != nil {
		t.Fatal(err)
	}
	if status != "started" || sets != 0 {
		t.Fatalf("cross-tenant mutation occurred: status=%s sets=%d", status, sets)
	}
	ctxB := middleware.WithTenant(ctx, tenantB, "student")
	if _, err = svc.GetSession(ctxB, user, sessionB); err != nil {
		t.Fatalf("current Tenant B must access B session: %v", err)
	}
	if _, err = svc.Get(ctxB, workoutA); err == nil {
		t.Fatal("after switch to B, Tenant A workout remained accessible")
	}
}
