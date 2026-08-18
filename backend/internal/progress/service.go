package progress

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"time"
)

type Level struct {
	Number int32  `json:"number"`
	Name   string `json:"name"`
	MinXP  int32  `json:"min_xp"`
}
type Summary struct {
	XP                 int32   `json:"xp"`
	Level              Level   `json:"level"`
	NextLevel          *Level  `json:"next_level,omitempty"`
	XPToNextLevel      int32   `json:"xp_to_next_level"`
	LevelProgress      float64 `json:"level_progress"`
	CurrentStreak      int32   `json:"current_streak"`
	LongestStreak      int32   `json:"longest_streak"`
	TotalWorkouts      int32   `json:"total_workouts"`
	CompletedExercises int32   `json:"completed_exercises"`
	TrainingSeconds    int64   `json:"training_seconds"`
}
type History struct {
	ID              string    `json:"id"`
	Title           string    `json:"title"`
	CompletedAt     time.Time `json:"completed_at"`
	DurationSeconds int32     `json:"duration_seconds"`
	XP              int32     `json:"xp_earned"`
}
type Week struct {
	WeekStart time.Time `json:"week_start"`
	Workouts  int32     `json:"workouts"`
	Seconds   int64     `json:"training_seconds"`
}
type Stats struct {
	TotalWorkouts      int32  `json:"total_workouts"`
	CompletedExercises int32  `json:"completed_exercises"`
	TrainingSeconds    int64  `json:"training_seconds"`
	Weeks              []Week `json:"weeks"`
}
type Achievement struct {
	Code        string     `json:"code"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Icon        string     `json:"icon"`
	XPReward    int32      `json:"xp_reward"`
	Unlocked    bool       `json:"unlocked"`
	UnlockedAt  *time.Time `json:"unlocked_at,omitempty"`
}
type Service struct{ pool *pgxpool.Pool }

func NewService(pool *pgxpool.Pool) *Service { return &Service{pool: pool} }

func (s *Service) Summary(ctx context.Context, userID string) (Summary, error) {
	var out Summary
	var nextNumber, nextXP *int32
	var nextName *string
	err := s.pool.QueryRow(ctx, `SELECT p.xp,l.level,l.name,l.min_xp,nl.level,nl.name,nl.min_xp,p.current_streak,p.longest_streak,up.total_workouts,up.total_completed_exercises,up.total_training_seconds FROM profiles p JOIN user_progress up ON up.user_id=p.user_id JOIN LATERAL (SELECT level,name,min_xp FROM levels WHERE min_xp<=p.xp ORDER BY min_xp DESC LIMIT 1) l ON true LEFT JOIN LATERAL (SELECT level,name,min_xp FROM levels WHERE min_xp>p.xp ORDER BY min_xp LIMIT 1) nl ON true WHERE p.user_id=$1::uuid`, userID).Scan(&out.XP, &out.Level.Number, &out.Level.Name, &out.Level.MinXP, &nextNumber, &nextName, &nextXP, &out.CurrentStreak, &out.LongestStreak, &out.TotalWorkouts, &out.CompletedExercises, &out.TrainingSeconds)
	if err != nil {
		return out, err
	}
	if nextNumber != nil {
		out.NextLevel = &Level{*nextNumber, *nextName, *nextXP}
		out.XPToNextLevel = *nextXP - out.XP
		out.LevelProgress = float64(out.XP-out.Level.MinXP) / float64(*nextXP-out.Level.MinXP)
	} else {
		out.LevelProgress = 1
	}
	return out, nil
}
func (s *Service) History(ctx context.Context, userID string) ([]History, error) {
	rows, err := s.pool.Query(ctx, `SELECT ws.id::text,w.title,ws.completed_at,ws.duration_seconds,ws.xp_earned FROM workout_sessions ws JOIN workouts w ON w.id=ws.workout_id WHERE ws.user_id=$1::uuid AND ws.status='completed' ORDER BY ws.completed_at DESC LIMIT 50`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []History{}
	for rows.Next() {
		var x History
		if err = rows.Scan(&x.ID, &x.Title, &x.CompletedAt, &x.DurationSeconds, &x.XP); err != nil {
			return nil, err
		}
		out = append(out, x)
	}
	return out, rows.Err()
}
func (s *Service) Stats(ctx context.Context, userID string) (Stats, error) {
	var out Stats
	if err := s.pool.QueryRow(ctx, `SELECT total_workouts,total_completed_exercises,total_training_seconds FROM user_progress WHERE user_id=$1::uuid`, userID).Scan(&out.TotalWorkouts, &out.CompletedExercises, &out.TrainingSeconds); err != nil {
		return out, err
	}
	rows, err := s.pool.Query(ctx, `SELECT date_trunc('week',completed_at)::date,COUNT(*)::int,COALESCE(SUM(duration_seconds),0)::bigint FROM workout_sessions WHERE user_id=$1::uuid AND status='completed' AND completed_at>=NOW()-INTERVAL '12 weeks' GROUP BY 1 ORDER BY 1`, userID)
	if err != nil {
		return out, err
	}
	defer rows.Close()
	out.Weeks = []Week{}
	for rows.Next() {
		var w Week
		if err = rows.Scan(&w.WeekStart, &w.Workouts, &w.Seconds); err != nil {
			return out, err
		}
		out.Weeks = append(out.Weeks, w)
	}
	return out, rows.Err()
}
func (s *Service) Achievements(ctx context.Context, userID string) ([]Achievement, error) {
	rows, err := s.pool.Query(ctx, `SELECT a.code,a.title,a.description,a.icon,a.xp_reward,ua.unlocked_at IS NOT NULL,ua.unlocked_at FROM achievements a LEFT JOIN user_achievements ua ON ua.achievement_id=a.id AND ua.user_id=$1::uuid ORDER BY (ua.unlocked_at IS NULL),a.condition_value,a.code`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Achievement{}
	for rows.Next() {
		var x Achievement
		if err = rows.Scan(&x.Code, &x.Title, &x.Description, &x.Icon, &x.XPReward, &x.Unlocked, &x.UnlockedAt); err != nil {
			return nil, err
		}
		out = append(out, x)
	}
	return out, rows.Err()
}
