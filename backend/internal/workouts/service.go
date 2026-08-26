package workouts

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"time"
)

const CompletionXP = 100

func countsAsFullWorkout(category string) bool { return category != "warmup" }

var (
	ErrNotFound  = errors.New("workout not found")
	ErrForbidden = errors.New("session forbidden")
	ErrClosed    = errors.New("session closed")
)

type Exercise struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	Sets           int16  `json:"sets"`
	TargetReps     *int16 `json:"target_reps,omitempty"`
	TargetDuration *int32 `json:"target_duration_seconds,omitempty"`
	Rest           int32  `json:"rest_seconds"`
	DemoMediaURL   string `json:"demo_media_url,omitempty"`
	DemoMediaType  string `json:"demo_media_type,omitempty"`
	DemoMediaMIME  string `json:"demo_media_mime_type,omitempty"`
	DemoPosterURL  string `json:"demo_poster_url,omitempty"`
}
type Workout struct {
	ID              string     `json:"id"`
	Title           string     `json:"title"`
	Description     string     `json:"description"`
	Minutes         int32      `json:"estimated_minutes"`
	Difficulty      string     `json:"difficulty"`
	ProgramID       string     `json:"program_id,omitempty"`
	ProgramName     string     `json:"program_name,omitempty"`
	Category        string     `json:"category"`
	WarmupEnabled   bool       `json:"warmup_enabled"`
	WarmupWorkoutID *string    `json:"warmup_workout_id,omitempty"`
	DefaultWarmup   *Warmup    `json:"default_warmup,omitempty"`
	CoverMediaURL   string     `json:"cover_media_url,omitempty"`
	Exercises       []Exercise `json:"exercises"`
}
type Warmup struct {
	ID      string `json:"id"`
	Title   string `json:"title"`
	Minutes int32  `json:"estimated_minutes"`
}
type CatalogItem struct {
	ID              string  `json:"id"`
	Title           string  `json:"title"`
	Description     string  `json:"description"`
	Minutes         int32   `json:"estimated_minutes"`
	Difficulty      string  `json:"difficulty"`
	ExerciseCount   int32   `json:"exercise_count"`
	ProgramID       string  `json:"program_id,omitempty"`
	ProgramName     string  `json:"program_name,omitempty"`
	Category        string  `json:"category"`
	WarmupEnabled   bool    `json:"warmup_enabled"`
	Status          *string `json:"status,omitempty"`
	ActiveSessionID *string `json:"active_session_id,omitempty"`
}
type Session struct {
	ID                   string    `json:"id"`
	WorkoutID            string    `json:"workout_id"`
	Status               string    `json:"status"`
	Duration             int32     `json:"duration_seconds"`
	XP                   int32     `json:"xp_earned"`
	CurrentStreak        int32     `json:"current_streak"`
	UnlockedAchievements []string  `json:"unlocked_achievements"`
	StartedAt            time.Time `json:"started_at"`
	PlannedWorkoutID     *string   `json:"planned_workout_id,omitempty"`
	Purpose              string    `json:"session_purpose"`
	FollowUpWorkoutID    *string   `json:"follow_up_workout_id,omitempty"`
	FollowUpWorkoutTitle string    `json:"follow_up_workout_title,omitempty"`
	FollowUpPlannedID    *string   `json:"follow_up_planned_workout_id,omitempty"`
	ContinuedSessionID   *string   `json:"continued_session_id,omitempty"`
	NextSession          *Session  `json:"next_session,omitempty"`
}
type CompletedSet struct {
	ExerciseID string `json:"exercise_id"`
	SetNumber  int16  `json:"set_number"`
	Reps       *int16 `json:"reps,omitempty"`
	Duration   *int32 `json:"duration_seconds,omitempty"`
	Completed  bool   `json:"completed"`
}
type ActiveSession struct {
	Session       Session        `json:"session"`
	Workout       Workout        `json:"workout"`
	CompletedSets []CompletedSet `json:"completed_sets"`
}
type SetInput struct {
	ExerciseID string `json:"exercise_id"`
	Number     int16  `json:"set_number"`
	Reps       *int16 `json:"reps,omitempty"`
	Duration   *int32 `json:"duration_seconds,omitempty"`
	Completed  bool   `json:"completed"`
}
type StartInput struct {
	PlannedWorkoutID  *string
	FollowUpWorkoutID *string
}
type Service struct{ pool *pgxpool.Pool }

