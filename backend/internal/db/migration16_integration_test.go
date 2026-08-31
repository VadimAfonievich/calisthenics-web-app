package db_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/jackc/pgx/v5"
)

var preservedTables = []string{"users", "profiles", "exercises", "lessons", "workouts", "programs", "skills", "media_assets", "workout_sessions", "exercise_sets", "user_program_progress", "user_skill_progress", "user_lesson_progress", "user_training_schedules", "user_planned_workouts", "user_exercise_stats"}

func migrationSQL(t *testing.T, name string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("migrations", name))
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func tableCounts(t *testing.T, ctx context.Context, conn *pgx.Conn) map[string]int {
	t.Helper()
	out := map[string]int{}
	for _, table := range preservedTables {
		var count int
		if err := conn.QueryRow(ctx, `SELECT count(*) FROM `+table).Scan(&count); err != nil {
			t.Fatalf("count %s: %v", table, err)
		}
		out[table] = count
	}
	return out
}

func TestRepresentativeV15ToV16MigrationPostgres(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL is not set")
	}
	ctx := context.Background()
	conn, err := pgx.Connect(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close(ctx)
	const schema = "phase19a_v15_fixture"
	_, _ = conn.Exec(ctx, `DROP SCHEMA IF EXISTS `+schema+` CASCADE`)
	if _, err = conn.Exec(ctx, `CREATE SCHEMA `+schema+`; SET search_path TO `+schema+`,public`); err != nil {
		t.Fatal(err)
	}
	defer conn.Exec(ctx, `DROP SCHEMA IF EXISTS `+schema+` CASCADE`)
	for version := 1; version <= 15; version++ {
		matches, globErr := filepath.Glob(filepath.Join("migrations", fmt.Sprintf("%06d_*.up.sql", version)))
		if globErr != nil || len(matches) != 1 {
			t.Fatalf("migration %d: matches=%v err=%v", version, matches, globErr)
		}
		data, readErr := os.ReadFile(matches[0])
		if readErr != nil {
			t.Fatal(readErr)
		}
		if _, err = conn.Exec(ctx, string(data)); err != nil {
			t.Fatalf("migration %d: %v", version, err)
		}
	}
	fixture, err := os.ReadFile(filepath.Join("testdata", "phase19a_legacy_v15.sql"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err = conn.Exec(ctx, string(fixture)); err != nil {
		t.Fatalf("v15 fixture: %v", err)
	}
	before := tableCounts(t, ctx, conn)
	var globalBefore, ownedBefore, adminBefore int
	if err = conn.QueryRow(ctx, `SELECT count(*) FILTER(WHERE owner_user_id IS NULL),count(*) FILTER(WHERE owner_user_id IS NOT NULL) FROM exercises`).Scan(&globalBefore, &ownedBefore); err != nil {
		t.Fatal(err)
	}
	if err = conn.QueryRow(ctx, `SELECT count(*) FROM admin_users`).Scan(&adminBefore); err != nil {
		t.Fatal(err)
	}

	if _, err = conn.Exec(ctx, migrationSQL(t, "000016_multi_tenancy.up.sql")); err != nil {
		t.Fatalf("migration 16 up: %v", err)
	}
	after := tableCounts(t, ctx, conn)
	for _, table := range preservedTables {
		if before[table] != after[table] {
			t.Errorf("%s count changed: %d -> %d", table, before[table], after[table])
		}
	}
	var globalAfter, ownedAfter int
	if err = conn.QueryRow(ctx, `SELECT count(*) FILTER(WHERE tenant_id IS NULL AND owner_user_id IS NULL AND standard_key IS NOT NULL),count(*) FILTER(WHERE tenant_id IS NOT NULL AND owner_user_id IS NOT NULL) FROM exercises`).Scan(&globalAfter, &ownedAfter); err != nil {
		t.Fatal(err)
	}
	if globalBefore != globalAfter || ownedBefore != ownedAfter {
		t.Fatalf("exercise ownership changed: global %d->%d owned %d->%d", globalBefore, globalAfter, ownedBefore, ownedAfter)
	}
	var adminAfter int
	if err = conn.QueryRow(ctx, `SELECT count(*) FROM admin_users`).Scan(&adminAfter); err != nil {
		t.Fatal(err)
	}
	if adminBefore != 3 || adminAfter != 1 {
		t.Fatalf("platform coach-role conversion: admin_users %d -> %d", adminBefore, adminAfter)
	}
	var owners, coachMemberships, duplicateMemberships int
	if err = conn.QueryRow(ctx, `SELECT count(*) FROM tenants WHERE owner_user_id IN ('97000000-0000-0000-0000-000000000002','97000000-0000-0000-0000-000000000003')`).Scan(&owners); err != nil {
		t.Fatal(err)
	}
	if err = conn.QueryRow(ctx, `SELECT count(*) FROM tenant_memberships m JOIN tenants t ON t.id=m.tenant_id WHERE m.role='coach' AND m.status='active' AND m.user_id=t.owner_user_id AND t.owner_user_id IN ('97000000-0000-0000-0000-000000000002','97000000-0000-0000-0000-000000000003')`).Scan(&coachMemberships); err != nil {
		t.Fatal(err)
	}
	if err = conn.QueryRow(ctx, `SELECT count(*) FROM (SELECT tenant_id,user_id,count(*) FROM tenant_memberships GROUP BY tenant_id,user_id HAVING count(*)>1) d`).Scan(&duplicateMemberships); err != nil {
		t.Fatal(err)
	}
	if owners != 2 || coachMemberships != 2 || duplicateMemberships != 0 {
		t.Fatalf("owner membership backfill: tenants=%d coaches=%d duplicates=%d", owners, coachMemberships, duplicateMemberships)
	}
	var badContent, badActivity, unrelated int
	if err = conn.QueryRow(ctx, `SELECT count(*) FROM (SELECT tenant_id,owner_user_id FROM lessons UNION ALL SELECT tenant_id,owner_user_id FROM exercises WHERE owner_user_id IS NOT NULL UNION ALL SELECT tenant_id,owner_user_id FROM programs UNION ALL SELECT tenant_id,owner_user_id FROM workouts UNION ALL SELECT tenant_id,owner_user_id FROM skills UNION ALL SELECT tenant_id,owner_user_id FROM media_assets) c JOIN tenants t ON t.id=c.tenant_id WHERE c.owner_user_id IS DISTINCT FROM t.owner_user_id`).Scan(&badContent); err != nil {
		t.Fatal(err)
	}
	if err = conn.QueryRow(ctx, `SELECT count(*) FROM (SELECT s.tenant_id,w.tenant_id expected FROM workout_sessions s JOIN workouts w ON w.id=s.workout_id UNION ALL SELECT p.tenant_id,x.tenant_id FROM user_program_progress p JOIN programs x ON x.id=p.program_id UNION ALL SELECT p.tenant_id,x.tenant_id FROM user_skill_progress p JOIN skills x ON x.id=p.skill_id UNION ALL SELECT p.tenant_id,x.tenant_id FROM user_lesson_progress p JOIN lessons x ON x.id=p.lesson_id UNION ALL SELECT p.tenant_id,x.tenant_id FROM user_training_schedules p JOIN workouts x ON x.id=p.workout_id UNION ALL SELECT p.tenant_id,x.tenant_id FROM user_planned_workouts p JOIN workouts x ON x.id=p.workout_id UNION ALL SELECT p.tenant_id,x.tenant_id FROM user_exercise_stats p JOIN exercises x ON x.id=p.exercise_id) a WHERE tenant_id IS DISTINCT FROM expected`).Scan(&badActivity); err != nil {
		t.Fatal(err)
	}
	if err = conn.QueryRow(ctx, `SELECT count(*) FROM tenant_memberships m JOIN tenants t ON t.id=m.tenant_id WHERE m.role='student' AND m.user_id='97000000-0000-0000-0000-000000000005' AND t.owner_user_id='97000000-0000-0000-0000-000000000003'`).Scan(&unrelated); err != nil {
		t.Fatal(err)
	}
	if badContent != 0 || badActivity != 0 || unrelated != 0 {
		t.Fatalf("unsafe backfill: content=%d activity=%d unrelated=%d", badContent, badActivity, unrelated)
	}

	if _, err = conn.Exec(ctx, migrationSQL(t, "000016_multi_tenancy.down.sql")); err != nil {
		t.Fatalf("down 16: %v", err)
	}
	if _, err = conn.Exec(ctx, migrationSQL(t, "000016_multi_tenancy.up.sql")); err != nil {
		t.Fatalf("up 16 after down: %v", err)
	}
	var tenantCount int
	if err = conn.QueryRow(ctx, `SELECT count(*) FROM tenants`).Scan(&tenantCount); err != nil || tenantCount < 2 {
		t.Fatalf("down/up tenant reconstruction: %d %v", tenantCount, err)
	}
}
