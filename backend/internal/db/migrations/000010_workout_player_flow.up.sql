ALTER TABLE workouts ADD COLUMN warmup_workout_id uuid REFERENCES workouts(id) ON DELETE SET NULL;

ALTER TABLE workout_sessions ADD COLUMN session_purpose text NOT NULL DEFAULT 'main'
  CHECK (session_purpose IN ('main','warmup'));
ALTER TABLE workout_sessions ADD COLUMN follow_up_workout_id uuid REFERENCES workouts(id) ON DELETE SET NULL;
ALTER TABLE workout_sessions ADD COLUMN follow_up_planned_workout_id uuid REFERENCES user_planned_workouts(id) ON DELETE SET NULL;
ALTER TABLE workout_sessions ADD COLUMN continued_session_id uuid REFERENCES workout_sessions(id) ON DELETE SET NULL;

CREATE INDEX idx_workouts_warmup_workout ON workouts(warmup_workout_id);
CREATE INDEX idx_workout_sessions_follow_up ON workout_sessions(user_id,follow_up_workout_id)
  WHERE follow_up_workout_id IS NOT NULL;
