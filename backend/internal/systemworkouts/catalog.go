package systemworkouts

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Exercise struct {
	Key      string
	Duration int
	Rest     int
}

type Workout struct {
	Key, Title, Description, Category, Difficulty string
	Minutes                                       int
	Exercises                                     []Exercise
}

func timed(keys []string, duration, rest int) []Exercise {
	out := make([]Exercise, len(keys))
	for i, key := range keys {
		out[i] = Exercise{Key: key, Duration: duration, Rest: rest}
	}
	return out
}

var Catalog = []Workout{
	{Key: "system-morning-quick", Title: "Быстрая зарядка", Description: "Короткая суставная зарядка на всё тело. Подходит утром или когда нужно быстро размяться.", Category: "morning", Difficulty: "beginner", Minutes: 6, Exercises: timed([]string{"neck-mobility", "neck-rotation", "shoulder-circles", "arm-open-close", "elbow-circles", "wrist-circles", "standing-torso-twist", "standing-side-bend", "pelvic-circles", "knee-circles-small", "heel-toe-rocks"}, 20, 10)},
	{Key: "system-morning-medium", Title: "Зарядка всего тела", Description: "Комбинированная зарядка: суставная мобилизация, лёгкая динамика и небольшая активация основных мышечных групп.", Category: "morning", Difficulty: "beginner", Minutes: 12, Exercises: timed([]string{"neck-side-tilt", "shoulder-circles", "arm-open-close", "wrist-circles", "standing-scapular-glide", "standing-torso-twist", "standing-side-bend", "standing-forward-fold-dynamic", "pelvic-circles", "hip-open-close", "leg-swing-front-back", "knee-flexion-extension", "ankle-circles", "jumping-jacks-light", "high-knees-light", "front-plank"}, 30, 15)},
	{Key: "system-morning-full", Title: "Полная зарядка", Description: "Полноценная зарядка на всё тело: суставная подготовка, динамическая работа, лёгкая силовая нагрузка и мышцы корпуса.", Category: "morning", Difficulty: "beginner", Minutes: 25, Exercises: timed([]string{"neck-mobility", "shoulder-circles", "arm-swings-front-back", "wrist-circles", "standing-scapular-glide", "standing-torso-twist", "standing-side-bend", "standing-forward-fold-dynamic", "standing-thoracic-open-book", "pelvic-circles", "hip-open-close", "leg-swing-front-back", "knee-circles-small", "ankle-circles", "jumping-jacks-light", "high-knees-light", "butt-kicks-light", "inchworm-walkout", "bodyweight-squat", "knee-push-up", "reverse-lunge", "front-plank", "mountain-climber", "bear-walk-light"}, 45, 15)},
	{Key: "system-warmup-standard", Title: "Стандартная разминка", Description: "Подготавливает суставы, плечи, лопатки, запястья, ноги и корпус к основной тренировке без лишнего утомления.", Category: "warmup", Difficulty: "beginner", Minutes: 12, Exercises: timed([]string{"neck-mobility", "shoulder-circles", "arm-open-close", "elbow-circles", "wrist-circles", "wrist-flexion-extension", "standing-scapular-glide", "standing-torso-twist", "standing-thoracic-open-book", "pelvic-circles", "hip-open-close", "leg-swing-front-back", "knee-flexion-extension", "ankle-circles", "jumping-jacks-light", "bodyweight-squat", "scapular-push-up"}, 30, 10)},
}

func Seed(ctx context.Context, pool *pgxpool.Pool) error {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	for order, workout := range Catalog {
		var id string
		err = tx.QueryRow(ctx, `INSERT INTO workouts(standard_key,title,description,difficulty,estimated_minutes,sort_order,category,warmup_enabled,is_default_warmup,status,published_at,tenant_id,owner_user_id) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,'published',NOW(),NULL,NULL) ON CONFLICT(standard_key) WHERE standard_key IS NOT NULL DO UPDATE SET title=EXCLUDED.title,description=EXCLUDED.description,difficulty=EXCLUDED.difficulty,estimated_minutes=EXCLUDED.estimated_minutes,sort_order=EXCLUDED.sort_order,category=EXCLUDED.category,warmup_enabled=EXCLUDED.warmup_enabled,is_default_warmup=EXCLUDED.is_default_warmup,status='published',published_at=COALESCE(workouts.published_at,NOW()),tenant_id=NULL,owner_user_id=NULL RETURNING id::text`, workout.Key, workout.Title, workout.Description, workout.Difficulty, workout.Minutes, order, workout.Category, workout.Category != "warmup", workout.Category == "warmup").Scan(&id)
		if err != nil {
			return fmt.Errorf("upsert %s: %w", workout.Key, err)
		}
		if _, err = tx.Exec(ctx, `DELETE FROM workout_exercises WHERE workout_id=$1::uuid`, id); err != nil {
			return err
		}
		for i, ex := range workout.Exercises {
			var exerciseID string
			err = tx.QueryRow(ctx, `SELECT id::text FROM exercises WHERE standard_key=$1 AND tenant_id IS NULL AND owner_user_id IS NULL AND status='published'`, ex.Key).Scan(&exerciseID)
			if err == pgx.ErrNoRows {
				return fmt.Errorf("system workout %s requires missing global published exercise %s", workout.Key, ex.Key)
			}
			if err != nil {
				return err
			}
			_, err = tx.Exec(ctx, `INSERT INTO workout_exercises(workout_id,exercise_id,sets,target_reps,target_duration_seconds,rest_seconds,sort_order) VALUES($1::uuid,$2::uuid,1,NULL,$3,$4,$5)`, id, exerciseID, ex.Duration, ex.Rest, i+1)
			if err != nil {
				return fmt.Errorf("insert %s/%s: %w", workout.Key, ex.Key, err)
			}
		}
	}
	return tx.Commit(ctx)
}