func NewService(p *pgxpool.Pool) *Service { return &Service{p} }
func (s *Service) workout(ctx context.Context, id string) (Workout, error) {
	var w Workout
	e := s.pool.QueryRow(ctx, `SELECT w.id::text,w.title,w.description,w.estimated_minutes,w.difficulty,COALESCE(p.id::text,''),COALESCE(p.name,''),w.category,w.warmup_enabled,w.warmup_workout_id::text,COALESCE(m.url,'') FROM workouts w LEFT JOIN programs p ON p.id=w.program_id LEFT JOIN media_assets m ON m.id=w.cover_media_id WHERE w.id=$1::uuid AND w.status='published'`, id).Scan(&w.ID, &w.Title, &w.Description, &w.Minutes, &w.Difficulty, &w.ProgramID, &w.ProgramName, &w.Category, &w.WarmupEnabled, &w.WarmupWorkoutID, &w.CoverMediaURL)
	if errors.Is(e, pgx.ErrNoRows) {
		return w, ErrNotFound
	}
	if e != nil {
		return w, e
	}
	if w.WarmupEnabled && w.Category != "warmup" {
		var warmup Warmup
		e = s.pool.QueryRow(ctx, `SELECT id::text,title,estimated_minutes FROM workouts WHERE category='warmup' AND status='published' AND ($1::uuid IS NULL OR id=$1::uuid) ORDER BY is_default_warmup DESC,sort_order,id LIMIT 1`, w.WarmupWorkoutID).Scan(&warmup.ID, &warmup.Title, &warmup.Minutes)
		if e == nil {
			w.DefaultWarmup = &warmup
		} else if !errors.Is(e, pgx.ErrNoRows) {
			return w, e
		}
	}
	rows, e := s.pool.Query(ctx, `SELECT we.exercise_id::text,e.name,we.sets,we.target_reps,we.target_duration_seconds,we.rest_seconds,COALESCE(dm.url,''),COALESCE(dm.type,''),COALESCE(dm.mime_type,''),COALESCE(dm.thumbnail_url,'') FROM workout_exercises we JOIN exercises e ON e.id=we.exercise_id LEFT JOIN media_assets dm ON dm.id=e.demo_media_id WHERE we.workout_id=$1::uuid ORDER BY we.sort_order`, id)
	if e != nil {
		return w, e
	}
	defer rows.Close()
	for rows.Next() {
		var x Exercise
		if e = rows.Scan(&x.ID, &x.Name, &x.Sets, &x.TargetReps, &x.TargetDuration, &x.Rest, &x.DemoMediaURL, &x.DemoMediaType, &x.DemoMediaMIME, &x.DemoPosterURL); e != nil {
			return w, e
		}
		w.Exercises = append(w.Exercises, x)
	}
	return w, rows.Err()
}
func (s *Service) List(ctx context.Context, userID string) ([]CatalogItem, error) {
	rows, err := s.pool.Query(ctx, `SELECT w.id::text,w.title,w.description,w.estimated_minutes,w.difficulty,COUNT(we.id)::int,COALESCE(p.id::text,''),COALESCE(p.name,''),w.category,w.warmup_enabled,active.status,active.id::text FROM workouts w LEFT JOIN programs p ON p.id=w.program_id LEFT JOIN workout_exercises we ON we.workout_id=w.id LEFT JOIN LATERAL (SELECT ws.id,ws.status FROM workout_sessions ws WHERE ws.workout_id=w.id AND ws.user_id=$1::uuid ORDER BY ws.started_at DESC LIMIT 1) active ON true WHERE w.status='published' GROUP BY w.id,p.id,active.status,active.id ORDER BY w.category,w.difficulty,p.name,w.sort_order,w.day_number NULLS LAST`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []CatalogItem{}
	for rows.Next() {
		var item CatalogItem
		if err = rows.Scan(&item.ID, &item.Title, &item.Description, &item.Minutes, &item.Difficulty, &item.ExerciseCount, &item.ProgramID, &item.ProgramName, &item.Category, &item.WarmupEnabled, &item.Status, &item.ActiveSessionID); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}
func (s *Service) Today(ctx context.Context) (Workout, error) {
	var id string
	e := s.pool.QueryRow(ctx, `SELECT w.id::text FROM workouts w WHERE w.status='published' AND w.category<>'warmup' ORDER BY w.difficulty,w.sort_order,w.day_number NULLS LAST LIMIT 1`).Scan(&id)
	if errors.Is(e, pgx.ErrNoRows) {
		return Workout{}, ErrNotFound
	}
	if e != nil {
		return Workout{}, e
	}
	return s.workout(ctx, id)
}
func (s *Service) Get(ctx context.Context, id string) (Workout, error) { return s.workout(ctx, id) }
func (s *Service) Start(ctx context.Context, u, w string, in StartInput) (Session, error) {
	workout, e := s.workout(ctx, w)
	if e != nil {
		return Session{}, e
	}
	purpose := "main"
	plannedForSession := in.PlannedWorkoutID
	if in.FollowUpWorkoutID != nil {
		main, err := s.workout(ctx, *in.FollowUpWorkoutID)
		if err != nil || main.DefaultWarmup == nil || main.DefaultWarmup.ID != w || workout.Category != "warmup" {
			return Session{}, ErrForbidden
		}
		purpose = "warmup"
		plannedForSession = nil
	}
	if in.PlannedWorkoutID != nil {
		var owner, plannedWorkout string
		if e := s.pool.QueryRow(ctx, `SELECT user_id::text,workout_id::text FROM user_planned_workouts WHERE id=$1::uuid AND status='scheduled'`, *in.PlannedWorkoutID).Scan(&owner, &plannedWorkout); errors.Is(e, pgx.ErrNoRows) {
			return Session{}, ErrNotFound
		} else if e != nil {
			return Session{}, e
		}
		expectedWorkout := w
		if in.FollowUpWorkoutID != nil {
			expectedWorkout = *in.FollowUpWorkoutID
		}
		if owner != u || plannedWorkout != expectedWorkout {
			return Session{}, ErrForbidden
		}
	}
	var x Session
	e = s.pool.QueryRow(ctx, `WITH updated AS (
UPDATE workout_sessions SET planned_workout_id=COALESCE(planned_workout_id,$3::uuid),session_purpose=$4,follow_up_workout_id=COALESCE(follow_up_workout_id,$5::uuid),follow_up_planned_workout_id=COALESCE(follow_up_planned_workout_id,$6::uuid),updated_at=NOW()
WHERE id=(SELECT id FROM workout_sessions WHERE user_id=$1::uuid AND workout_id=$2::uuid AND status='started' AND (follow_up_workout_id IS NULL OR follow_up_workout_id=$5::uuid) ORDER BY started_at DESC LIMIT 1)
RETURNING *), created AS (
INSERT INTO workout_sessions(user_id,workout_id,planned_workout_id,session_purpose,follow_up_workout_id,follow_up_planned_workout_id)
SELECT $1::uuid,$2::uuid,$3::uuid,$4,$5::uuid,$6::uuid WHERE NOT EXISTS(SELECT 1 FROM updated) RETURNING *)
SELECT id::text,workout_id::text,status,duration_seconds,xp_earned,started_at,planned_workout_id::text,session_purpose,follow_up_workout_id::text,follow_up_planned_workout_id::text,continued_session_id::text FROM updated
UNION ALL SELECT id::text,workout_id::text,status,duration_seconds,xp_earned,started_at,planned_workout_id::text,session_purpose,follow_up_workout_id::text,follow_up_planned_workout_id::text,continued_session_id::text FROM created LIMIT 1`, u, w, plannedForSession, purpose, in.FollowUpWorkoutID, in.PlannedWorkoutID).Scan(&x.ID, &x.WorkoutID, &x.Status, &x.Duration, &x.XP, &x.StartedAt, &x.PlannedWorkoutID, &x.Purpose, &x.FollowUpWorkoutID, &x.FollowUpPlannedID, &x.ContinuedSessionID)
	return x, e
}
func (s *Service) GetSession(ctx context.Context, userID, sessionID string) (ActiveSession, error) {
	var out ActiveSession
	err := s.pool.QueryRow(ctx, `SELECT ws.id::text,ws.workout_id::text,ws.status,ws.duration_seconds,ws.xp_earned,ws.started_at,ws.planned_workout_id::text,ws.session_purpose,ws.follow_up_workout_id::text,COALESCE(next.title,''),ws.follow_up_planned_workout_id::text,ws.continued_session_id::text FROM workout_sessions ws LEFT JOIN workouts next ON next.id=ws.follow_up_workout_id WHERE ws.id=$1::uuid AND ws.user_id=$2::uuid`, sessionID, userID).Scan(&out.Session.ID, &out.Session.WorkoutID, &out.Session.Status, &out.Session.Duration, &out.Session.XP, &out.Session.StartedAt, &out.Session.PlannedWorkoutID, &out.Session.Purpose, &out.Session.FollowUpWorkoutID, &out.Session.FollowUpWorkoutTitle, &out.Session.FollowUpPlannedID, &out.Session.ContinuedSessionID)
	if errors.Is(err, pgx.ErrNoRows) {
		return out, ErrForbidden
	}
	if err != nil {
		return out, err
	}
	out.Workout, err = s.workout(ctx, out.Session.WorkoutID)
	if err != nil {
		return out, err
	}
	rows, err := s.pool.Query(ctx, `SELECT exercise_id::text,set_number,reps,duration_seconds,completed FROM exercise_sets WHERE session_id=$1::uuid ORDER BY created_at,set_number`, sessionID)
	if err != nil {
		return out, err
	}
	defer rows.Close()
	out.CompletedSets = []CompletedSet{}
	for rows.Next() {
		var item CompletedSet
		if err = rows.Scan(&item.ExerciseID, &item.SetNumber, &item.Reps, &item.Duration, &item.Completed); err != nil {
			return out, err
		}
		out.CompletedSets = append(out.CompletedSets, item)
	}
	return out, rows.Err()
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
func ensureContinuation(ctx context.Context, tx pgx.Tx, user string, warmup *Session) error {
	if warmup.FollowUpWorkoutID == nil {
		return nil
	}
	var next Session
	if warmup.ContinuedSessionID != nil {
		e := tx.QueryRow(ctx, `SELECT id::text,workout_id::text,status,duration_seconds,xp_earned,started_at,planned_workout_id::text,session_purpose FROM workout_sessions WHERE id=$1::uuid AND user_id=$2::uuid`, *warmup.ContinuedSessionID, user).Scan(&next.ID, &next.WorkoutID, &next.Status, &next.Duration, &next.XP, &next.StartedAt, &next.PlannedWorkoutID, &next.Purpose)
		if e != nil {
			return e
		}
		warmup.NextSession = &next
		return nil
	}
	e := tx.QueryRow(ctx, `WITH existing AS (
SELECT * FROM workout_sessions WHERE user_id=$1::uuid AND workout_id=$2::uuid AND status='started' ORDER BY started_at DESC LIMIT 1), created AS (
INSERT INTO workout_sessions(user_id,workout_id,planned_workout_id,session_purpose) SELECT $1::uuid,$2::uuid,$3::uuid,'main' WHERE NOT EXISTS(SELECT 1 FROM existing) RETURNING *)
SELECT id::text,workout_id::text,status,duration_seconds,xp_earned,started_at,planned_workout_id::text,session_purpose FROM existing
UNION ALL SELECT id::text,workout_id::text,status,duration_seconds,xp_earned,started_at,planned_workout_id::text,session_purpose FROM created LIMIT 1`, user, *warmup.FollowUpWorkoutID, warmup.FollowUpPlannedID).Scan(&next.ID, &next.WorkoutID, &next.Status, &next.Duration, &next.XP, &next.StartedAt, &next.PlannedWorkoutID, &next.Purpose)
	if e != nil {
		return e
	}
	_, e = tx.Exec(ctx, `UPDATE workout_sessions SET continued_session_id=$2::uuid,updated_at=NOW() WHERE id=$1::uuid`, warmup.ID, next.ID)
	if e != nil {
		return e
	}
	warmup.ContinuedSessionID = &next.ID
	warmup.NextSession = &next
	return nil
}
func (s *Service) Complete(ctx context.Context, u, sid string, duration int32) (Session, error) {
	tx, e := s.pool.Begin(ctx)
	if e != nil {
		return Session{}, e
	}
	defer tx.Rollback(ctx)
	var x Session
	var category string
	e = tx.QueryRow(ctx, `SELECT ws.id::text,ws.workout_id::text,ws.status,ws.duration_seconds,ws.xp_earned,w.category,ws.planned_workout_id::text,ws.session_purpose,ws.follow_up_workout_id::text,COALESCE(next.title,''),ws.follow_up_planned_workout_id::text,ws.continued_session_id::text FROM workout_sessions ws JOIN workouts w ON w.id=ws.workout_id LEFT JOIN workouts next ON next.id=ws.follow_up_workout_id WHERE ws.id=$1::uuid AND ws.user_id=$2::uuid FOR UPDATE OF ws`, sid, u).Scan(&x.ID, &x.WorkoutID, &x.Status, &x.Duration, &x.XP, &category, &x.PlannedWorkoutID, &x.Purpose, &x.FollowUpWorkoutID, &x.FollowUpWorkoutTitle, &x.FollowUpPlannedID, &x.ContinuedSessionID)
	if errors.Is(e, pgx.ErrNoRows) {
		return x, ErrForbidden
	}
	if e != nil {
		return x, e
	}
	if x.Status == "completed" {
		_ = tx.QueryRow(ctx, `SELECT current_streak FROM profiles WHERE user_id=$1::uuid`, u).Scan(&x.CurrentStreak)
		x.UnlockedAchievements = []string{}
		if e = ensureContinuation(ctx, tx, u, &x); e != nil {
			return x, e
		}
		return x, tx.Commit(ctx)
	}
	if x.Status != "started" {
		return x, ErrClosed
	}
	var timezone string
	var current, longest int32
	var lastDate *time.Time
	e = tx.QueryRow(ctx, `SELECT timezone,current_streak,longest_streak,last_workout_date FROM profiles WHERE user_id=$1::uuid FOR UPDATE`, u).Scan(&timezone, &current, &longest, &lastDate)
	if e != nil {
		return x, e
	}
	location, err := time.LoadLocation(timezone)
	if err != nil {
		location = time.UTC
	}
	localNow := time.Now().In(location)
	today := time.Date(localNow.Year(), localNow.Month(), localNow.Day(), 0, 0, 0, 0, time.UTC)
	x.CurrentStreak = nextStreak(lastDate, current, today)
	if x.CurrentStreak > longest {
		longest = x.CurrentStreak
	}
	e = tx.QueryRow(ctx, `UPDATE workout_sessions SET status='completed',completed_at=NOW(),duration_seconds=$2 WHERE id=$1::uuid RETURNING status,duration_seconds`, sid, duration).Scan(&x.Status, &x.Duration)
	if e != nil {
		return x, e
	}
	var completedExercises int32
	e = tx.QueryRow(ctx, `SELECT COUNT(DISTINCT exercise_id)::int FROM exercise_sets WHERE session_id=$1::uuid AND completed`, sid).Scan(&completedExercises)
	if e != nil {
		return x, e
	}
	_, e = tx.Exec(ctx, `INSERT INTO user_exercise_stats(user_id,exercise_id,total_sets,total_reps,max_reps,total_duration_seconds,max_duration_seconds,last_performed_at) SELECT $2::uuid,exercise_id,COUNT(*)::int,COALESCE(SUM(reps),0)::int,COALESCE(MAX(reps),0)::int,COALESCE(SUM(duration_seconds),0)::bigint,COALESCE(MAX(duration_seconds),0)::int,NOW() FROM exercise_sets WHERE session_id=$1::uuid AND completed GROUP BY exercise_id ON CONFLICT(user_id,exercise_id) DO UPDATE SET total_sets=user_exercise_stats.total_sets+EXCLUDED.total_sets,total_reps=user_exercise_stats.total_reps+EXCLUDED.total_reps,max_reps=GREATEST(user_exercise_stats.max_reps,EXCLUDED.max_reps),total_duration_seconds=user_exercise_stats.total_duration_seconds+EXCLUDED.total_duration_seconds,max_duration_seconds=GREATEST(user_exercise_stats.max_duration_seconds,EXCLUDED.max_duration_seconds),last_performed_at=NOW()`, sid, u)
	if e != nil {
		return x, e
	}
	if !countsAsFullWorkout(category) {
		x.XP = 0
		x.CurrentStreak = current
		x.UnlockedAchievements = []string{}
		_, e = tx.Exec(ctx, `UPDATE user_progress SET total_completed_exercises=total_completed_exercises+$3,total_training_seconds=total_training_seconds+$2 WHERE user_id=$1::uuid`, u, duration, completedExercises)
		if e != nil {
			return x, fmt.Errorf("update warmup activity: %w", e)
		}
		if e = ensureContinuation(ctx, tx, u, &x); e != nil {
			return x, e
		}
		if e = tx.Commit(ctx); e != nil {
			return x, e
		}
		return x, nil
	}
	_, e = tx.Exec(ctx, `UPDATE user_progress SET total_workouts=total_workouts+1,total_completed_exercises=total_completed_exercises+$3,total_training_seconds=total_training_seconds+$2 WHERE user_id=$1::uuid`, u, duration, completedExercises)
	if e != nil {
		return x, fmt.Errorf("update progress: %w", e)
	}
	rows, err := tx.Query(ctx, `WITH metrics AS (SELECT up.total_workouts,p.current_streak FROM user_progress up JOIN profiles p ON p.user_id=up.user_id WHERE up.user_id=$1::uuid), eligible AS (SELECT a.id,a.code,a.xp_reward FROM achievements a,metrics m WHERE (a.condition_type='workouts_completed' AND m.total_workouts>=a.condition_value) OR (a.condition_type='streak_days' AND $2>=a.condition_value) OR (a.code='FIRST_PULL_UP' AND EXISTS(SELECT 1 FROM user_exercise_stats s JOIN exercises e ON e.id=s.exercise_id WHERE s.user_id=$1::uuid AND e.slug='podtyagivaniya' AND s.total_sets>0)) OR (a.code='TEN_PUSH_UPS' AND EXISTS(SELECT 1 FROM user_exercise_stats s JOIN exercises e ON e.id=s.exercise_id WHERE s.user_id=$1::uuid AND e.slug='otzhimaniya' AND s.max_reps>=a.condition_value))) INSERT INTO user_achievements(user_id,achievement_id) SELECT $1::uuid,id FROM eligible ON CONFLICT DO NOTHING RETURNING (SELECT code FROM achievements WHERE id=achievement_id),(SELECT xp_reward FROM achievements WHERE id=achievement_id)`, u, x.CurrentStreak)
	if err != nil {
		return x, err
	}
	achievementXP := int32(0)
	x.UnlockedAchievements = []string{}
	for rows.Next() {
		var code string
		var reward int32
		if err = rows.Scan(&code, &reward); err != nil {
			rows.Close()
			return x, err
		}
		x.UnlockedAchievements = append(x.UnlockedAchievements, code)
		achievementXP += reward
	}
	rows.Close()
	if err = rows.Err(); err != nil {
		return x, err
	}
	x.XP = CompletionXP + achievementXP
	_, e = tx.Exec(ctx, `UPDATE workout_sessions SET xp_earned=$2 WHERE id=$1::uuid`, sid, x.XP)
	if e != nil {
		return x, e
	}
	_, e = tx.Exec(ctx, `UPDATE user_planned_workouts pw SET status='completed',updated_at=NOW() FROM workout_sessions ws WHERE ws.id=$1::uuid AND ws.planned_workout_id=pw.id AND pw.user_id=$2::uuid`, sid, u)
	if e != nil {
		return x, e
	}
	_, e = tx.Exec(ctx, `UPDATE profiles SET xp=xp+$2,level=(SELECT level FROM levels WHERE min_xp<=xp+$2 ORDER BY min_xp DESC LIMIT 1),current_streak=$3,longest_streak=$4,last_workout_date=$5 WHERE user_id=$1::uuid`, u, x.XP, x.CurrentStreak, longest, today)
	if e != nil {
		return x, e
	}
	if e = tx.Commit(ctx); e != nil {
		return x, e
	}
	return x, nil
}
