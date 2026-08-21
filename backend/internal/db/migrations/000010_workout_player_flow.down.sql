DROP INDEX idx_workout_sessions_follow_up;
DROP INDEX idx_workouts_warmup_workout;
ALTER TABLE workout_sessions DROP COLUMN continued_session_id;
ALTER TABLE workout_sessions DROP COLUMN follow_up_planned_workout_id;
ALTER TABLE workout_sessions DROP COLUMN follow_up_workout_id;
ALTER TABLE workout_sessions DROP COLUMN session_purpose;
ALTER TABLE workouts DROP COLUMN warmup_workout_id;
