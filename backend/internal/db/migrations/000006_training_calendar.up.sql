CREATE TABLE user_training_schedules (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id uuid NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
  workout_id uuid NOT NULL REFERENCES workouts(id) ON DELETE RESTRICT,
  preferred_time time,
  timezone text NOT NULL CHECK (timezone <> ''),
  active boolean NOT NULL DEFAULT true,
  start_date date NOT NULL,
  end_date date,
  created_at timestamptz NOT NULL DEFAULT NOW(),
  updated_at timestamptz NOT NULL DEFAULT NOW(),
  CHECK (end_date IS NULL OR end_date >= start_date)
);

CREATE TABLE user_training_schedule_days (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  schedule_id uuid NOT NULL REFERENCES user_training_schedules(id) ON DELETE CASCADE,
  weekday smallint NOT NULL CHECK (weekday BETWEEN 1 AND 7),
  UNIQUE (schedule_id, weekday)
);

CREATE TABLE user_planned_workouts (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id uuid NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
  workout_id uuid NOT NULL REFERENCES workouts(id) ON DELETE RESTRICT,
  scheduled_date date NOT NULL,
  scheduled_time time,
  timezone text NOT NULL CHECK (timezone <> ''),
  source_schedule_id uuid REFERENCES user_training_schedules(id) ON DELETE SET NULL,
  status text NOT NULL DEFAULT 'scheduled' CHECK (status IN ('scheduled','completed','skipped','cancelled')),
  created_at timestamptz NOT NULL DEFAULT NOW(),
  updated_at timestamptz NOT NULL DEFAULT NOW()
);

ALTER TABLE workout_sessions ADD COLUMN planned_workout_id uuid REFERENCES user_planned_workouts(id) ON DELETE SET NULL;
CREATE UNIQUE INDEX idx_workout_sessions_planned_workout ON workout_sessions(planned_workout_id) WHERE planned_workout_id IS NOT NULL;
CREATE INDEX idx_training_schedules_user ON user_training_schedules(user_id);
CREATE INDEX idx_training_schedules_workout ON user_training_schedules(workout_id);
CREATE INDEX idx_training_schedule_days_schedule ON user_training_schedule_days(schedule_id);
CREATE INDEX idx_planned_workouts_user_date ON user_planned_workouts(user_id, scheduled_date);
CREATE INDEX idx_planned_workouts_workout ON user_planned_workouts(workout_id);
CREATE INDEX idx_planned_workouts_status ON user_planned_workouts(status);
CREATE UNIQUE INDEX idx_planned_workouts_schedule_date ON user_planned_workouts(user_id,source_schedule_id,scheduled_date) WHERE source_schedule_id IS NOT NULL;
CREATE TRIGGER user_training_schedules_set_updated_at BEFORE UPDATE ON user_training_schedules FOR EACH ROW EXECUTE FUNCTION set_updated_at();
CREATE TRIGGER user_planned_workouts_set_updated_at BEFORE UPDATE ON user_planned_workouts FOR EACH ROW EXECUTE FUNCTION set_updated_at();
