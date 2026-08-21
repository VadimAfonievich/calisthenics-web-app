package workouts

import (
	"context"
	"os"
	"testing"

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
