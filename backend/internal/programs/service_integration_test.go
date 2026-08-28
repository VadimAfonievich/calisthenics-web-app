package programs

import (
	"context"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

func TestSelfServiceProgramProgressPostgres(t *testing.T) {
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

	const program = "93000000-0000-0000-0000-000000000001"
	const level1 = "93000000-0000-0000-0000-000000000011"
	const level2 = "93000000-0000-0000-0000-000000000012"
	const workout1 = "93000000-0000-0000-0000-000000000021"
	const workout2 = "93000000-0000-0000-0000-000000000022"
	const student1 = "93000000-0000-0000-0000-000000000031"
	const student2 = "93000000-0000-0000-0000-000000000032"

	defer func() {
		for _, q := range []string{
			`DELETE FROM workout_sessions WHERE user_id IN ($1,$2)`,
			`DELETE FROM user_program_progress WHERE user_id IN ($1,$2)`,
			`DELETE FROM workouts WHERE id IN ($3,$4)`,
			`DELETE FROM program_levels WHERE id IN ($5,$6)`,
			`DELETE FROM programs WHERE id=$7`,
			`DELETE FROM user_progress WHERE user_id IN ($1,$2)`,
			`DELETE FROM profiles WHERE user_id IN ($1,$2)`,
			`DELETE FROM users WHERE id IN ($1,$2)`,
		} {
			_, _ = pool.Exec(ctx, q, student1, student2, workout1, workout2, level1, level2, program)
		}
	}()

	for i, id := range []string{student1, student2} {
		if _, err = pool.Exec(ctx, `INSERT INTO users(id,telegram_id,first_name) VALUES($1,$2,$3)`, id, 930000031+i, "Self-service student"); err != nil {
			t.Fatal(err)
		}
		if _, err = pool.Exec(ctx, `INSERT INTO profiles(user_id,display_name) VALUES($1,'Self-service student')`, id); err != nil {
			t.Fatal(err)
		}
		if _, err = pool.Exec(ctx, `INSERT INTO user_progress(user_id) VALUES($1)`, id); err != nil {
			t.Fatal(err)
		}
	}
	if _, err = pool.Exec(ctx, `INSERT INTO programs(id,name,slug,description,difficulty,duration_weeks,published,status,category) VALUES($1,'E2E program','phase-19-e2e','test','beginner',2,true,'published','SKILL')`, program); err != nil {
		t.Fatal(err)
	}
	if _, err = pool.Exec(ctx, `INSERT INTO program_levels(id,program_id,level_number,title,description,difficulty,unlock_rule_type,sort_order) VALUES($1,$3,1,'First','test','beginner','none',1),($2,$3,2,'Second','test','beginner','previous_level',2)`, level1, level2, program); err != nil {
		t.Fatal(err)
	}
	if _, err = pool.Exec(ctx, `INSERT INTO workouts(id,program_id,program_level_id,day_number,title,description,estimated_minutes,sort_order,status) VALUES($1,$3,$4,1,'First workout','test',10,1,'published'),($2,$3,$5,2,'Second workout','test',10,2,'published')`, workout1, workout2, program, level1, level2); err != nil {
		t.Fatal(err)
	}

	svc := NewService(pool)
	first, err := svc.Start(ctx, student1, program)
	if err != nil || first.Status != "active" || first.CurrentLevel != 1 {
		t.Fatalf("start: %+v %v", first, err)
	}
	if _, err = svc.Start(ctx, student1, program); err != nil {
		t.Fatal(err)
	}
	if _, err = svc.Start(ctx, student2, program); err != nil {
		t.Fatal(err)
	}
	var rows int
	if err = pool.QueryRow(ctx, `SELECT count(*) FROM user_program_progress WHERE user_id=$1 AND program_id=$2`, student1, program).Scan(&rows); err != nil || rows != 1 {
		t.Fatalf("start is not idempotent: rows=%d err=%v", rows, err)
	}
	if _, err = pool.Exec(ctx, `INSERT INTO workout_sessions(user_id,workout_id,status,completed_at) VALUES($1,$2,'completed',NOW())`, student1, workout1); err != nil {
		t.Fatal(err)
	}
	detail, err := svc.Get(ctx, student1, program)
	if err != nil || detail.CurrentLevel != 2 || detail.Levels[0].Status != "completed" || detail.Levels[1].Status != "current" {
		t.Fatalf("level unlock: %+v %v", detail, err)
	}
	other, err := svc.Get(ctx, student2, program)
	if err != nil || other.CurrentLevel != 1 {
		t.Fatalf("student isolation: %+v %v", other, err)
	}
	if _, err = pool.Exec(ctx, `INSERT INTO workout_sessions(user_id,workout_id,status,completed_at) VALUES($1,$2,'completed',NOW())`, student1, workout2); err != nil {
		t.Fatal(err)
	}
	detail, err = svc.Get(ctx, student1, program)
	if err != nil || detail.ProgressStatus != "completed" {
		t.Fatalf("program completion: %+v %v", detail, err)
	}
}
