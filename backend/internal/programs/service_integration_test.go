package programs

import (
	"context"
	"errors"
	"github.com/calisthenics-coach/calisthenics-mini-app/backend/internal/middleware"
	workoutsvc "github.com/calisthenics-coach/calisthenics-mini-app/backend/internal/workouts"
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
	const tenant = "93000000-0000-0000-0000-000000000041"
	const tenantB = "93000000-0000-0000-0000-000000000042"
	const programB = "93000000-0000-0000-0000-000000000002"
	const levelB = "93000000-0000-0000-0000-000000000013"

	defer func() {
		_, _ = pool.Exec(ctx, `DELETE FROM workout_sessions WHERE user_id IN ($1,$2)`, student1, student2)
		_, _ = pool.Exec(ctx, `DELETE FROM user_program_level_mastery WHERE user_id IN ($1,$2)`, student1, student2)
		_, _ = pool.Exec(ctx, `DELETE FROM user_program_progress WHERE user_id IN ($1,$2)`, student1, student2)
		_, _ = pool.Exec(ctx, `DELETE FROM workouts WHERE id IN ($1,$2)`, workout1, workout2)
		_, _ = pool.Exec(ctx, `DELETE FROM program_levels WHERE id IN ($1,$2)`, level1, level2)
		_, _ = pool.Exec(ctx, `DELETE FROM programs WHERE id=$1`, program)
		_, _ = pool.Exec(ctx, `DELETE FROM program_levels WHERE id=$1`, levelB)
		_, _ = pool.Exec(ctx, `DELETE FROM programs WHERE id=$1`, programB)
		_, _ = pool.Exec(ctx, `DELETE FROM tenant_memberships WHERE tenant_id IN ($1,$2)`, tenant, tenantB)
		_, _ = pool.Exec(ctx, `DELETE FROM tenants WHERE id IN ($1,$2)`, tenant, tenantB)
		_, _ = pool.Exec(ctx, `DELETE FROM user_progress WHERE user_id IN ($1,$2)`, student1, student2)
		_, _ = pool.Exec(ctx, `DELETE FROM profiles WHERE user_id IN ($1,$2)`, student1, student2)
		_, _ = pool.Exec(ctx, `DELETE FROM users WHERE id IN ($1,$2)`, student1, student2)
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
	if _, err = pool.Exec(ctx, `INSERT INTO tenants(id,slug,name,owner_user_id) VALUES($1,'program-e2e','Program E2E',$2)`, tenant, student1); err != nil {
		t.Fatal(err)
	}
	if _, err = pool.Exec(ctx, `INSERT INTO tenant_memberships(tenant_id,user_id,role) VALUES($1,$2,'coach'),($1,$3,'student')`, tenant, student1, student2); err != nil {
		t.Fatal(err)
	}
	if _, err = pool.Exec(ctx, `INSERT INTO tenants(id,slug,name,owner_user_id) VALUES($1,'program-e2e-b','Program E2E B',$2)`, tenantB, student2); err != nil {
		t.Fatal(err)
	}
	if _, err = pool.Exec(ctx, `INSERT INTO tenant_memberships(tenant_id,user_id,role) VALUES($1,$2,'coach'),($1,$3,'student')`, tenantB, student2, student1); err != nil {
		t.Fatal(err)
	}
	ctx = middleware.WithTenant(ctx, tenant, "student")
	if _, err = pool.Exec(ctx, `INSERT INTO programs(id,name,slug,description,difficulty,duration_weeks,published,status,category,tenant_id,owner_user_id) VALUES($1,'E2E program','phase-19-e2e','test','beginner',2,true,'published','SKILL',$2,$3)`, program, tenant, student1); err != nil {
		t.Fatal(err)
	}
	if _, err = pool.Exec(ctx, `INSERT INTO programs(id,name,slug,description,difficulty,duration_weeks,published,status,category,tenant_id,owner_user_id) VALUES($1,'E2E program B','phase-19-e2e-b','test','beginner',2,true,'published','SKILL',$2,$3)`, programB, tenantB, student2); err != nil {
		t.Fatal(err)
	}
	if _, err = pool.Exec(ctx, `INSERT INTO program_levels(id,program_id,level_number,title,description,difficulty,unlock_rule_type,sort_order) VALUES($1,$2,1,'First B','test','beginner','none',1)`, levelB, programB); err != nil {
		t.Fatal(err)
	}
	if _, err = pool.Exec(ctx, `INSERT INTO program_levels(id,program_id,level_number,title,description,difficulty,unlock_rule_type,sort_order) VALUES($1,$3,1,'First','test','beginner','none',1),($2,$3,2,'Second','test','beginner','previous_level',2)`, level1, level2, program); err != nil {
		t.Fatal(err)
	}
	if _, err = pool.Exec(ctx, `INSERT INTO workouts(id,program_id,program_level_id,day_number,title,description,estimated_minutes,sort_order,status,tenant_id,owner_user_id) VALUES($1,$3,$4,1,'First workout','test',10,1,'published',$6,$7),($2,$3,$5,2,'Second workout','test',10,2,'published',$6,$7)`, workout1, workout2, program, level1, level2, tenant, student1); err != nil {
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
	ctxB := middleware.WithTenant(context.Background(), tenantB, "student")
	if _, err = svc.Start(ctxB, student1, programB); err != nil {
		t.Fatal(err)
	}
	var rows int
	if err = pool.QueryRow(ctx, `SELECT count(*) FROM user_program_progress WHERE user_id=$1 AND program_id=$2`, student1, program).Scan(&rows); err != nil || rows != 1 {
		t.Fatalf("start is not idempotent: rows=%d err=%v", rows, err)
	}
	if _, err = pool.Exec(ctx, `INSERT INTO workout_sessions(user_id,tenant_id,workout_id,status,completed_at) VALUES($1,$3,$2,'completed',NOW())`, student1, workout1, tenant); err != nil {
		t.Fatal(err)
	}
	detail, err := svc.Get(ctx, student1, program)
	if err != nil || detail.CurrentLevel != 1 || detail.Levels[0].Status != "current" || detail.Levels[1].Status != "locked" {
		t.Fatalf("workout must not unlock level: %+v %v", detail, err)
	}
	if _, err = workoutsvc.NewService(pool).Start(ctx, student1, workout2, workoutsvc.StartInput{}); !errors.Is(err, workoutsvc.ErrForbidden) {
		t.Fatalf("locked stage workout start error=%v", err)
	}
	confirmed, err := svc.ConfirmMastery(ctx, student1, program, level1)
	if err != nil || confirmed.CurrentLevel != 2 {
		t.Fatalf("confirm mastery: %+v %v", confirmed, err)
	}
	otherTenant, otherTenantErr := svc.Get(ctxB, student1, programB)
	if otherTenantErr != nil || otherTenant.CurrentLevel != 1 {
		t.Fatalf("tenant B changed by tenant A mastery: %+v %v", otherTenant, otherTenantErr)
	}
	if _, crossErr := svc.ConfirmMastery(ctxB, student1, programB, level1); !errors.Is(crossErr, ErrStageLocked) {
		t.Fatalf("cross-tenant mastery error=%v", crossErr)
	}
	again, err := svc.ConfirmMastery(ctx, student1, program, level1)
	if err != nil || again.CurrentLevel != 2 {
		t.Fatalf("idempotent confirm: %+v %v", again, err)
	}
	detail, err = svc.Get(ctx, student1, program)
	if err != nil || detail.Levels[0].Status != "completed" || detail.Levels[1].Status != "current" {
		t.Fatalf("level unlock after mastery: %+v %v", detail, err)
	}
	if started, startErr := workoutsvc.NewService(pool).Start(ctx, student1, workout2, workoutsvc.StartInput{}); startErr != nil || started.WorkoutID != workout2 {
		t.Fatalf("unlocked workout start: %+v %v", started, startErr)
	}
	other, err := svc.Get(ctx, student2, program)
	if err != nil || other.CurrentLevel != 1 {
		t.Fatalf("student isolation: %+v %v", other, err)
	}
	if _, err = pool.Exec(ctx, `INSERT INTO workout_sessions(user_id,tenant_id,workout_id,status,completed_at) VALUES($1,$3,$2,'completed',NOW())`, student1, workout2, tenant); err != nil {
		t.Fatal(err)
	}
	detail, err = svc.Get(ctx, student1, program)
	if err != nil || detail.ProgressStatus != "active" {
		t.Fatalf("workout completed program: %+v %v", detail, err)
	}
	final, err := svc.ConfirmMastery(ctx, student1, program, level2)
	if err != nil || !final.ProgramCompleted {
		t.Fatalf("final mastery: %+v %v", final, err)
	}
	detail, err = svc.Get(ctx, student1, program)
	if err != nil || detail.ProgressStatus != "completed" {
		t.Fatalf("program completion after mastery: %+v %v", detail, err)
	}
}
