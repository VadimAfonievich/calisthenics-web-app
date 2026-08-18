package workouts

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const CompletionXP = 100

var (
	ErrNotFound  = errors.New("workout not found")
	ErrForbidden = errors.New("session forbidden")
	ErrClosed    = errors.New("session closed")
)

type Exercise struct {
	ID, Name       string
	Sets           int16
	TargetReps     *int16 `json:"target_reps,omitempty"`
	TargetDuration *int32 `json:"target_duration_seconds,omitempty"`
	Rest           int32  `json:"rest_seconds"`
}
type Workout struct {
	ID, Title, Description string
	Minutes                int32      `json:"estimated_minutes"`
	Exercises              []Exercise `json:"exercises"`
}
type Session struct {
	ID, WorkoutID, Status string
	Duration              int32 `json:"duration_seconds"`
	XP                    int32 `json:"xp_earned"`
}
type SetInput struct {
	ExerciseID string `json:"exercise_id"`
	Number     int16  `json:"set_number"`
	Reps       *int16 `json:"reps,omitempty"`
	Duration   *int32 `json:"duration_seconds,omitempty"`
	Completed  bool   `json:"completed"`
}
type Service struct{ pool *pgxpool.Pool }

func NewService(p *pgxpool.Pool) *Service { return &Service{p} }
func (s *Service) workout(ctx context.Context, id string) (Workout, error) {
	var w Workout
	e := s.pool.QueryRow(ctx, `SELECT id::text,title,description,estimated_minutes FROM workouts WHERE id=$1::uuid`, id).Scan(&w.ID, &w.Title, &w.Description, &w.Minutes)
	if errors.Is(e, pgx.ErrNoRows) {
		return w, ErrNotFound
	}
	if e != nil {
		return w, e
	}
	rows, e := s.pool.Query(ctx, `SELECT we.exercise_id::text,e.name,we.sets,we.target_reps,we.target_duration_seconds,we.rest_seconds FROM workout_exercises we JOIN exercises e ON e.id=we.exercise_id WHERE we.workout_id=$1::uuid ORDER BY we.sort_order`, id)
	if e != nil {
		return w, e
	}
	defer rows.Close()
	for rows.Next() {
		var x Exercise
		if e = rows.Scan(&x.ID, &x.Name, &x.Sets, &x.TargetReps, &x.TargetDuration, &x.Rest); e != nil {
			return w, e
		}
		w.Exercises = append(w.Exercises, x)
	}
	return w, rows.Err()
}
func (s *Service) Today(ctx context.Context) (Workout, error) {
	var id string
	e := s.pool.QueryRow(ctx, `SELECT w.id::text FROM workouts w JOIN programs p ON p.id=w.program_id WHERE p.published ORDER BY p.difficulty,w.day_number LIMIT 1`).Scan(&id)
	if errors.Is(e, pgx.ErrNoRows) {
		return Workout{}, ErrNotFound
	}
	if e != nil {
		return Workout{}, e
	}
	return s.workout(ctx, id)
}
func (s *Service) Get(ctx context.Context, id string) (Workout, error) { return s.workout(ctx, id) }
func (s *Service) Start(ctx context.Context, u, w string) (Session, error) {
	if _, e := s.workout(ctx, w); e != nil {
		return Session{}, e
	}
	var x Session
	e := s.pool.QueryRow(ctx, `INSERT INTO workout_sessions(user_id,workout_id) VALUES($1::uuid,$2::uuid) RETURNING id::text,workout_id::text,status,duration_seconds,xp_earned`, u, w).Scan(&x.ID, &x.WorkoutID, &x.Status, &x.Duration, &x.XP)
	return x, e
}
func (s *Service) RecordSet(ctx context.Context, u, sid string, in SetInput) error {
	tag, e := s.pool.Exec(ctx, `INSERT INTO exercise_sets(session_id,exercise_id,set_number,reps,duration_seconds,completed) SELECT $1::uuid,$2::uuid,$3,$4,$5,$6 FROM workout_sessions ws WHERE ws.id=$1::uuid AND ws.user_id=$7::uuid AND ws.status='started' ON CONFLICT(session_id,exercise_id,set_number) DO UPDATE SET reps=EXCLUDED.reps,duration_seconds=EXCLUDED.duration_seconds,completed=EXCLUDED.completed`, sid, in.ExerciseID, in.Number, in.Reps, in.Duration, in.Completed, u)
	if e != nil {
		return e
	}
	if tag.RowsAffected() == 0 {
		return ErrForbidden
	}
	return nil
}
func (s *Service) Complete(ctx context.Context, u, sid string, duration int32) (Session, error) {
	tx, e := s.pool.Begin(ctx)
	if e != nil {
		return Session{}, e
	}
	defer tx.Rollback(ctx)
	var x Session
	e = tx.QueryRow(ctx, `SELECT id::text,workout_id::text,status,duration_seconds,xp_earned FROM workout_sessions WHERE id=$1::uuid AND user_id=$2::uuid FOR UPDATE`, sid, u).Scan(&x.ID, &x.WorkoutID, &x.Status, &x.Duration, &x.XP)
	if errors.Is(e, pgx.ErrNoRows) {
		return x, ErrForbidden
	}
	if e != nil {
		return x, e
	}
	if x.Status == "completed" {
		return x, nil
	}
	if x.Status != "started" {
		return x, ErrClosed
	}
	e = tx.QueryRow(ctx, `UPDATE workout_sessions SET status='completed',completed_at=NOW(),duration_seconds=$2,xp_earned=$3 WHERE id=$1::uuid RETURNING status,duration_seconds,xp_earned`, sid, duration, CompletionXP).Scan(&x.Status, &x.Duration, &x.XP)
	if e != nil {
		return x, e
	}
	_, e = tx.Exec(ctx, `UPDATE profiles SET xp=xp+$2 WHERE user_id=$1::uuid`, u, CompletionXP)
	if e != nil {
		return x, e
	}
	_, e = tx.Exec(ctx, `UPDATE user_progress SET total_workouts=total_workouts+1,total_training_seconds=total_training_seconds+$2 WHERE user_id=$1::uuid`, u, duration)
	if e != nil {
		return x, fmt.Errorf("update progress: %w", e)
	}
	if e = tx.Commit(ctx); e != nil {
		return x, e
	}
	return x, nil
}
