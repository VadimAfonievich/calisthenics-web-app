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
		for _, q := range []string{`DELETE FROM role_change_audit WHERE actor_user_id=$1`, `DELETE FROM admin_users WHERE user_id IN ($1,$2)`, `DELETE FROM profiles WHERE user_id IN ($1,$2)`, `DELETE FROM users WHERE id IN ($1,$2)`} {
			_, _ = pool.Exec(ctx, q, actor, target)
		}
	}()
	for i, id := range []string{actor, target} {
		if _, err = pool.Exec(ctx, `INSERT INTO users(id,telegram_id,first_name) VALUES($1,$2,'Role test')`, id, 940000001+i); err != nil {
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
	user, err := store.GetByID(ctx, target)
	if err != nil || user.Role != "coach" {
		t.Fatalf("promote: %+v %v", user, err)
	}
	if err = store.SetCoachRole(ctx, actor, target, "user"); err != nil {
		t.Fatal(err)
	}
	user, err = store.GetByID(ctx, target)
	if err != nil || user.Role != "user" {
		t.Fatalf("demote: %+v %v", user, err)
	}
	if err = store.SetCoachRole(ctx, actor, actor, "coach"); err == nil {
		t.Fatal("self role mutation must be rejected")
	}
	var audits int
	if err = pool.QueryRow(ctx, `SELECT count(*) FROM role_change_audit WHERE actor_user_id=$1 AND target_user_id=$2`, actor, target).Scan(&audits); err != nil || audits != 2 {
		t.Fatalf("audit rows=%d err=%v", audits, err)
	}
}
