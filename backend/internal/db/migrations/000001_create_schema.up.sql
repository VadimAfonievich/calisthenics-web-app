CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS trigger
LANGUAGE plpgsql
AS $$
BEGIN
  NEW.updated_at = NOW();
  RETURN NEW;
END;
$$;

CREATE TABLE users (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  telegram_id bigint NOT NULL UNIQUE CHECK (telegram_id > 0),
  username text,
  first_name text NOT NULL,
  last_name text,
  photo_url text,
  created_at timestamptz NOT NULL DEFAULT NOW(),
  updated_at timestamptz NOT NULL DEFAULT NOW()
);

CREATE TABLE profiles (
  user_id uuid PRIMARY KEY REFERENCES users(id) ON DELETE RESTRICT,
  display_name text NOT NULL,
  level integer NOT NULL DEFAULT 1 CHECK (level >= 1),
  xp integer NOT NULL DEFAULT 0 CHECK (xp >= 0),
  current_streak integer NOT NULL DEFAULT 0 CHECK (current_streak >= 0),
  longest_streak integer NOT NULL DEFAULT 0 CHECK (longest_streak >= 0),
  preferred_difficulty text NOT NULL DEFAULT 'beginner' CHECK (preferred_difficulty IN ('beginner', 'intermediate', 'advanced')),
  timezone text NOT NULL DEFAULT 'UTC' CHECK (timezone ~ '^[A-Za-z_]+/[A-Za-z_]+(?:/[A-Za-z_]+)?$|^UTC$'),
  created_at timestamptz NOT NULL DEFAULT NOW(),
  updated_at timestamptz NOT NULL DEFAULT NOW()
);

CREATE TABLE lesson_categories (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  name text NOT NULL,
  slug text NOT NULL UNIQUE,
  sort_order integer NOT NULL DEFAULT 0 CHECK (sort_order >= 0),
  created_at timestamptz NOT NULL DEFAULT NOW(),
  updated_at timestamptz NOT NULL DEFAULT NOW()
);

CREATE TABLE lessons (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  category_id uuid NOT NULL REFERENCES lesson_categories(id) ON DELETE RESTRICT,
  title text NOT NULL,
  slug text NOT NULL UNIQUE,
  short_description text NOT NULL,
  content text NOT NULL,
  video_url text,
  image_url text,
  difficulty text NOT NULL CHECK (difficulty IN ('beginner', 'intermediate', 'advanced')),
  duration_minutes integer NOT NULL CHECK (duration_minutes > 0),
  sort_order integer NOT NULL DEFAULT 0 CHECK (sort_order >= 0),
  published boolean NOT NULL DEFAULT false,
  created_at timestamptz NOT NULL DEFAULT NOW(),
  updated_at timestamptz NOT NULL DEFAULT NOW()
);

CREATE TABLE user_lesson_progress (
  user_id uuid NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
  lesson_id uuid NOT NULL REFERENCES lessons(id) ON DELETE RESTRICT,
  completed boolean NOT NULL DEFAULT false,
  progress_percent smallint NOT NULL DEFAULT 0 CHECK (progress_percent BETWEEN 0 AND 100),
  completed_at timestamptz,
  created_at timestamptz NOT NULL DEFAULT NOW(),
  updated_at timestamptz NOT NULL DEFAULT NOW(),
  PRIMARY KEY (user_id, lesson_id),
  CHECK ((completed AND progress_percent = 100 AND completed_at IS NOT NULL) OR (NOT completed AND completed_at IS NULL))
);

CREATE TABLE exercises (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  name text NOT NULL,
  slug text NOT NULL UNIQUE,
  description text NOT NULL,
  instructions text NOT NULL,
  common_mistakes text NOT NULL,
  difficulty text NOT NULL CHECK (difficulty IN ('beginner', 'intermediate', 'advanced')),
  muscle_groups text[] NOT NULL DEFAULT '{}'::text[] CHECK (cardinality(muscle_groups) > 0),
  equipment text[] NOT NULL DEFAULT '{}'::text[],
  video_url text,
  image_url text,
  created_at timestamptz NOT NULL DEFAULT NOW(),
  updated_at timestamptz NOT NULL DEFAULT NOW()
);

CREATE TABLE programs (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  name text NOT NULL,
  slug text NOT NULL UNIQUE,
  description text NOT NULL,
  difficulty text NOT NULL CHECK (difficulty IN ('beginner', 'intermediate', 'advanced')),
  duration_weeks integer NOT NULL CHECK (duration_weeks > 0),
  published boolean NOT NULL DEFAULT false,
  created_at timestamptz NOT NULL DEFAULT NOW(),
  updated_at timestamptz NOT NULL DEFAULT NOW()
);

