package coach

import (
	"context"
	"os"
	"testing"

	"github.com/calisthenics-coach/calisthenics-mini-app/backend/internal/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestCoachUUIDAttackMatrixPostgres(t *testing.T) {
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
	const coachA = "99000000-0000-0000-0000-000000000001"
	const coachB = "99000000-0000-0000-0000-000000000002"
	const tenantA = "99000000-0000-0000-0000-000000000011"
	const tenantB = "99000000-0000-0000-0000-000000000012"
	const lessonB = "99000000-0000-0000-0000-000000000021"
	const exerciseA = "99000000-0000-0000-0000-000000000031"
	const exerciseB = "99000000-0000-0000-0000-000000000032"
	const workoutA = "99000000-0000-0000-0000-000000000041"
	const workoutB = "99000000-0000-0000-0000-000000000042"
	const programB = "99000000-0000-0000-0000-000000000051"
	const skillA = "99000000-0000-0000-0000-000000000061"
	const skillB = "99000000-0000-0000-0000-000000000062"
	const mediaB = "99000000-0000-0000-0000-000000000071"
	cleanup := func() {
		_, _ = pool.Exec(ctx, `DELETE FROM skill_requirements WHERE skill_id IN ($1,$2) OR required_skill_id IN ($1,$2)`, skillA, skillB)
		_, _ = pool.Exec(ctx, `DELETE FROM workout_exercises WHERE workout_id IN ($1,$2)`, workoutA, workoutB)
		_, _ = pool.Exec(ctx, `DELETE FROM lessons WHERE id=$1`, lessonB)
		_, _ = pool.Exec(ctx, `DELETE FROM workouts WHERE id IN ($1,$2)`, workoutA, workoutB)
		_, _ = pool.Exec(ctx, `DELETE FROM programs WHERE id=$1`, programB)
		_, _ = pool.Exec(ctx, `DELETE FROM skills WHERE id IN ($1,$2)`, skillA, skillB)
		_, _ = pool.Exec(ctx, `DELETE FROM exercises WHERE id IN ($1,$2)`, exerciseA, exerciseB)
		_, _ = pool.Exec(ctx, `DELETE FROM media_assets WHERE id=$1`, mediaB)
		_, _ = pool.Exec(ctx, `DELETE FROM tenant_memberships WHERE tenant_id IN ($1,$2)`, tenantA, tenantB)
		_, _ = pool.Exec(ctx, `DELETE FROM tenants WHERE id IN ($1,$2)`, tenantA, tenantB)
		_, _ = pool.Exec(ctx, `DELETE FROM profiles WHERE user_id IN ($1,$2)`, coachA, coachB)
		_, _ = pool.Exec(ctx, `DELETE FROM users WHERE id IN ($1,$2)`, coachA, coachB)
	}
	cleanup()
	defer cleanup()
	for index, id := range []string{coachA, coachB} {
		if _, err = pool.Exec(ctx, `INSERT INTO users(id,telegram_id,first_name) VALUES($1,$2,'Attack coach')`, id, 990000001+index); err != nil {
			t.Fatal(err)
		}
		if _, err = pool.Exec(ctx, `INSERT INTO profiles(user_id,display_name) VALUES($1,'Attack coach')`, id); err != nil {
			t.Fatal(err)
		}
	}
	if _, err = pool.Exec(ctx, `INSERT INTO tenants(id,slug,name,owner_user_id) VALUES($1,'attack-a','Attack A',$3),($2,'attack-b','Attack B',$4)`, tenantA, tenantB, coachA, coachB); err != nil {
		t.Fatal(err)
	}
	if _, err = pool.Exec(ctx, `INSERT INTO tenant_memberships(tenant_id,user_id,role) VALUES($1,$3,'coach'),($2,$4,'coach')`, tenantA, tenantB, coachA, coachB); err != nil {
		t.Fatal(err)
	}
	var category string
	if err = pool.QueryRow(ctx, `SELECT id::text FROM lesson_categories ORDER BY sort_order LIMIT 1`).Scan(&category); err != nil {
		t.Fatal(err)
	}
	if _, err = pool.Exec(ctx, `INSERT INTO media_assets(id,owner_user_id,tenant_id,type,status,storage_provider,storage_key,url,original_filename,mime_type,size_bytes) VALUES($1,$2,$3,'image','ready','fixture','attack-b-media','https://fixture/b.jpg','b.jpg','image/jpeg',100)`, mediaB, coachB, tenantB); err != nil {
		t.Fatal(err)
	}
	if _, err = pool.Exec(ctx, `INSERT INTO exercises(id,name,slug,description,instructions,common_mistakes,difficulty,muscle_groups,owner_user_id,tenant_id,status) VALUES($1,'Attack Exercise A','attack-exercise-a','A','A','A','beginner',ARRAY['core'],$3,$4,'published'),($2,'Unique B Exercise Search','attack-exercise-b','B','B','B','beginner',ARRAY['core'],$5,$6,'published')`, exerciseA, exerciseB, coachA, tenantA, coachB, tenantB); err != nil {
		t.Fatal(err)
	}
	if _, err = pool.Exec(ctx, `INSERT INTO lessons(id,category_id,title,slug,short_description,content,difficulty,duration_minutes,published,owner_user_id,tenant_id,status) VALUES($1,$2,'Unique B Lesson Search','attack-lesson-b','B','B','beginner',5,true,$3,$4,'published')`, lessonB, category, coachB, tenantB); err != nil {
		t.Fatal(err)
	}
	if _, err = pool.Exec(ctx, `INSERT INTO programs(id,name,slug,description,difficulty,duration_weeks,published,category,owner_user_id,tenant_id,status) VALUES($1,'Unique B Program Search','attack-program-b','B','beginner',2,true,'SKILL',$2,$3,'published')`, programB, coachB, tenantB); err != nil {
		t.Fatal(err)
	}
	if _, err = pool.Exec(ctx, `INSERT INTO workouts(id,title,description,estimated_minutes,owner_user_id,tenant_id,status,category,warmup_enabled) VALUES($1,'Attack Workout A','A',10,$3,$4,'published','strength',false),($2,'Unique B Workout Search','B',10,$5,$6,'published','strength',false)`, workoutA, workoutB, coachA, tenantA, coachB, tenantB); err != nil {
		t.Fatal(err)
	}
	if _, err = pool.Exec(ctx, `INSERT INTO workout_exercises(workout_id,exercise_id,sets,target_reps,rest_seconds) VALUES($1,$2,1,5,30)`, workoutB, exerciseB); err != nil {
		t.Fatal(err)
	}
	if _, err = pool.Exec(ctx, `INSERT INTO skills(id,code,name,description,category,difficulty,icon,xp_reward,final_criterion_type,final_criterion_value,owner_user_id,tenant_id,status) VALUES($1,'ATTACK_A','Attack Skill A','A','SKILL','beginner','A',1,'repetitions',1,$3,$4,'published'),($2,'ATTACK_B','Unique B Skill Search','B','SKILL','beginner','B',1,'repetitions',1,$5,$6,'published')`, skillA, skillB, coachA, tenantA, coachB, tenantB); err != nil {
		t.Fatal(err)
	}

	svc := NewService(pool)
	tenantCtx := middleware.WithTenant(ctx, tenantA, "coach")
	for _, kindID := range []struct{ kind, id string }{{"lessons", lessonB}, {"exercises", exerciseB}, {"workouts", workoutB}, {"programs", programB}, {"skills", skillB}} {
		kindID := kindID
		t.Run("CannotReadOtherTenant_"+kindID.kind, func(t *testing.T) {
			if _, e := svc.Get(tenantCtx, kindID.kind, kindID.id, coachA, "coach"); e == nil {
				t.Fatal("cross-tenant GET succeeded")
			}
		})
		for _, status := range []string{"draft", "published", "archived"} {
			status := status
			t.Run("CannotLifecycleOtherTenant_"+kindID.kind+"_"+status, func(t *testing.T) {
				if e := svc.Lifecycle(tenantCtx, kindID.kind, kindID.id, coachA, "coach", status); e == nil {
					t.Fatal("cross-tenant lifecycle succeeded")
				}
			})
		}
		t.Run("CannotDuplicateOtherTenant_"+kindID.kind, func(t *testing.T) {
			if _, e := svc.Duplicate(tenantCtx, kindID.kind, kindID.id, coachA, "coach"); e == nil {
				t.Fatal("cross-tenant duplicate succeeded")
			}
		})
	}
	t.Run("CannotUpdateOtherTenantLesson", func(t *testing.T) {
		if _, e := svc.SaveLesson(tenantCtx, coachA, "coach", ptrString(lessonB), LessonInput{CategoryID: category, Title: "attack", Slug: "attack-update-lesson", Content: "x", Difficulty: "beginner", DurationMinutes: 5, Blocks: []Block{{Type: "paragraph", Text: "x"}}}); e == nil {
			t.Fatal("cross-tenant lesson update succeeded")
		}
	})
	updates := []struct {
		name, kind, id string
		input          BuilderInput
	}{
		{"Exercise", "exercises", exerciseB, BuilderInput{}}, {"Workout", "workouts", workoutB, BuilderInput{}}, {"Program", "programs", programB, BuilderInput{}}, {"Skill", "skills", skillB, BuilderInput{}},
	}
	for _, update := range updates {
		update := update
		t.Run("CannotUpdateOtherTenant"+update.name, func(t *testing.T) {
			if _, e := svc.SaveBuilder(tenantCtx, update.kind, coachA, "coach", &update.id, update.input); e == nil {
				t.Fatal("cross-tenant builder update succeeded")
			}
		})
	}
	for _, kind := range []string{"lessons", "exercises", "workouts", "programs", "skills"} {
		items, e := svc.List(tenantCtx, kind, coachA, "coach", "Unique B", "")
		if e != nil {
			t.Fatal(e)
		}
		if len(items) != 0 {
			t.Fatalf("%s search leaked B: %+v", kind, items)
		}
	}
	options, err := svc.Options(tenantCtx, coachA, "coach")
	if err != nil {
		t.Fatal(err)
	}
	for section, items := range map[string][]Option{"exercises": options.Exercises, "workouts": options.Workouts, "programs": options.Programs, "skills": options.Skills, "media": options.Media} {
		for _, item := range items {
			if item.ID == exerciseB || item.ID == workoutB || item.ID == programB || item.ID == skillB || item.ID == mediaB {
				t.Fatalf("options %s leaked B id %s", section, item.ID)
			}
		}
	}
	mediaBefore, _ := svc.ListMedia(tenantCtx, coachA, "coach")
	for _, item := range mediaBefore {
		if item.ID == mediaB {
			t.Fatal("media list leaked B")
		}
	}
	if err = svc.DeleteMedia(tenantCtx, coachA, mediaB, "coach"); err == nil {
		t.Fatal("deleted B media")
	}
	var exists bool
	if err = pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM media_assets WHERE id=$1)`, mediaB).Scan(&exists); err != nil || !exists {
		t.Fatalf("B media mutated: %v %v", exists, err)
	}
	if _, err = pool.Exec(ctx, `INSERT INTO workout_exercises(workout_id,exercise_id,sets,target_reps,rest_seconds) VALUES($1,$2,1,5,30)`, workoutA, exerciseB); err == nil {
		t.Fatal("database accepted cross-tenant exercise reference")
	}
	if _, err = pool.Exec(ctx, `UPDATE exercises SET cover_media_id=$1 WHERE id=$2`, mediaB, exerciseA); err == nil {
		t.Fatal("database accepted cross-tenant media reference")
	}
	if _, err = pool.Exec(ctx, `INSERT INTO skill_requirements(skill_id,required_skill_id,requirement_type) VALUES($1,$2,'skill_mastered')`, skillA, skillB); err == nil {
		t.Fatal("database accepted cross-tenant skill prerequisite")
	}
	var statuses int
	if err = pool.QueryRow(ctx, `SELECT count(*) FROM (VALUES ($1::uuid,'published'),($2::uuid,'published'),($3::uuid,'published'),($4::uuid,'published'),($5::uuid,'published')) expected(id,status) JOIN (SELECT id,status FROM lessons UNION ALL SELECT id,status FROM exercises UNION ALL SELECT id,status FROM workouts UNION ALL SELECT id,status FROM programs UNION ALL SELECT id,status FROM skills) actual USING(id,status)`, lessonB, exerciseB, workoutB, programB, skillB).Scan(&statuses); err != nil || statuses != 5 {
		t.Fatalf("attack mutated B lifecycle state: %d %v", statuses, err)
	}
}

func ptrString(value string) *string { return &value }
