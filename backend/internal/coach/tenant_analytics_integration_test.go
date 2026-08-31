package coach

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/calisthenics-coach/calisthenics-mini-app/backend/internal/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestTenantAnalyticsTenAndFourPostgres(t *testing.T) {
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
	const tenantA = "96000000-0000-0000-0000-000000000001"
	const tenantB = "96000000-0000-0000-0000-000000000002"
	const coachA = "96000000-0000-0000-0000-000000000011"
	const coachB = "96000000-0000-0000-0000-000000000012"
	const workoutA = "96000000-0000-0000-0000-000000000021"
	const workoutB = "96000000-0000-0000-0000-000000000022"
	const lessonA = "96000000-0000-0000-0000-000000000031"
	const lessonB = "96000000-0000-0000-0000-000000000032"
	const skillA = "96000000-0000-0000-0000-000000000041"
	const skillB = "96000000-0000-0000-0000-000000000042"
	cleanup := func() {
		_, _ = pool.Exec(ctx, `DELETE FROM user_skill_progress WHERE tenant_id IN ($1,$2)`, tenantA, tenantB)
		_, _ = pool.Exec(ctx, `DELETE FROM user_lesson_progress WHERE tenant_id IN ($1,$2)`, tenantA, tenantB)
		_, _ = pool.Exec(ctx, `DELETE FROM workout_sessions WHERE tenant_id IN ($1,$2)`, tenantA, tenantB)
		_, _ = pool.Exec(ctx, `DELETE FROM skills WHERE id IN ($1,$2)`, skillA, skillB)
		_, _ = pool.Exec(ctx, `DELETE FROM lessons WHERE id IN ($1,$2)`, lessonA, lessonB)
		_, _ = pool.Exec(ctx, `DELETE FROM workouts WHERE id IN ($1,$2)`, workoutA, workoutB)
		_, _ = pool.Exec(ctx, `DELETE FROM tenant_memberships WHERE tenant_id IN ($1,$2)`, tenantA, tenantB)
		_, _ = pool.Exec(ctx, `DELETE FROM tenants WHERE id IN ($1,$2)`, tenantA, tenantB)
		_, _ = pool.Exec(ctx, `DELETE FROM profiles WHERE user_id IN (SELECT id FROM users WHERE telegram_id BETWEEN 960000000 AND 960000099)`)
		_, _ = pool.Exec(ctx, `DELETE FROM users WHERE telegram_id BETWEEN 960000000 AND 960000099`)
	}
	cleanup()
	defer cleanup()

	for index := 0; index < 14; index++ {
		id := fmt.Sprintf("96000000-0000-0000-0001-%012d", index+1)
		if _, err = pool.Exec(ctx, `INSERT INTO users(id,telegram_id,first_name) VALUES($1,$2,'Analytics student')`, id, 960000020+index); err != nil {
			t.Fatal(err)
		}
		if _, err = pool.Exec(ctx, `INSERT INTO profiles(user_id,display_name) VALUES($1,'Analytics student')`, id); err != nil {
			t.Fatal(err)
		}
	}
	for index, id := range []string{coachA, coachB} {
		if _, err = pool.Exec(ctx, `INSERT INTO users(id,telegram_id,first_name) VALUES($1,$2,'Analytics coach')`, id, 960000001+index); err != nil {
			t.Fatal(err)
		}
		if _, err = pool.Exec(ctx, `INSERT INTO profiles(user_id,display_name) VALUES($1,'Analytics coach')`, id); err != nil {
			t.Fatal(err)
		}
	}
	if _, err = pool.Exec(ctx, `INSERT INTO tenants(id,slug,name,owner_user_id) VALUES($1,'analytics-a','Analytics A',$3),($2,'analytics-b','Analytics B',$4)`, tenantA, tenantB, coachA, coachB); err != nil {
		t.Fatal(err)
	}
	if _, err = pool.Exec(ctx, `INSERT INTO tenant_memberships(tenant_id,user_id,role) VALUES($1,$3,'coach'),($2,$4,'coach')`, tenantA, tenantB, coachA, coachB); err != nil {
		t.Fatal(err)
	}
	for index := 0; index < 10; index++ {
		id := fmt.Sprintf("96000000-0000-0000-0001-%012d", index+1)
		if _, err = pool.Exec(ctx, `INSERT INTO tenant_memberships(tenant_id,user_id,role) VALUES($1,$2,'student')`, tenantA, id); err != nil {
			t.Fatal(err)
		}
	}
	// The first student belongs to both spaces but only trains in B.
	for _, index := range []int{0, 10, 11, 12} {
		id := fmt.Sprintf("96000000-0000-0000-0001-%012d", index+1)
		if _, err = pool.Exec(ctx, `INSERT INTO tenant_memberships(tenant_id,user_id,role) VALUES($1,$2,'student')`, tenantB, id); err != nil {
			t.Fatal(err)
		}
	}
	if _, err = pool.Exec(ctx, `INSERT INTO workouts(id,title,description,estimated_minutes,owner_user_id,tenant_id,status) VALUES($1,'Workout A','A',10,$3,$4,'published'),($2,'Workout B','B',10,$5,$6,'published')`, workoutA, workoutB, coachA, tenantA, coachB, tenantB); err != nil {
		t.Fatal(err)
	}
	student := func(index int) string { return fmt.Sprintf("96000000-0000-0000-0001-%012d", index+1) }
	for _, index := range []int{1, 2, 3} {
		if _, err = pool.Exec(ctx, `INSERT INTO workout_sessions(user_id,workout_id,status,completed_at,duration_seconds,tenant_id) VALUES($1,$2,'completed',NOW(),600,$3)`, student(index), workoutA, tenantA); err != nil {
			t.Fatal(err)
		}
	}
	for _, index := range []int{0, 10} {
		if _, err = pool.Exec(ctx, `INSERT INTO workout_sessions(user_id,workout_id,status,completed_at,duration_seconds,tenant_id) VALUES($1,$2,'completed',NOW(),900,$3)`, student(index), workoutB, tenantB); err != nil {
			t.Fatal(err)
		}
	}
	var category string
	if err = pool.QueryRow(ctx, `SELECT id::text FROM lesson_categories ORDER BY sort_order LIMIT 1`).Scan(&category); err != nil {
		t.Fatal(err)
	}
	if _, err = pool.Exec(ctx, `INSERT INTO lessons(id,category_id,title,slug,short_description,content,difficulty,duration_minutes,published,status,owner_user_id,tenant_id) VALUES($1,$3,'Lesson A','analytics-lesson-a','A','A','beginner',5,true,'published',$4,$5),($2,$3,'Lesson B','analytics-lesson-b','B','B','beginner',5,true,'published',$6,$7)`, lessonA, lessonB, category, coachA, tenantA, coachB, tenantB); err != nil {
		t.Fatal(err)
	}
	if _, err = pool.Exec(ctx, `INSERT INTO user_lesson_progress(user_id,lesson_id,completed,progress_percent,completed_at,tenant_id) VALUES($1,$2,true,100,NOW(),$3),($4,$5,true,100,NOW(),$6)`, student(1), lessonA, tenantA, student(0), lessonB, tenantB); err != nil {
		t.Fatal(err)
	}
	if _, err = pool.Exec(ctx, `INSERT INTO skills(id,code,name,description,category,difficulty,icon,xp_reward,final_criterion_type,final_criterion_value,owner_user_id,tenant_id,status) VALUES($1,'analytics-skill-a','Skill A','A','SKILL','beginner','A',1,'repetitions',1,$3,$4,'published'),($2,'analytics-skill-b','Skill B','B','SKILL','beginner','B',1,'repetitions',1,$5,$6,'published')`, skillA, skillB, coachA, tenantA, coachB, tenantB); err != nil {
		t.Fatal(err)
	}
	if _, err = pool.Exec(ctx, `INSERT INTO user_skill_progress(user_id,skill_id,current_level,status,started_at,tenant_id) VALUES($1,$2,2,'in_progress',NOW(),$3),($4,$5,3,'in_progress',NOW(),$6)`, student(1), skillA, tenantA, student(0), skillB, tenantB); err != nil {
		t.Fatal(err)
	}

	service := NewService(pool)
	a, err := service.Analytics(middleware.WithTenant(ctx, tenantA, "coach"), coachA)
	if err != nil {
		t.Fatal(err)
	}
	b, err := service.Analytics(middleware.WithTenant(ctx, tenantB, "coach"), coachB)
	if err != nil {
		t.Fatal(err)
	}
	if a.TotalUsers != 10 || b.TotalUsers != 4 || a.ActiveUsers7D != 3 || b.ActiveUsers7D != 2 || a.TotalWorkoutsCompleted != 3 || b.TotalWorkoutsCompleted != 2 || a.TotalLessonsCompleted != 1 || b.TotalLessonsCompleted != 1 {
		t.Fatalf("tenant analytics leaked: A=%+v B=%+v", a, b)
	}
	if len(a.PopularWorkouts) != 1 || a.PopularWorkouts[0].Name != "Workout A" || len(b.PopularWorkouts) != 1 || b.PopularWorkouts[0].Name != "Workout B" || len(a.SkillProgress) != 1 || a.SkillProgress[0].Name != "Skill A" || len(b.SkillProgress) != 1 || b.SkillProgress[0].Name != "Skill B" {
		t.Fatalf("metric detail leaked: A=%+v B=%+v", a, b)
	}
}
