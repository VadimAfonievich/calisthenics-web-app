package users

import (
	"context"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

func TestSuperAdminCoachRoleAuditPostgres(t *testing.T) {
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
	const actor = "94000000-0000-0000-0000-000000000001"
	const target = "94000000-0000-0000-0000-000000000002"
	defer func() {
		_, _ = pool.Exec(ctx, `DELETE FROM role_change_audit WHERE actor_user_id=$1`, actor)
		_, _ = pool.Exec(ctx, `DELETE FROM tenant_memberships WHERE user_id IN ($1,$2)`, actor, target)
		_, _ = pool.Exec(ctx, `DELETE FROM tenants WHERE owner_user_id=$1`, target)
		_, _ = pool.Exec(ctx, `DELETE FROM admin_users WHERE user_id IN ($1,$2)`, actor, target)
		_, _ = pool.Exec(ctx, `DELETE FROM profiles WHERE user_id IN ($1,$2)`, actor, target)
		_, _ = pool.Exec(ctx, `DELETE FROM users WHERE id IN ($1,$2)`, actor, target)
	}()
	for i, id := range []string{actor, target} {
		username := ""
		if id == target {
			username = "Afonich"
		}
		if _, err = pool.Exec(ctx, `INSERT INTO users(id,telegram_id,username,first_name) VALUES($1,$2,NULLIF($3,''),'Role test')`, id, 940000001+i, username); err != nil {
			t.Fatal(err)
		}
		if _, err = pool.Exec(ctx, `INSERT INTO profiles(user_id,display_name) VALUES($1,'Role test')`, id); err != nil {
			t.Fatal(err)
		}
	}
	if _, err = pool.Exec(ctx, `INSERT INTO admin_users(user_id,role) VALUES($1,'super_admin')`, actor); err != nil {
		t.Fatal(err)
	}
	store := NewStore(pool)
	if err = store.SetCoachRole(ctx, actor, target, "coach"); err != nil {
		t.Fatal(err)
	}
	if err = store.SetCoachRole(ctx, actor, target, "coach"); err != nil {
		t.Fatalf("repeated promotion must be idempotent: %v", err)
	}
	user, err := store.GetByID(ctx, target)
	if err != nil || user.Role != "user" || len(user.Tenants) != 1 || user.Tenants[0].Role != "coach" {
		t.Fatalf("promote: %+v %v", user, err)
	}
	if user.Tenants[0].Slug != "afonich" {
		t.Fatalf("default slug=%q want afonich", user.Tenants[0].Slug)
	}
	if err = store.SetCoachRole(ctx, actor, target, "user"); err != nil {
		t.Fatal(err)
	}
	if err = store.SetCoachRole(ctx, actor, target, "user"); err != nil {
		t.Fatalf("repeated demotion must be idempotent: %v", err)
	}
	user, err = store.GetByID(ctx, target)
	if err != nil || user.Role != "user" {
		t.Fatalf("demote: %+v %v", user, err)
	}
	if err = store.SetCoachRole(ctx, actor, actor, "coach"); err == nil {
		t.Fatal("self role mutation must be rejected")
	}
	var audits int
	if err = pool.QueryRow(ctx, `SELECT count(*) FROM role_change_audit WHERE actor_user_id=$1 AND target_user_id=$2`, actor, target).Scan(&audits); err != nil || audits != 4 {
		t.Fatalf("audit rows=%d err=%v", audits, err)
	}
	var tenants, activeCoaches int
	if err = pool.QueryRow(ctx, `SELECT count(*),(SELECT count(*) FROM tenant_memberships WHERE user_id=$1 AND role='coach' AND status='active') FROM tenants WHERE owner_user_id=$1`, target).Scan(&tenants, &activeCoaches); err != nil || tenants != 1 || activeCoaches != 0 {
		t.Fatalf("demotion preservation: tenants=%d active_coaches=%d err=%v", tenants, activeCoaches, err)
	}
}

func TestCoachSpaceSettingsAndSlugSecurityPostgres(t *testing.T) {
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
	const coachA = "94500000-0000-0000-0000-000000000001"
	const coachB = "94500000-0000-0000-0000-000000000002"
	const student = "94500000-0000-0000-0000-000000000003"
	const tenantA = "94500000-0000-0000-0000-000000000011"
	const tenantB = "94500000-0000-0000-0000-000000000012"
	const lesson = "94500000-0000-0000-0000-000000000021"
	defer func() {
		_, _ = pool.Exec(ctx, `DELETE FROM user_lesson_progress WHERE user_id=$1`, student)
		_, _ = pool.Exec(ctx, `DELETE FROM lessons WHERE id=$1`, lesson)
		_, _ = pool.Exec(ctx, `DELETE FROM tenant_memberships WHERE tenant_id IN ($1,$2) OR user_id=$3`, tenantA, tenantB, student)
		_, _ = pool.Exec(ctx, `DELETE FROM tenants WHERE id IN ($1,$2)`, tenantA, tenantB)
		_, _ = pool.Exec(ctx, `DELETE FROM profiles WHERE user_id IN ($1,$2,$3)`, coachA, coachB, student)
		_, _ = pool.Exec(ctx, `DELETE FROM users WHERE id IN ($1,$2,$3)`, coachA, coachB, student)
	}()
	for i, id := range []string{coachA, coachB, student} {
		username := ""
		if id == coachA {
			username = "Afonich"
		}
		if _, err = pool.Exec(ctx, `INSERT INTO users(id,telegram_id,username,first_name) VALUES($1,$2,NULLIF($3,''),'Coach')`, id, 945000001+i, username); err != nil {
			t.Fatal(err)
		}
		if _, err = pool.Exec(ctx, `INSERT INTO profiles(user_id,display_name) VALUES($1,'Coach')`, id); err != nil {
			t.Fatal(err)
		}
	}
	if _, err = pool.Exec(ctx, `INSERT INTO tenants(id,slug,name,description,owner_user_id) VALUES($1,'old-school','Old A','Old description',$3),($2,'other-school','School B','B description',$4)`, tenantA, tenantB, coachA, coachB); err != nil {
		t.Fatal(err)
	}
	if _, err = pool.Exec(ctx, `INSERT INTO tenant_memberships(tenant_id,user_id,role) VALUES($1,$3,'coach'),($2,$4,'coach')`, tenantA, tenantB, coachA, coachB); err != nil {
		t.Fatal(err)
	}
	var category string
	if err = pool.QueryRow(ctx, `SELECT id::text FROM lesson_categories LIMIT 1`).Scan(&category); err != nil {
		t.Fatal(err)
	}
	if _, err = pool.Exec(ctx, `INSERT INTO lessons(id,category_id,title,slug,short_description,content,difficulty,duration_minutes,owner_user_id,tenant_id) VALUES($1,$2::uuid,'Keep me','keep-me','D','C','beginner',5,$3,$4)`, lesson, category, coachA, tenantA); err != nil {
		t.Fatal(err)
	}
	store := NewStore(pool)
	updated, err := store.UpdateOwnTenant(ctx, coachA, tenantA, "New A", "Student-facing description")
	if err != nil || updated.Name != "New A" || updated.Description != "Student-facing description" {
		t.Fatalf("settings: %+v %v", updated, err)
	}
	for _, slug := range []string{"bad_slug", "admin", "other-school"} {
		if _, err = store.UpdateOwnTenantSlug(ctx, coachA, tenantA, slug); err == nil {
			t.Fatalf("slug %q must be rejected", slug)
		}
	}
	if _, err = store.UpdateOwnTenantSlug(ctx, coachA, tenantB, "stolen-school"); err == nil {
		t.Fatal("Coach A changed Tenant B")
	}
	changed, err := store.UpdateOwnTenantSlug(ctx, coachA, tenantA, "new-school")
	if err != nil || changed.ID != tenantA || changed.Slug != "new-school" {
		t.Fatalf("slug change: %+v %v", changed, err)
	}
	var memberships, content int
	if err = pool.QueryRow(ctx, `SELECT (SELECT count(*) FROM tenant_memberships WHERE tenant_id=$1),(SELECT count(*) FROM lessons WHERE id=$2 AND tenant_id=$1)`, tenantA, lesson).Scan(&memberships, &content); err != nil || memberships != 1 || content != 1 {
		t.Fatalf("relationships changed: memberships=%d content=%d err=%v", memberships, content, err)
	}
	if old, oldErr := store.BootstrapTenant(ctx, student, "old-school"); oldErr == nil || old != nil {
		t.Fatalf("old slug still resolves: %+v %v", old, oldErr)
	}
	current, currentErr := store.BootstrapTenant(ctx, student, "new-school")
	if currentErr != nil || current == nil || current.ID != tenantA || current.Description != "Student-facing description" {
		t.Fatalf("new slug: %+v %v", current, currentErr)
	}
}
