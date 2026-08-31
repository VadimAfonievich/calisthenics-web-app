package handler_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/calisthenics-coach/calisthenics-mini-app/backend/internal/calendar"
	"github.com/calisthenics-coach/calisthenics-mini-app/backend/internal/lessons"
	"github.com/calisthenics-coach/calisthenics-mini-app/backend/internal/middleware"
	"github.com/calisthenics-coach/calisthenics-mini-app/backend/internal/programs"
	"github.com/calisthenics-coach/calisthenics-mini-app/backend/internal/skills"
	"github.com/calisthenics-coach/calisthenics-mini-app/backend/internal/workouts"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestStudentUUIDAttackMatrixPostgres(t *testing.T) {
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
	const ownerA = "a9000000-0000-0000-0000-000000000001"
	const ownerB = "a9000000-0000-0000-0000-000000000002"
	const studentA = "a9000000-0000-0000-0000-000000000003"
	const studentB = "a9000000-0000-0000-0000-000000000004"
	const userX = "a9000000-0000-0000-0000-000000000005"
	const tenantA = "a9000000-0000-0000-0000-000000000011"
	const tenantB = "a9000000-0000-0000-0000-000000000012"
	const lessonA = "a9000000-0000-0000-0000-000000000021"
	const lessonB = "a9000000-0000-0000-0000-000000000022"
	const exerciseA = "a9000000-0000-0000-0000-000000000031"
	const exerciseB = "a9000000-0000-0000-0000-000000000032"
	const programA = "a9000000-0000-0000-0000-000000000041"
	const programB = "a9000000-0000-0000-0000-000000000042"
	const levelA = "a9000000-0000-0000-0000-000000000051"
	const levelB = "a9000000-0000-0000-0000-000000000052"
	const workoutA = "a9000000-0000-0000-0000-000000000061"
	const workoutB = "a9000000-0000-0000-0000-000000000062"
	const skillA = "a9000000-0000-0000-0000-000000000071"
	const skillB = "a9000000-0000-0000-0000-000000000072"
	const skillLevelA = "a9000000-0000-0000-0000-000000000081"
	const skillLevelB = "a9000000-0000-0000-0000-000000000082"
	const criterionA = "a9000000-0000-0000-0000-000000000091"
	const criterionB = "a9000000-0000-0000-0000-000000000092"
	const sessionB = "a9000000-0000-0000-0000-0000000000a1"
	const plannedA = "a9000000-0000-0000-0000-0000000000b1"
	const plannedB = "a9000000-0000-0000-0000-0000000000b2"
	const scheduleB = "a9000000-0000-0000-0000-0000000000c1"
	const mediaA = "a9000000-0000-0000-0000-0000000000d1"
	const mediaB = "a9000000-0000-0000-0000-0000000000d2"
	cleanup := func() {
		_, _ = pool.Exec(ctx, `DELETE FROM exercise_sets WHERE session_id=$1`, sessionB)
		_, _ = pool.Exec(ctx, `DELETE FROM workout_sessions WHERE id=$1`, sessionB)
		_, _ = pool.Exec(ctx, `DELETE FROM user_skill_criteria WHERE user_id IN ($1,$2)`, studentA, userX)
		_, _ = pool.Exec(ctx, `DELETE FROM user_skill_level_progress WHERE user_id IN ($1,$2)`, studentA, userX)
		_, _ = pool.Exec(ctx, `DELETE FROM user_skill_progress WHERE user_id IN ($1,$2)`, studentA, userX)
		_, _ = pool.Exec(ctx, `DELETE FROM user_program_progress WHERE user_id IN ($1,$2)`, studentA, userX)
		_, _ = pool.Exec(ctx, `DELETE FROM user_lesson_progress WHERE user_id IN ($1,$2)`, studentA, userX)
		_, _ = pool.Exec(ctx, `DELETE FROM user_planned_workouts WHERE id IN ($1,$2)`, plannedA, plannedB)
		_, _ = pool.Exec(ctx, `DELETE FROM user_training_schedule_days WHERE schedule_id=$1`, scheduleB)
		_, _ = pool.Exec(ctx, `DELETE FROM user_training_schedules WHERE id=$1`, scheduleB)
		_, _ = pool.Exec(ctx, `DELETE FROM skill_criteria WHERE id IN ($1,$2)`, criterionA, criterionB)
		_, _ = pool.Exec(ctx, `DELETE FROM skill_levels WHERE id IN ($1,$2)`, skillLevelA, skillLevelB)
		_, _ = pool.Exec(ctx, `DELETE FROM skills WHERE id IN ($1,$2)`, skillA, skillB)
		_, _ = pool.Exec(ctx, `DELETE FROM workout_exercises WHERE workout_id IN ($1,$2)`, workoutA, workoutB)
		_, _ = pool.Exec(ctx, `DELETE FROM workouts WHERE id IN ($1,$2)`, workoutA, workoutB)
		_, _ = pool.Exec(ctx, `DELETE FROM program_levels WHERE id IN ($1,$2)`, levelA, levelB)
		_, _ = pool.Exec(ctx, `DELETE FROM programs WHERE id IN ($1,$2)`, programA, programB)
		_, _ = pool.Exec(ctx, `DELETE FROM lessons WHERE id IN ($1,$2)`, lessonA, lessonB)
		_, _ = pool.Exec(ctx, `DELETE FROM exercises WHERE id IN ($1,$2)`, exerciseA, exerciseB)
		_, _ = pool.Exec(ctx, `DELETE FROM media_assets WHERE id IN ($1,$2)`, mediaA, mediaB)
		_, _ = pool.Exec(ctx, `DELETE FROM tenant_memberships WHERE tenant_id IN ($1,$2)`, tenantA, tenantB)
		_, _ = pool.Exec(ctx, `DELETE FROM tenants WHERE id IN ($1,$2)`, tenantA, tenantB)
		_, _ = pool.Exec(ctx, `DELETE FROM user_achievements WHERE user_id IN ($1,$2,$3)`, studentA, studentB, userX)
		_, _ = pool.Exec(ctx, `DELETE FROM user_progress WHERE user_id IN ($1,$2,$3,$4,$5)`, ownerA, ownerB, studentA, studentB, userX)
		_, _ = pool.Exec(ctx, `DELETE FROM profiles WHERE user_id IN ($1,$2,$3,$4,$5)`, ownerA, ownerB, studentA, studentB, userX)
		_, _ = pool.Exec(ctx, `DELETE FROM users WHERE id IN ($1,$2,$3,$4,$5)`, ownerA, ownerB, studentA, studentB, userX)
	}
	cleanup()
	defer cleanup()
	for i, id := range []string{ownerA, ownerB, studentA, studentB, userX} {
		if _, err = pool.Exec(ctx, `INSERT INTO users(id,telegram_id,first_name) VALUES($1,$2,'Student security')`, id, 990100001+i); err != nil {
			t.Fatal(err)
		}
		if _, err = pool.Exec(ctx, `INSERT INTO profiles(user_id,display_name) VALUES($1,'Student security')`, id); err != nil {
			t.Fatal(err)
		}
		if _, err = pool.Exec(ctx, `INSERT INTO user_progress(user_id) VALUES($1)`, id); err != nil {
			t.Fatal(err)
		}
	}
	if _, err = pool.Exec(ctx, `INSERT INTO tenants(id,slug,name,owner_user_id) VALUES($1,'student-sec-a','Student Security A',$3),($2,'student-sec-b','Student Security B',$4)`, tenantA, tenantB, ownerA, ownerB); err != nil {
		t.Fatal(err)
	}
	if _, err = pool.Exec(ctx, `INSERT INTO tenant_memberships(tenant_id,user_id,role) VALUES($1,$3,'coach'),($2,$4,'coach'),($1,$5,'student'),($2,$6,'student'),($1,$7,'student'),($2,$7,'student')`, tenantA, tenantB, ownerA, ownerB, studentA, studentB, userX); err != nil {
		t.Fatal(err)
	}
	if _, err = pool.Exec(ctx, `INSERT INTO media_assets(id,owner_user_id,tenant_id,type,status,storage_provider,storage_key,url,original_filename,mime_type,size_bytes) VALUES($1,$3,$4,'image','ready','fixture','student-sec-a-media','https://fixture/a.jpg','a.jpg','image/jpeg',10),($2,$5,$6,'image','ready','fixture','student-sec-b-media','https://fixture/b.jpg','b.jpg','image/jpeg',10)`, mediaA, mediaB, ownerA, tenantA, ownerB, tenantB); err != nil {
		t.Fatal(err)
	}
	if _, err = pool.Exec(ctx, `INSERT INTO exercises(id,name,slug,description,instructions,common_mistakes,difficulty,muscle_groups,owner_user_id,tenant_id,status) VALUES($1,'Student Sec A Exercise','student-sec-a-exercise','A','A','A','beginner',ARRAY['core'],$3,$4,'published'),($2,'Student Sec B Exercise','student-sec-b-exercise','B','B','B','beginner',ARRAY['core'],$5,$6,'published')`, exerciseA, exerciseB, ownerA, tenantA, ownerB, tenantB); err != nil {
		t.Fatal(err)
	}
	var category string
	if err = pool.QueryRow(ctx, `SELECT id::text FROM lesson_categories ORDER BY sort_order LIMIT 1`).Scan(&category); err != nil {
		t.Fatal(err)
	}
	if _, err = pool.Exec(ctx, `INSERT INTO lessons(id,category_id,title,slug,short_description,content,difficulty,duration_minutes,published,status,owner_user_id,tenant_id,cover_media_id) VALUES($1,$3,'Student Sec A Lesson','student-sec-a-lesson','A','A','beginner',5,true,'published',$4,$5,$6),($2,$3,'Student Sec B Lesson','student-sec-b-lesson','B','B','beginner',5,true,'published',$7,$8,$9)`, lessonA, lessonB, category, ownerA, tenantA, mediaA, ownerB, tenantB, mediaB); err != nil {
		t.Fatal(err)
	}
	if _, err = pool.Exec(ctx, `INSERT INTO programs(id,name,slug,description,difficulty,duration_weeks,published,status,category,owner_user_id,tenant_id) VALUES($1,'Student Sec A Program','student-sec-a-program','A','beginner',2,true,'published','SKILL',$3,$4),($2,'Student Sec B Program','student-sec-b-program','B','beginner',2,true,'published','SKILL',$5,$6)`, programA, programB, ownerA, tenantA, ownerB, tenantB); err != nil {
		t.Fatal(err)
	}
	if _, err = pool.Exec(ctx, `INSERT INTO program_levels(id,program_id,level_number,title,description,difficulty,unlock_rule_type,sort_order) VALUES($1,$3,1,'A','A','beginner','none',1),($2,$4,1,'B','B','beginner','none',1)`, levelA, levelB, programA, programB); err != nil {
		t.Fatal(err)
	}
	if _, err = pool.Exec(ctx, `INSERT INTO workouts(id,program_id,program_level_id,day_number,title,description,estimated_minutes,owner_user_id,tenant_id,status,category,warmup_enabled) VALUES($1,$3,$4,1,'Student Sec A Workout','A',10,$5,$6,'published','strength',false),($2,$7,$8,1,'Student Sec B Workout','B',10,$9,$10,'published','strength',false)`, workoutA, workoutB, programA, levelA, ownerA, tenantA, programB, levelB, ownerB, tenantB); err != nil {
		t.Fatal(err)
	}
	if _, err = pool.Exec(ctx, `INSERT INTO workout_exercises(workout_id,exercise_id,sets,target_reps,rest_seconds) VALUES($1,$2,1,5,30),($3,$4,1,5,30)`, workoutA, exerciseA, workoutB, exerciseB); err != nil {
		t.Fatal(err)
	}
	if _, err = pool.Exec(ctx, `INSERT INTO skills(id,code,name,description,category,difficulty,icon,xp_reward,final_criterion_type,final_criterion_value,owner_user_id,tenant_id,status) VALUES($1,'STUDENT_SEC_A','Student Sec A Skill','A','SKILL','beginner','A',10,'repetitions',1,$3,$4,'published'),($2,'STUDENT_SEC_B','Student Sec B Skill','B','SKILL','beginner','B',10,'repetitions',1,$5,$6,'published')`, skillA, skillB, ownerA, tenantA, ownerB, tenantB); err != nil {
		t.Fatal(err)
	}
	if _, err = pool.Exec(ctx, `INSERT INTO skill_levels(id,skill_id,level_number,name,description,criterion_type,criterion_value,sort_order) VALUES($1,$3,1,'A','A','repetitions',1,1),($2,$4,1,'B','B','repetitions',1,1)`, skillLevelA, skillLevelB, skillA, skillB); err != nil {
		t.Fatal(err)
	}
	if _, err = pool.Exec(ctx, `INSERT INTO skill_criteria(id,skill_id,code,title,criterion_type,target_value,sort_order) VALUES($1,$3,'SEC_A','A','manual_confirmation',1,1),($2,$4,'SEC_B','B','manual_confirmation',1,1)`, criterionA, criterionB, skillA, skillB); err != nil {
		t.Fatal(err)
	}
	if _, err = pool.Exec(ctx, `INSERT INTO workout_sessions(id,user_id,workout_id,tenant_id) VALUES($1,$2,$3,$4)`, sessionB, userX, workoutB, tenantB); err != nil {
		t.Fatal(err)
	}
	if _, err = pool.Exec(ctx, `INSERT INTO user_planned_workouts(id,user_id,workout_id,tenant_id,scheduled_date,timezone) VALUES($1,$2,$3,$4,CURRENT_DATE,'UTC'),($5,$2,$6,$7,CURRENT_DATE,'UTC')`, plannedA, userX, workoutA, tenantA, plannedB, workoutB, tenantB); err != nil {
		t.Fatal(err)
	}
	if _, err = pool.Exec(ctx, `INSERT INTO user_training_schedules(id,user_id,workout_id,tenant_id,timezone,start_date) VALUES($1,$2,$3,$4,'UTC',CURRENT_DATE)`, scheduleB, userX, workoutB, tenantB); err != nil {
		t.Fatal(err)
	}

	lessonsSvc := lessons.NewService(pool)
	workoutsSvc := workouts.NewService(pool)
	programsSvc := programs.NewService(pool)
	skillsSvc := skills.NewService(pool)
	calendarSvc := calendar.NewService(pool)
	ctxA := middleware.WithTenant(ctx, tenantA, "student")
	t.Run("StudentCannotAccessOtherTenantLesson", func(t *testing.T) {
		items, e := lessonsSvc.List(ctxA, studentA)
		if e != nil {
			t.Fatal(e)
		}
		for _, x := range items {
			if x.ID == lessonB {
				t.Fatal("lesson list leaked B")
			}
		}
		if _, e = lessonsSvc.Get(ctxA, studentA, lessonB); e == nil {
			t.Fatal("read B lesson")
		}
		if _, e = lessonsSvc.Complete(ctxA, studentA, lessonB); e == nil {
			t.Fatal("completed B lesson")
		}
	})
	t.Run("StudentCannotAccessOtherTenantWorkoutOrSession", func(t *testing.T) {
		items, e := workoutsSvc.List(ctxA, studentA)
		if e != nil {
			t.Fatal(e)
		}
		for _, x := range items {
			if x.ID == workoutB {
				t.Fatal("workout list leaked B")
			}
		}
		if _, e = workoutsSvc.Get(ctxA, workoutB); e == nil {
			t.Fatal("read B workout")
		}
		if _, e = workoutsSvc.Start(ctxA, studentA, workoutB, workouts.StartInput{}); e == nil {
			t.Fatal("started B workout")
		}
		if _, e = workoutsSvc.GetSession(ctxA, userX, sessionB); e == nil {
			t.Fatal("resumed B session")
		}
		reps := int16(5)
		if e = workoutsSvc.RecordSet(ctxA, userX, sessionB, workouts.SetInput{ExerciseID: exerciseB, Number: 1, Reps: &reps, Completed: true}); e == nil {
			t.Fatal("saved B set")
		}
		if _, e = workoutsSvc.Complete(ctxA, userX, sessionB, 60); e == nil {
			t.Fatal("finished B session")
		}
	})
	t.Run("StudentCannotAccessOtherTenantProgram", func(t *testing.T) {
		items, e := programsSvc.List(ctxA, studentA)
		if e != nil {
			t.Fatal(e)
		}
		for _, x := range items {
			if x.ID == programB {
				t.Fatal("program list leaked B")
			}
		}
		if _, e = programsSvc.Get(ctxA, studentA, programB); e == nil {
			t.Fatal("read B program")
		}
		if _, e = programsSvc.Start(ctxA, studentA, programB); e == nil {
			t.Fatal("started B program")
		}
	})
	t.Run("StudentCannotAccessOtherTenantSkill", func(t *testing.T) {
		items, e := skillsSvc.List(ctxA, studentA)
		if e != nil {
			t.Fatal(e)
		}
		for _, x := range items {
			if x.ID == skillB {
				t.Fatal("skill list leaked B")
			}
		}
		m, e := skillsSvc.Map(ctxA, studentA)
		if e != nil {
			t.Fatal(e)
		}
		for _, x := range m.Nodes {
			if x.ID == skillB {
				t.Fatal("skill map leaked B")
			}
		}
		if _, e = skillsSvc.Get(ctxA, studentA, skillB); e == nil {
			t.Fatal("read B skill")
		}
		if _, e = skillsSvc.ConfirmCriterion(ctxA, studentA, skillB, criterionB); e == nil {
			t.Fatal("confirmed B criterion")
		}
		if e = skillsSvc.CompleteLevel(ctxA, studentA, skillB, 1, 1); e == nil {
			t.Fatal("completed B skill level")
		}
		if _, e = skillsSvc.Master(ctxA, studentA, skillB, 1); e == nil {
			t.Fatal("mastered B skill")
		}
	})
	t.Run("StudentCannotAccessOtherTenantCalendar", func(t *testing.T) {
		from := time.Now().AddDate(0, 0, -1).Format("2006-01-02")
		to := time.Now().AddDate(0, 0, 1).Format("2006-01-02")
		items, e := calendarSvc.Calendar(ctxA, studentA, from, to)
		if e != nil {
			t.Fatal(e)
		}
		for _, x := range items {
			if x.WorkoutID == workoutB {
				t.Fatal("calendar leaked B")
			}
		}
		if _, e = calendarSvc.GetPlanned(ctxA, userX, plannedB); e == nil {
			t.Fatal("read B planned workout")
		}
		if e = calendarSvc.SkipPlanned(ctxA, userX, plannedB); e == nil {
			t.Fatal("mutated B planned workout")
		}
		schedules, e := calendarSvc.ListSchedules(ctxA, studentA)
		if e != nil {
			t.Fatal(e)
		}
		for _, x := range schedules {
			if x.ID == scheduleB {
				t.Fatal("schedule leaked B")
			}
		}
	})

	var xp, memberships, achievements, lessonProgress, programProgress, skillProgress, sessions, sets int
	if err = pool.QueryRow(ctx, `SELECT xp FROM profiles WHERE user_id=$1`, studentA).Scan(&xp); err != nil {
		t.Fatal(err)
	}
	if err = pool.QueryRow(ctx, `SELECT count(*) FROM tenant_memberships WHERE user_id=$1`, studentA).Scan(&memberships); err != nil {
		t.Fatal(err)
	}
	if err = pool.QueryRow(ctx, `SELECT count(*) FROM user_achievements WHERE user_id=$1`, studentA).Scan(&achievements); err != nil {
		t.Fatal(err)
	}
	if err = pool.QueryRow(ctx, `SELECT (SELECT count(*) FROM user_lesson_progress WHERE user_id=$1),(SELECT count(*) FROM user_program_progress WHERE user_id=$1),(SELECT count(*) FROM user_skill_progress WHERE user_id=$1),(SELECT count(*) FROM workout_sessions WHERE user_id=$1),(SELECT count(*) FROM exercise_sets es JOIN workout_sessions ws ON ws.id=es.session_id WHERE ws.user_id=$1)`, studentA).Scan(&lessonProgress, &programProgress, &skillProgress, &sessions, &sets); err != nil {
		t.Fatal(err)
	}
	if xp != 0 || memberships != 1 || achievements != 0 || lessonProgress != 0 || programProgress != 0 || skillProgress != 0 || sessions != 0 || sets != 0 {
		t.Fatalf("blocked attacks had side effects: xp=%d memberships=%d achievements=%d lesson=%d program=%d skill=%d sessions=%d sets=%d", xp, memberships, achievements, lessonProgress, programProgress, skillProgress, sessions, sets)
	}
	var sessionStatus, plannedStatus string
	if err = pool.QueryRow(ctx, `SELECT status FROM workout_sessions WHERE id=$1`, sessionB).Scan(&sessionStatus); err != nil {
		t.Fatal(err)
	}
	if err = pool.QueryRow(ctx, `SELECT status FROM user_planned_workouts WHERE id=$1`, plannedB).Scan(&plannedStatus); err != nil {
		t.Fatal(err)
	}
	if sessionStatus != "started" || plannedStatus != "scheduled" {
		t.Fatalf("foreign records mutated: session=%s planned=%s", sessionStatus, plannedStatus)
	}

	ctxB := middleware.WithTenant(ctx, tenantB, "student")
	if _, err = lessonsSvc.Get(ctxB, userX, lessonB); err != nil {
		t.Fatalf("B lesson unavailable in B: %v", err)
	}
	if _, err = workoutsSvc.Get(ctxB, workoutB); err != nil {
		t.Fatalf("B workout unavailable in B: %v", err)
	}
	if _, err = programsSvc.Get(ctxB, userX, programB); err != nil {
		t.Fatalf("B program unavailable in B: %v", err)
	}
	if _, err = skillsSvc.Get(ctxB, userX, skillB); err != nil {
		t.Fatalf("B skill unavailable in B: %v", err)
	}
	if _, err = calendarSvc.GetPlanned(ctxB, userX, plannedB); err != nil {
		t.Fatalf("B calendar unavailable in B: %v", err)
	}
	if _, err = lessonsSvc.Get(ctxB, userX, lessonA); err == nil {
		t.Fatal("A lesson accessible after switch to B")
	}
	if _, err = workoutsSvc.Get(ctxB, workoutA); err == nil {
		t.Fatal("A workout accessible after switch to B")
	}
	if _, err = programsSvc.Get(ctxB, userX, programA); err == nil {
		t.Fatal("A program accessible after switch to B")
	}
	if _, err = skillsSvc.Get(ctxB, userX, skillA); err == nil {
		t.Fatal("A skill accessible after switch to B")
	}
	if _, err = calendarSvc.GetPlanned(ctxB, userX, plannedA); err == nil {
		t.Fatal("A calendar accessible after switch to B")
	}
}
