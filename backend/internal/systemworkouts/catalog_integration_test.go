package systemworkouts

import (
	"context"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

func TestSeedIsIdempotentAndSecure(t *testing.T) {
	dsn := os.Getenv("STANDARD_EXERCISES_DATABASE_URL")
	if dsn == "" {
		t.Skip("STANDARD_EXERCISES_DATABASE_URL is not set")
	}
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	if err = Seed(ctx, pool); err != nil {
		t.Fatal(err)
	}
	if err = Seed(ctx, pool); err != nil {
		t.Fatal(err)
	}
	keys := []string{"system-morning-quick", "system-morning-medium", "system-morning-full", "system-warmup-standard"}
	var workouts, invalidExercises, invalidTargets int
	if err = pool.QueryRow(ctx, `SELECT count(*) FROM workouts WHERE standard_key=ANY($1)`, keys).Scan(&workouts); err != nil {
		t.Fatal(err)
	}
	if err = pool.QueryRow(ctx, `SELECT count(*) FROM workout_exercises we JOIN workouts w ON w.id=we.workout_id JOIN exercises e ON e.id=we.exercise_id WHERE w.standard_key=ANY($1) AND (e.tenant_id IS NOT NULL OR e.owner_user_id IS NOT NULL OR e.standard_key IS NULL)`, keys).Scan(&invalidExercises); err != nil {
		t.Fatal(err)
	}
	if err = pool.QueryRow(ctx, `SELECT count(*) FROM workout_exercises we JOIN workouts w ON w.id=we.workout_id WHERE w.standard_key=ANY($1) AND (we.target_reps IS NOT NULL OR we.target_duration_seconds IS NULL)`, keys).Scan(&invalidTargets); err != nil {
		t.Fatal(err)
	}
	if workouts != 4 || invalidExercises != 0 || invalidTargets != 0 {
		t.Fatalf("workouts=%d invalid exercises=%d targets=%d", workouts, invalidExercises, invalidTargets)
	}
}
