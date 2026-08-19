DROP TRIGGER IF EXISTS user_planned_workouts_set_updated_at ON user_planned_workouts;
DROP TRIGGER IF EXISTS user_training_schedules_set_updated_at ON user_training_schedules;
ALTER TABLE workout_sessions DROP COLUMN IF EXISTS planned_workout_id;
DROP TABLE IF EXISTS user_planned_workouts;
DROP TABLE IF EXISTS user_training_schedule_days;
DROP TABLE IF EXISTS user_training_schedules;