CREATE TABLE workouts (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  program_id uuid NOT NULL REFERENCES programs(id) ON DELETE RESTRICT,
  day_number integer NOT NULL CHECK (day_number > 0),
  title text NOT NULL,
  description text NOT NULL,
  estimated_minutes integer NOT NULL CHECK (estimated_minutes > 0),
  sort_order integer NOT NULL DEFAULT 0 CHECK (sort_order >= 0),
  created_at timestamptz NOT NULL DEFAULT NOW(),
  updated_at timestamptz NOT NULL DEFAULT NOW(),
  UNIQUE (program_id, day_number)
);

CREATE TABLE workout_exercises (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  workout_id uuid NOT NULL REFERENCES workouts(id) ON DELETE RESTRICT,
  exercise_id uuid NOT NULL REFERENCES exercises(id) ON DELETE RESTRICT,
  sets smallint NOT NULL CHECK (sets > 0),
  target_reps smallint,
  target_duration_seconds integer,
  rest_seconds integer NOT NULL DEFAULT 60 CHECK (rest_seconds >= 0),
  sort_order integer NOT NULL DEFAULT 0 CHECK (sort_order >= 0),
  notes text,
  UNIQUE (workout_id, exercise_id),
  CHECK ((target_reps IS NOT NULL AND target_reps > 0 AND target_duration_seconds IS NULL) OR (target_reps IS NULL AND target_duration_seconds IS NOT NULL AND target_duration_seconds > 0))
);

CREATE TABLE workout_sessions (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id uuid NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
  workout_id uuid NOT NULL REFERENCES workouts(id) ON DELETE RESTRICT,
  started_at timestamptz NOT NULL DEFAULT NOW(),
  completed_at timestamptz,
  status text NOT NULL DEFAULT 'started' CHECK (status IN ('started', 'completed', 'cancelled')),
  duration_seconds integer NOT NULL DEFAULT 0 CHECK (duration_seconds >= 0),
  xp_earned integer NOT NULL DEFAULT 0 CHECK (xp_earned >= 0),
  created_at timestamptz NOT NULL DEFAULT NOW(),
  updated_at timestamptz NOT NULL DEFAULT NOW(),
  CHECK ((status = 'completed' AND completed_at IS NOT NULL) OR (status IN ('started', 'cancelled') AND completed_at IS NULL))
);

CREATE TABLE exercise_sets (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  session_id uuid NOT NULL REFERENCES workout_sessions(id) ON DELETE RESTRICT,
  exercise_id uuid NOT NULL REFERENCES exercises(id) ON DELETE RESTRICT,
  set_number smallint NOT NULL CHECK (set_number > 0),
  reps smallint,
  duration_seconds integer,
  weight numeric(7,2),
  completed boolean NOT NULL DEFAULT false,
  created_at timestamptz NOT NULL DEFAULT NOW(),
  updated_at timestamptz NOT NULL DEFAULT NOW(),
  UNIQUE (session_id, exercise_id, set_number),
  CHECK (reps IS NULL OR reps >= 0),
  CHECK (duration_seconds IS NULL OR duration_seconds >= 0),
  CHECK (weight IS NULL OR weight >= 0),
  CHECK ((reps IS NOT NULL AND duration_seconds IS NULL) OR (reps IS NULL AND duration_seconds IS NOT NULL))
);

CREATE TABLE user_progress (
  user_id uuid PRIMARY KEY REFERENCES users(id) ON DELETE RESTRICT,
  total_workouts integer NOT NULL DEFAULT 0 CHECK (total_workouts >= 0),
  total_completed_exercises integer NOT NULL DEFAULT 0 CHECK (total_completed_exercises >= 0),
  total_training_seconds bigint NOT NULL DEFAULT 0 CHECK (total_training_seconds >= 0),
  updated_at timestamptz NOT NULL DEFAULT NOW()
);

CREATE TABLE user_exercise_stats (
  user_id uuid NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
  exercise_id uuid NOT NULL REFERENCES exercises(id) ON DELETE RESTRICT,
  total_sets integer NOT NULL DEFAULT 0 CHECK (total_sets >= 0),
  total_reps integer NOT NULL DEFAULT 0 CHECK (total_reps >= 0),
  max_reps integer NOT NULL DEFAULT 0 CHECK (max_reps >= 0),
  total_duration_seconds bigint NOT NULL DEFAULT 0 CHECK (total_duration_seconds >= 0),
  max_duration_seconds integer NOT NULL DEFAULT 0 CHECK (max_duration_seconds >= 0),
  last_performed_at timestamptz,
  updated_at timestamptz NOT NULL DEFAULT NOW(),
  PRIMARY KEY (user_id, exercise_id)
);

