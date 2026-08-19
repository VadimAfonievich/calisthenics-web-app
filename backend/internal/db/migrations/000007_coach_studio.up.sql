ALTER TABLE admin_users DROP CONSTRAINT admin_users_role_check;
ALTER TABLE admin_users ADD CONSTRAINT admin_users_role_check CHECK (role IN ('coach','admin','super_admin'));

CREATE TABLE media_assets (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  owner_user_id uuid REFERENCES users(id) ON DELETE RESTRICT,
  type text NOT NULL CHECK (type IN ('image','video')),
  status text NOT NULL DEFAULT 'ready' CHECK (status IN ('pending','ready','failed','archived')),
  storage_provider text NOT NULL,
  storage_key text NOT NULL UNIQUE,
  url text NOT NULL,
  thumbnail_url text,
  original_filename text NOT NULL,
  mime_type text NOT NULL,
  size_bytes bigint NOT NULL CHECK (size_bytes >= 0),
  width integer CHECK (width IS NULL OR width > 0),
  height integer CHECK (height IS NULL OR height > 0),
  duration_seconds integer CHECK (duration_seconds IS NULL OR duration_seconds >= 0),
  created_at timestamptz NOT NULL DEFAULT NOW(),
  updated_at timestamptz NOT NULL DEFAULT NOW()
);

ALTER TABLE lessons ADD COLUMN owner_user_id uuid REFERENCES users(id) ON DELETE RESTRICT,
  ADD COLUMN status text NOT NULL DEFAULT 'draft' CHECK (status IN ('draft','published','archived')),
  ADD COLUMN content_blocks jsonb NOT NULL DEFAULT '[]'::jsonb CHECK (jsonb_typeof(content_blocks)='array'),
  ADD COLUMN cover_media_id uuid REFERENCES media_assets(id) ON DELETE SET NULL,
  ADD COLUMN published_at timestamptz,
  ADD COLUMN published_by uuid REFERENCES users(id) ON DELETE SET NULL;
UPDATE lessons SET status=CASE WHEN published THEN 'published' ELSE 'draft' END,published_at=CASE WHEN published THEN updated_at END;

ALTER TABLE exercises ADD COLUMN owner_user_id uuid REFERENCES users(id) ON DELETE RESTRICT,
  ADD COLUMN status text NOT NULL DEFAULT 'published' CHECK (status IN ('draft','published','archived')),
  ADD COLUMN movement_type text NOT NULL DEFAULT 'reps' CHECK (movement_type IN ('reps','duration','distance','custom')),
  ADD COLUMN coach_tips text NOT NULL DEFAULT '',
  ADD COLUMN cover_media_id uuid REFERENCES media_assets(id) ON DELETE SET NULL,
  ADD COLUMN published_at timestamptz DEFAULT NOW(),
  ADD COLUMN published_by uuid REFERENCES users(id) ON DELETE SET NULL;

ALTER TABLE programs ADD COLUMN owner_user_id uuid REFERENCES users(id) ON DELETE RESTRICT,
  ADD COLUMN status text NOT NULL DEFAULT 'draft' CHECK (status IN ('draft','published','archived')),
  ADD COLUMN cover_media_id uuid REFERENCES media_assets(id) ON DELETE SET NULL,
  ADD COLUMN coach_description text NOT NULL DEFAULT '',
  ADD COLUMN published_at timestamptz,
  ADD COLUMN published_by uuid REFERENCES users(id) ON DELETE SET NULL;
UPDATE programs SET status=CASE WHEN published THEN 'published' ELSE 'draft' END,published_at=CASE WHEN published THEN updated_at END;

ALTER TABLE workouts ADD COLUMN owner_user_id uuid REFERENCES users(id) ON DELETE RESTRICT,
  ADD COLUMN status text NOT NULL DEFAULT 'published' CHECK (status IN ('draft','published','archived')),
  ADD COLUMN cover_media_id uuid REFERENCES media_assets(id) ON DELETE SET NULL,
  ADD COLUMN published_at timestamptz DEFAULT NOW(),
  ADD COLUMN published_by uuid REFERENCES users(id) ON DELETE SET NULL;

ALTER TABLE skills ADD COLUMN owner_user_id uuid REFERENCES users(id) ON DELETE RESTRICT,
  ADD COLUMN status text NOT NULL DEFAULT 'published' CHECK (status IN ('draft','published','archived')),
  ADD COLUMN cover_media_id uuid REFERENCES media_assets(id) ON DELETE SET NULL,
  ADD COLUMN published_at timestamptz DEFAULT NOW(),
  ADD COLUMN published_by uuid REFERENCES users(id) ON DELETE SET NULL;

ALTER TABLE skill_levels DROP CONSTRAINT skill_levels_criterion_type_check;
ALTER TABLE skill_levels ADD CONSTRAINT skill_levels_criterion_type_check CHECK (criterion_type IN ('workout_completed','duration_seconds','repetitions','manual_confirmation','workout_count','exercise_reps','exercise_duration','skill_hold_duration','manual_user_confirmation','manual_coach_confirmation'));
ALTER TABLE skills DROP CONSTRAINT skills_final_criterion_type_check;
ALTER TABLE skills ADD CONSTRAINT skills_final_criterion_type_check CHECK (final_criterion_type IN ('duration_seconds','repetitions','manual_confirmation','workout_count','exercise_reps','exercise_duration','skill_hold_duration','manual_user_confirmation','manual_coach_confirmation'));

CREATE OR REPLACE FUNCTION sync_content_lifecycle() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
  IF TG_OP='INSERT' AND NEW.published THEN NEW.status='published'; END IF;
  IF TG_OP='UPDATE' AND NEW.published IS DISTINCT FROM OLD.published THEN NEW.status=CASE WHEN NEW.published THEN 'published' ELSE 'draft' END; END IF;
  IF TG_OP='UPDATE' AND NEW.status IS DISTINCT FROM OLD.status THEN NEW.published=(NEW.status='published'); END IF;
  IF NEW.status='published' AND NEW.published_at IS NULL THEN NEW.published_at=NOW(); END IF;
  RETURN NEW;
END $$;
CREATE TRIGGER lessons_sync_lifecycle BEFORE INSERT OR UPDATE ON lessons FOR EACH ROW EXECUTE FUNCTION sync_content_lifecycle();
CREATE TRIGGER programs_sync_lifecycle BEFORE INSERT OR UPDATE ON programs FOR EACH ROW EXECUTE FUNCTION sync_content_lifecycle();
CREATE TRIGGER media_assets_set_updated_at BEFORE UPDATE ON media_assets FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE INDEX idx_lessons_owner_status ON lessons(owner_user_id,status);
CREATE INDEX idx_exercises_owner_status ON exercises(owner_user_id,status);
CREATE INDEX idx_programs_owner_status ON programs(owner_user_id,status);
CREATE INDEX idx_workouts_owner_status ON workouts(owner_user_id,status);
CREATE INDEX idx_skills_owner_status ON skills(owner_user_id,status);
CREATE INDEX idx_media_assets_owner_created ON media_assets(owner_user_id,created_at DESC);
