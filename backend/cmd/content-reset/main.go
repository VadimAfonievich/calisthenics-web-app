package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5"
)

type counts struct{ lessons, exercises, workouts, programs, skills int }
type conflictCounts struct{ workoutPrograms, workoutWarmups, skillRequirements, skillPrograms int }

var resetStatements = []string{
	`UPDATE lessons SET status='archived',published=false WHERE owner_user_id IS NULL AND status<>'archived'`,
	`UPDATE workouts SET status='archived',is_default_warmup=false WHERE owner_user_id IS NULL AND status<>'archived'`,
	`UPDATE programs SET status='archived',published=false WHERE owner_user_id IS NULL AND status<>'archived'`,
	`UPDATE skills SET status='archived',hidden=true WHERE owner_user_id IS NULL AND status<>'archived'`,
	`UPDATE exercises SET status='published' WHERE owner_user_id IS NULL AND status<>'published'`,
}

func main() {
	dryRun := flag.Bool("dry-run", false, "show the reset plan without changing data")
	systemOnly := flag.Bool("system-only", false, "target only owner_user_id IS NULL content")
	confirm := flag.Bool("confirm", false, "explicitly confirm the archive reset")
	flag.Parse()
	if !*systemOnly {
		log.Fatal("refusing to run: --system-only is required")
	}
	if !*dryRun && !*confirm {
		log.Fatal("refusing to mutate: use --dry-run or --confirm")
	}
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL is required")
	}
	ctx := context.Background()
	conn, err := pgx.Connect(ctx, dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close(ctx)
	var c counts
	err = conn.QueryRow(ctx, `SELECT
	 (SELECT count(*) FROM lessons WHERE owner_user_id IS NULL AND status<>'archived'),
	 (SELECT count(*) FROM exercises WHERE owner_user_id IS NULL),
	 (SELECT count(*) FROM workouts WHERE owner_user_id IS NULL AND status<>'archived'),
	 (SELECT count(*) FROM programs WHERE owner_user_id IS NULL AND status<>'archived'),
	 (SELECT count(*) FROM skills WHERE owner_user_id IS NULL AND status<>'archived')`).Scan(&c.lessons, &c.exercises, &c.workouts, &c.programs, &c.skills)
	if err != nil {
		log.Fatal(err)
	}
	var conflicts conflictCounts
	err = conn.QueryRow(ctx, `SELECT
	 (SELECT count(*) FROM workouts w JOIN programs p ON p.id=w.program_id WHERE w.owner_user_id IS NOT NULL AND w.status='published' AND p.owner_user_id IS NULL),
	 (SELECT count(*) FROM workouts w JOIN workouts warmup ON warmup.id=w.warmup_workout_id WHERE w.owner_user_id IS NOT NULL AND w.status='published' AND warmup.owner_user_id IS NULL),
	 (SELECT count(*) FROM skills s JOIN skill_requirements r ON r.skill_id=s.id JOIN skills required ON required.id=r.required_skill_id WHERE s.owner_user_id IS NOT NULL AND s.status='published' AND required.owner_user_id IS NULL),
	 (SELECT count(*) FROM skill_levels sl JOIN skills s ON s.id=sl.skill_id JOIN program_levels pl ON pl.id=sl.program_level_id JOIN programs p ON p.id=pl.program_id WHERE s.owner_user_id IS NOT NULL AND s.status='published' AND p.owner_user_id IS NULL)`).Scan(&conflicts.workoutPrograms, &conflicts.workoutWarmups, &conflicts.skillRequirements, &conflicts.skillPrograms)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Lessons to remove: %d\nExercises to remove: 0\nWorkouts to remove: %d\nPrograms to remove: %d\nSkills to remove: %d\n", c.lessons, c.workouts, c.programs, c.skills)
	totalConflicts := conflicts.workoutPrograms + conflicts.workoutWarmups + conflicts.skillRequirements + conflicts.skillPrograms
	fmt.Printf("System exercises retained: %d\nDependency conflicts: %d\n", c.exercises, totalConflicts)
	fmt.Printf("  coach workouts -> system programs: %d\n  coach workouts -> system warmups: %d\n  coach skills -> system prerequisites: %d\n  coach skills -> system programs: %d\n", conflicts.workoutPrograms, conflicts.workoutWarmups, conflicts.skillRequirements, conflicts.skillPrograms)
	if *dryRun {
		fmt.Println("DRY RUN: no changes applied")
		return
	}
	if totalConflicts > 0 {
		log.Fatal("refusing reset: resolve dependency conflicts first")
	}
	tx, err := conn.Begin(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer tx.Rollback(ctx)
	for _, q := range resetStatements {
		if _, err = tx.Exec(ctx, q); err != nil {
			log.Fatal(err)
		}
	}
	if err = tx.Commit(ctx); err != nil {
		log.Fatal(err)
	}
	fmt.Println("System content reset applied as a reversible logical archive. Coach and user data were not modified.")
}