CREATE TABLE achievements (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  code text NOT NULL UNIQUE,
  title text NOT NULL,
  description text NOT NULL,
  icon text NOT NULL,
  xp_reward integer NOT NULL DEFAULT 0 CHECK (xp_reward >= 0),
  condition_type text NOT NULL CHECK (condition_type IN ('workouts_completed', 'streak_days', 'exercise_completed', 'exercise_max_reps')),
  condition_value integer NOT NULL CHECK (condition_value > 0),
  created_at timestamptz NOT NULL DEFAULT NOW()
);

CREATE TABLE user_achievements (
  user_id uuid NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
  achievement_id uuid NOT NULL REFERENCES achievements(id) ON DELETE RESTRICT,
  unlocked_at timestamptz NOT NULL DEFAULT NOW(),
  PRIMARY KEY (user_id, achievement_id)
);

CREATE TABLE admin_users (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id uuid NOT NULL UNIQUE REFERENCES users(id) ON DELETE RESTRICT,
  role text NOT NULL CHECK (role IN ('admin', 'super_admin')),
  created_at timestamptz NOT NULL DEFAULT NOW(),
  updated_at timestamptz NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_lessons_category_id ON lessons(category_id);
CREATE INDEX idx_lessons_published ON lessons(published) WHERE published;
CREATE INDEX idx_user_lesson_progress_user_id ON user_lesson_progress(user_id);
CREATE INDEX idx_user_lesson_progress_lesson_id ON user_lesson_progress(lesson_id);
CREATE INDEX idx_workouts_program_id ON workouts(program_id);
CREATE INDEX idx_workout_exercises_workout_id ON workout_exercises(workout_id);
CREATE INDEX idx_workout_sessions_user_started_at ON workout_sessions(user_id, started_at DESC);
CREATE INDEX idx_workout_sessions_workout_id ON workout_sessions(workout_id);
CREATE INDEX idx_workout_sessions_started_at ON workout_sessions(started_at DESC);
CREATE INDEX idx_exercise_sets_session_id ON exercise_sets(session_id);
CREATE INDEX idx_user_achievements_user_id ON user_achievements(user_id);

CREATE TRIGGER users_set_updated_at BEFORE UPDATE ON users FOR EACH ROW EXECUTE FUNCTION set_updated_at();
CREATE TRIGGER profiles_set_updated_at BEFORE UPDATE ON profiles FOR EACH ROW EXECUTE FUNCTION set_updated_at();
CREATE TRIGGER lesson_categories_set_updated_at BEFORE UPDATE ON lesson_categories FOR EACH ROW EXECUTE FUNCTION set_updated_at();
CREATE TRIGGER lessons_set_updated_at BEFORE UPDATE ON lessons FOR EACH ROW EXECUTE FUNCTION set_updated_at();
CREATE TRIGGER user_lesson_progress_set_updated_at BEFORE UPDATE ON user_lesson_progress FOR EACH ROW EXECUTE FUNCTION set_updated_at();
CREATE TRIGGER exercises_set_updated_at BEFORE UPDATE ON exercises FOR EACH ROW EXECUTE FUNCTION set_updated_at();
CREATE TRIGGER programs_set_updated_at BEFORE UPDATE ON programs FOR EACH ROW EXECUTE FUNCTION set_updated_at();
CREATE TRIGGER workouts_set_updated_at BEFORE UPDATE ON workouts FOR EACH ROW EXECUTE FUNCTION set_updated_at();
CREATE TRIGGER workout_sessions_set_updated_at BEFORE UPDATE ON workout_sessions FOR EACH ROW EXECUTE FUNCTION set_updated_at();
CREATE TRIGGER exercise_sets_set_updated_at BEFORE UPDATE ON exercise_sets FOR EACH ROW EXECUTE FUNCTION set_updated_at();
CREATE TRIGGER user_progress_set_updated_at BEFORE UPDATE ON user_progress FOR EACH ROW EXECUTE FUNCTION set_updated_at();
CREATE TRIGGER user_exercise_stats_set_updated_at BEFORE UPDATE ON user_exercise_stats FOR EACH ROW EXECUTE FUNCTION set_updated_at();
CREATE TRIGGER admin_users_set_updated_at BEFORE UPDATE ON admin_users FOR EACH ROW EXECUTE FUNCTION set_updated_at();
