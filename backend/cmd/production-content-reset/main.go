package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/jackc/pgx/v5"
)

type target struct {
	Name string
	SQL  string
}

var deleteTargets = []target{
	{"exercise_sets", `DELETE FROM exercise_sets`},
	{"workout_sessions", `DELETE FROM workout_sessions`},
	{"user_planned_workouts", `DELETE FROM user_planned_workouts`},
	{"user_training_schedule_days", `DELETE FROM user_training_schedule_days`},
	{"user_training_schedules", `DELETE FROM user_training_schedules`},
	{"user_lesson_progress", `DELETE FROM user_lesson_progress`},
	{"user_skill_criteria", `DELETE FROM user_skill_criteria`},
	{"user_skill_level_progress", `DELETE FROM user_skill_level_progress`},
	{"user_skill_progress", `DELETE FROM user_skill_progress`},
	{"user_exercise_stats", `DELETE FROM user_exercise_stats`},
	{"user_achievements", `DELETE FROM user_achievements`},
	{"workout_exercises", `DELETE FROM workout_exercises`},
	{"skill_requirements", `DELETE FROM skill_requirements`},
	{"skill_criteria", `DELETE FROM skill_criteria`},
	{"skill_levels", `DELETE FROM skill_levels`},
	{"lessons", `DELETE FROM lessons`},
	{"workouts", `DELETE FROM workouts`},
	{"program_levels", `DELETE FROM program_levels`},
	{"programs", `DELETE FROM programs`},
	{"skills", `DELETE FROM skills`},
	{"exercises", `DELETE FROM exercises`},
	{"media_assets", `DELETE FROM media_assets`},
}

var resetTargets = []target{
	{"profiles", `UPDATE profiles SET level=1,xp=0,current_streak=0,longest_streak=0,last_workout_date=NULL`},
	{"user_progress", `UPDATE user_progress SET total_workouts=0,total_completed_exercises=0,total_training_seconds=0`},
}

type preserved struct{ Users, Profiles, AdminUsers int64 }

func preservedCounts(ctx context.Context, q interface {
	QueryRow(context.Context, string, ...any) pgx.Row
}) (preserved, error) {
	var p preserved
	err := q.QueryRow(ctx, `SELECT (SELECT count(*) FROM users),(SELECT count(*) FROM profiles),(SELECT count(*) FROM admin_users)`).Scan(&p.Users, &p.Profiles, &p.AdminUsers)
	return p, err
}

func tableCount(ctx context.Context, q interface {
	QueryRow(context.Context, string, ...any) pgx.Row
}, table string) (int64, error) {
	allowed := false
	for _, x := range append(append([]target{}, deleteTargets...), resetTargets...) {
		if x.Name == table {
			allowed = true
			break
		}
	}
	if !allowed {
		return 0, errors.New("refusing unknown table: " + table)
	}
	var count int64
	err := q.QueryRow(ctx, `SELECT count(*) FROM `+table).Scan(&count)
	return count, err
}

func printPlan(ctx context.Context, conn *pgx.Conn) error {
	for _, x := range deleteTargets {
		count, err := tableCount(ctx, conn, x.Name)
		if err != nil {
			return err
		}
		fmt.Printf("DELETE %-34s %d\n", x.Name, count)
	}
	for _, x := range resetTargets {
		count, err := tableCount(ctx, conn, x.Name)
		if err != nil {
			return err
		}
		fmt.Printf("RESET  %-34s %d\n", x.Name, count)
	}
	p, err := preservedCounts(ctx, conn)
	if err != nil {
		return err
	}
	fmt.Printf("PRESERVE users=%d profiles=%d admin_users=%d\n", p.Users, p.Profiles, p.AdminUsers)
	return nil
}

func verifyMigration(ctx context.Context, conn *pgx.Conn) error {
	var version uint
	var dirty bool
	if err := conn.QueryRow(ctx, `SELECT version,dirty FROM schema_migrations`).Scan(&version, &dirty); err != nil {
		return err
	}
	if dirty || version != 13 {
		return fmt.Errorf("requires schema version=13 dirty=false; got version=%d dirty=%t", version, dirty)
	}
	return nil
}

func execute(ctx context.Context, conn *pgx.Conn) error {
	before, err := preservedCounts(ctx, conn)
	if err != nil {
		return err
	}
	tx, err := conn.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	for _, x := range deleteTargets {
		if _, err = tx.Exec(ctx, x.SQL); err != nil {
			return fmt.Errorf("%s: %w", x.Name, err)
		}
	}
	for _, x := range resetTargets {
		if _, err = tx.Exec(ctx, x.SQL); err != nil {
			return fmt.Errorf("%s: %w", x.Name, err)
		}
	}
	after, err := preservedCounts(ctx, tx)
	if err != nil {
		return err
	}
	if after != before {
		return fmt.Errorf("preserved identity counts changed: before=%+v after=%+v", before, after)
	}
	for _, x := range deleteTargets {
		count, e := tableCount(ctx, tx, x.Name)
		if e != nil {
			return e
		}
		if count != 0 {
			return fmt.Errorf("%s still has %d rows", x.Name, count)
		}
	}
	return tx.Commit(ctx)
}

func run() error {
	fs := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	all := fs.Bool("all-test-content", false, "target all disposable production test content")
	dry := fs.Bool("dry-run", false, "print destructive plan without writes")
	confirm := fs.Bool("confirm", false, "execute destructive reset transaction")
	if err := fs.Parse(os.Args[1:]); err != nil {
		return err
	}
	if !*all {
		return errors.New("refusing: --all-test-content is required")
	}
	if *dry == *confirm {
		return errors.New("choose exactly one of --dry-run or --confirm")
	}
	dsn := os.Getenv("DATABASE_URL")
	if strings.TrimSpace(dsn) == "" {
		return errors.New("DATABASE_URL is required")
	}
	ctx := context.Background()
	conn, err := pgx.Connect(ctx, dsn)
	if err != nil {
		return err
	}
	defer conn.Close(ctx)
	if err = verifyMigration(ctx, conn); err != nil {
		return err
	}
	if err = printPlan(ctx, conn); err != nil {
		return err
	}
	if *dry {
		fmt.Println("DRY RUN: no changes applied")
		return nil
	}
	if err = execute(ctx, conn); err != nil {
		return err
	}
	fmt.Println("ALL TEST CONTENT RESET COMPLETE: user identities, profiles and admin roles preserved")
	return nil
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
