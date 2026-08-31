CREATE TABLE tenants (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  slug text NOT NULL UNIQUE CHECK (slug ~ '^[a-z0-9]+(?:-[a-z0-9]+)*$' AND length(slug) BETWEEN 2 AND 63),
  name text NOT NULL CHECK (length(btrim(name)) BETWEEN 2 AND 120),
  owner_user_id uuid NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
  status text NOT NULL DEFAULT 'active' CHECK (status IN ('active','suspended','archived')),
  description text NOT NULL DEFAULT '',
  avatar_url text,
  created_at timestamptz NOT NULL DEFAULT NOW(),
  updated_at timestamptz NOT NULL DEFAULT NOW()
);
CREATE UNIQUE INDEX tenants_one_active_owner ON tenants(owner_user_id) WHERE status='active';
CREATE TRIGGER tenants_set_updated_at BEFORE UPDATE ON tenants FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TABLE tenant_memberships (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id uuid NOT NULL REFERENCES tenants(id) ON DELETE RESTRICT,
  user_id uuid NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
  role text NOT NULL CHECK (role IN ('coach','student')),
  status text NOT NULL DEFAULT 'active' CHECK (status IN ('active','inactive','blocked')),
  joined_at timestamptz NOT NULL DEFAULT NOW(),
  updated_at timestamptz NOT NULL DEFAULT NOW(),
  UNIQUE(tenant_id,user_id)
);
CREATE INDEX tenant_memberships_user_active ON tenant_memberships(user_id,tenant_id) WHERE status='active';
CREATE TRIGGER tenant_memberships_set_updated_at BEFORE UPDATE ON tenant_memberships FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- Existing owner-authored content is backfilled deterministically into one space per owner.
-- System content (owner_user_id IS NULL, notably standard exercises) deliberately remains global.
INSERT INTO tenants(slug,name,owner_user_id)
SELECT 'coach-'||left(md5(owner_id::text),16), COALESCE(NULLIF(p.display_name,''),'Coach space'), owner_id
FROM (
  SELECT owner_user_id owner_id FROM lessons WHERE owner_user_id IS NOT NULL UNION
  SELECT owner_user_id FROM exercises WHERE owner_user_id IS NOT NULL UNION
  SELECT owner_user_id FROM programs WHERE owner_user_id IS NOT NULL UNION
  SELECT owner_user_id FROM workouts WHERE owner_user_id IS NOT NULL UNION
  SELECT owner_user_id FROM skills WHERE owner_user_id IS NOT NULL UNION
  SELECT owner_user_id FROM media_assets WHERE owner_user_id IS NOT NULL UNION
  SELECT user_id FROM admin_users WHERE role='coach'
) owners JOIN profiles p ON p.user_id=owners.owner_id;
INSERT INTO tenant_memberships(tenant_id,user_id,role)
SELECT id,owner_user_id,'coach' FROM tenants ON CONFLICT(tenant_id,user_id) DO NOTHING;

ALTER TABLE lessons ADD COLUMN tenant_id uuid REFERENCES tenants(id) ON DELETE RESTRICT;
ALTER TABLE exercises ADD COLUMN tenant_id uuid REFERENCES tenants(id) ON DELETE RESTRICT;
ALTER TABLE programs ADD COLUMN tenant_id uuid REFERENCES tenants(id) ON DELETE RESTRICT;
ALTER TABLE workouts ADD COLUMN tenant_id uuid REFERENCES tenants(id) ON DELETE RESTRICT;
ALTER TABLE skills ADD COLUMN tenant_id uuid REFERENCES tenants(id) ON DELETE RESTRICT;
ALTER TABLE media_assets ADD COLUMN tenant_id uuid REFERENCES tenants(id) ON DELETE RESTRICT;
UPDATE lessons x SET tenant_id=t.id FROM tenants t WHERE x.owner_user_id=t.owner_user_id;
UPDATE exercises x SET tenant_id=t.id FROM tenants t WHERE x.owner_user_id=t.owner_user_id;
UPDATE programs x SET tenant_id=t.id FROM tenants t WHERE x.owner_user_id=t.owner_user_id;
UPDATE workouts x SET tenant_id=t.id FROM tenants t WHERE x.owner_user_id=t.owner_user_id;
UPDATE skills x SET tenant_id=t.id FROM tenants t WHERE x.owner_user_id=t.owner_user_id;
UPDATE media_assets x SET tenant_id=t.id FROM tenants t WHERE x.owner_user_id=t.owner_user_id;
-- Legacy seed/system exercises predate standard_key; their stable slug is the
-- deterministic global key. Owner-authored rows are never included here.
UPDATE exercises SET standard_key=slug WHERE owner_user_id IS NULL AND standard_key IS NULL;
ALTER TABLE exercises ADD CONSTRAINT exercises_global_or_tenant CHECK (
  (tenant_id IS NULL AND owner_user_id IS NULL AND standard_key IS NOT NULL) OR
  (tenant_id IS NOT NULL AND owner_user_id IS NOT NULL AND standard_key IS NULL)
);
ALTER TABLE lessons ADD CONSTRAINT lessons_tenant_owned CHECK (tenant_id IS NOT NULL OR owner_user_id IS NULL);
ALTER TABLE programs ADD CONSTRAINT programs_tenant_owned CHECK (tenant_id IS NOT NULL OR owner_user_id IS NULL);
ALTER TABLE workouts ADD CONSTRAINT workouts_tenant_owned CHECK (tenant_id IS NOT NULL OR owner_user_id IS NULL);
ALTER TABLE skills ADD CONSTRAINT skills_tenant_owned CHECK (tenant_id IS NOT NULL OR owner_user_id IS NULL);
ALTER TABLE media_assets ADD CONSTRAINT media_tenant_owned CHECK (tenant_id IS NOT NULL OR owner_user_id IS NULL);
CREATE INDEX lessons_tenant_status ON lessons(tenant_id,status);
CREATE INDEX exercises_tenant_status ON exercises(tenant_id,status);
CREATE INDEX programs_tenant_status ON programs(tenant_id,status);
CREATE INDEX workouts_tenant_status ON workouts(tenant_id,status);
CREATE INDEX skills_tenant_status ON skills(tenant_id,status);
CREATE INDEX media_assets_tenant_created ON media_assets(tenant_id,created_at DESC);

CREATE OR REPLACE FUNCTION assign_content_tenant() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
 IF NEW.owner_user_id IS NOT NULL AND NEW.tenant_id IS NULL THEN
   SELECT t.id INTO NEW.tenant_id FROM tenants t JOIN tenant_memberships m ON m.tenant_id=t.id
   WHERE m.user_id=NEW.owner_user_id AND m.role='coach' AND m.status='active' AND t.status='active'
   ORDER BY (t.owner_user_id=NEW.owner_user_id) DESC LIMIT 1;
 END IF;
 IF NEW.owner_user_id IS NOT NULL AND NOT EXISTS(SELECT 1 FROM tenant_memberships m WHERE m.tenant_id=NEW.tenant_id AND m.user_id=NEW.owner_user_id AND m.role='coach' AND m.status='active') THEN
   RAISE EXCEPTION 'content owner is not a coach in tenant';
 END IF;
 RETURN NEW;
END $$;
CREATE TRIGGER lessons_assign_tenant BEFORE INSERT OR UPDATE OF owner_user_id,tenant_id ON lessons FOR EACH ROW EXECUTE FUNCTION assign_content_tenant();
CREATE TRIGGER exercises_assign_tenant BEFORE INSERT OR UPDATE OF owner_user_id,tenant_id ON exercises FOR EACH ROW EXECUTE FUNCTION assign_content_tenant();
CREATE TRIGGER programs_assign_tenant BEFORE INSERT OR UPDATE OF owner_user_id,tenant_id ON programs FOR EACH ROW EXECUTE FUNCTION assign_content_tenant();
CREATE TRIGGER workouts_assign_tenant BEFORE INSERT OR UPDATE OF owner_user_id,tenant_id ON workouts FOR EACH ROW EXECUTE FUNCTION assign_content_tenant();
CREATE TRIGGER skills_assign_tenant BEFORE INSERT OR UPDATE OF owner_user_id,tenant_id ON skills FOR EACH ROW EXECUTE FUNCTION assign_content_tenant();
CREATE TRIGGER media_assign_tenant BEFORE INSERT OR UPDATE OF owner_user_id,tenant_id ON media_assets FOR EACH ROW EXECUTE FUNCTION assign_content_tenant();

ALTER TABLE workout_sessions ADD COLUMN tenant_id uuid REFERENCES tenants(id) ON DELETE RESTRICT;
UPDATE workout_sessions s SET tenant_id=w.tenant_id FROM workouts w WHERE w.id=s.workout_id;
ALTER TABLE user_program_progress ADD COLUMN tenant_id uuid REFERENCES tenants(id) ON DELETE RESTRICT;
UPDATE user_program_progress x SET tenant_id=p.tenant_id FROM programs p WHERE p.id=x.program_id;
ALTER TABLE user_skill_progress ADD COLUMN tenant_id uuid REFERENCES tenants(id) ON DELETE RESTRICT;
UPDATE user_skill_progress x SET tenant_id=s.tenant_id FROM skills s WHERE s.id=x.skill_id;
ALTER TABLE user_skill_level_progress ADD COLUMN tenant_id uuid REFERENCES tenants(id) ON DELETE RESTRICT;
UPDATE user_skill_level_progress x SET tenant_id=s.tenant_id FROM skill_levels l JOIN skills s ON s.id=l.skill_id WHERE l.id=x.skill_level_id;
ALTER TABLE user_skill_criteria ADD COLUMN tenant_id uuid REFERENCES tenants(id) ON DELETE RESTRICT;
UPDATE user_skill_criteria x SET tenant_id=s.tenant_id FROM skill_criteria c JOIN skills s ON s.id=c.skill_id WHERE c.id=x.criterion_id;
ALTER TABLE user_lesson_progress ADD COLUMN tenant_id uuid REFERENCES tenants(id) ON DELETE RESTRICT;
UPDATE user_lesson_progress x SET tenant_id=l.tenant_id FROM lessons l WHERE l.id=x.lesson_id;
ALTER TABLE user_training_schedules ADD COLUMN tenant_id uuid REFERENCES tenants(id) ON DELETE RESTRICT;
UPDATE user_training_schedules x SET tenant_id=w.tenant_id FROM workouts w WHERE w.id=x.workout_id;
ALTER TABLE user_planned_workouts ADD COLUMN tenant_id uuid REFERENCES tenants(id) ON DELETE RESTRICT;
UPDATE user_planned_workouts x SET tenant_id=w.tenant_id FROM workouts w WHERE w.id=x.workout_id;
ALTER TABLE user_exercise_stats ADD COLUMN tenant_id uuid REFERENCES tenants(id) ON DELETE RESTRICT;
UPDATE user_exercise_stats x SET tenant_id=e.tenant_id FROM exercises e WHERE e.id=x.exercise_id;
ALTER TABLE user_exercise_stats DROP CONSTRAINT user_exercise_stats_pkey;
CREATE UNIQUE INDEX user_exercise_stats_tenant_key ON user_exercise_stats(user_id,tenant_id,exercise_id) NULLS NOT DISTINCT;

-- Membership backfill follows actual tenant activity and is idempotent.
INSERT INTO tenant_memberships(tenant_id,user_id,role)
SELECT DISTINCT tenant_id,user_id,'student' FROM (
 SELECT tenant_id,user_id FROM workout_sessions WHERE tenant_id IS NOT NULL UNION
 SELECT tenant_id,user_id FROM user_program_progress WHERE tenant_id IS NOT NULL UNION
 SELECT tenant_id,user_id FROM user_skill_progress WHERE tenant_id IS NOT NULL UNION
 SELECT tenant_id,user_id FROM user_lesson_progress WHERE tenant_id IS NOT NULL UNION
 SELECT tenant_id,user_id FROM user_training_schedules WHERE tenant_id IS NOT NULL UNION
 SELECT tenant_id,user_id FROM user_planned_workouts WHERE tenant_id IS NOT NULL
) activity ON CONFLICT(tenant_id,user_id) DO NOTHING;

ALTER TABLE user_program_progress DROP CONSTRAINT user_program_progress_pkey;
ALTER TABLE user_program_progress ADD PRIMARY KEY(user_id,tenant_id,program_id);
CREATE UNIQUE INDEX user_program_progress_legacy_conflict ON user_program_progress(user_id,program_id);
CREATE INDEX workout_sessions_user_tenant_started ON workout_sessions(user_id,tenant_id,started_at DESC);

CREATE OR REPLACE FUNCTION enforce_tenant_references() RETURNS trigger LANGUAGE plpgsql AS $$
DECLARE parent_tenant uuid; child_tenant uuid;
BEGIN
 IF TG_TABLE_NAME='workout_exercises' THEN
   SELECT tenant_id INTO parent_tenant FROM workouts WHERE id=NEW.workout_id;
   SELECT tenant_id INTO child_tenant FROM exercises WHERE id=NEW.exercise_id;
   IF parent_tenant IS NULL OR (child_tenant IS NOT NULL AND child_tenant<>parent_tenant) THEN RAISE EXCEPTION 'cross-tenant workout exercise'; END IF;
 ELSIF TG_TABLE_NAME='workouts' THEN
   IF NEW.program_level_id IS NOT NULL THEN
     SELECT p.tenant_id INTO parent_tenant FROM program_levels pl JOIN programs p ON p.id=pl.program_id WHERE pl.id=NEW.program_level_id;
     IF NEW.tenant_id IS DISTINCT FROM parent_tenant THEN RAISE EXCEPTION 'cross-tenant program workout'; END IF;
   END IF;
 ELSIF TG_TABLE_NAME='skill_requirements' THEN
   SELECT tenant_id INTO parent_tenant FROM skills WHERE id=NEW.skill_id;
   SELECT tenant_id INTO child_tenant FROM skills WHERE id=NEW.required_skill_id;
   IF parent_tenant IS DISTINCT FROM child_tenant THEN RAISE EXCEPTION 'cross-tenant skill prerequisite'; END IF;
 ELSIF TG_TABLE_NAME='workout_sessions' THEN
   SELECT tenant_id INTO parent_tenant FROM workouts WHERE id=NEW.workout_id;
   IF NEW.tenant_id IS NULL OR NEW.tenant_id IS DISTINCT FROM parent_tenant THEN RAISE EXCEPTION 'cross-tenant workout session'; END IF;
 ELSIF TG_TABLE_NAME='user_training_schedules' OR TG_TABLE_NAME='user_planned_workouts' THEN
   SELECT tenant_id INTO parent_tenant FROM workouts WHERE id=NEW.workout_id;
   IF NEW.tenant_id IS NULL OR NEW.tenant_id IS DISTINCT FROM parent_tenant THEN RAISE EXCEPTION 'cross-tenant calendar workout'; END IF;
 ELSIF TG_TABLE_NAME='exercise_sets' THEN
   SELECT tenant_id INTO parent_tenant FROM workout_sessions WHERE id=NEW.session_id;
   SELECT tenant_id INTO child_tenant FROM exercises WHERE id=NEW.exercise_id;
   IF parent_tenant IS NULL OR (child_tenant IS NOT NULL AND child_tenant<>parent_tenant) OR NOT EXISTS(SELECT 1 FROM workout_sessions ws JOIN workout_exercises we ON we.workout_id=ws.workout_id WHERE ws.id=NEW.session_id AND we.exercise_id=NEW.exercise_id) THEN RAISE EXCEPTION 'cross-tenant exercise set'; END IF;
 ELSIF TG_TABLE_NAME='skill_levels' THEN
   IF NEW.program_level_id IS NOT NULL THEN
     SELECT s.tenant_id INTO parent_tenant FROM skills s WHERE s.id=NEW.skill_id;
     SELECT p.tenant_id INTO child_tenant FROM program_levels pl JOIN programs p ON p.id=pl.program_id WHERE pl.id=NEW.program_level_id;
     IF parent_tenant IS DISTINCT FROM child_tenant THEN RAISE EXCEPTION 'cross-tenant skill program'; END IF;
   END IF;
 END IF;
 RETURN NEW;
END $$;
CREATE TRIGGER workout_exercises_tenant_guard BEFORE INSERT OR UPDATE ON workout_exercises FOR EACH ROW EXECUTE FUNCTION enforce_tenant_references();
CREATE TRIGGER workouts_tenant_guard BEFORE INSERT OR UPDATE OF tenant_id,program_level_id ON workouts FOR EACH ROW EXECUTE FUNCTION enforce_tenant_references();
CREATE TRIGGER skill_requirements_tenant_guard BEFORE INSERT OR UPDATE ON skill_requirements FOR EACH ROW EXECUTE FUNCTION enforce_tenant_references();
CREATE TRIGGER workout_sessions_tenant_guard BEFORE INSERT OR UPDATE OF tenant_id,workout_id ON workout_sessions FOR EACH ROW EXECUTE FUNCTION enforce_tenant_references();
CREATE TRIGGER schedules_tenant_guard BEFORE INSERT OR UPDATE OF tenant_id,workout_id ON user_training_schedules FOR EACH ROW EXECUTE FUNCTION enforce_tenant_references();
CREATE TRIGGER planned_tenant_guard BEFORE INSERT OR UPDATE OF tenant_id,workout_id ON user_planned_workouts FOR EACH ROW EXECUTE FUNCTION enforce_tenant_references();
CREATE TRIGGER exercise_sets_tenant_guard BEFORE INSERT OR UPDATE OF session_id,exercise_id ON exercise_sets FOR EACH ROW EXECUTE FUNCTION enforce_tenant_references();
CREATE TRIGGER skill_levels_tenant_guard BEFORE INSERT OR UPDATE OF skill_id,program_level_id ON skill_levels FOR EACH ROW EXECUTE FUNCTION enforce_tenant_references();

CREATE OR REPLACE FUNCTION enforce_content_media_tenant() RETURNS trigger LANGUAGE plpgsql AS $$
DECLARE media_tenant uuid; content_tenant uuid; media_id uuid;
BEGIN
 content_tenant := (to_jsonb(NEW)->>'tenant_id')::uuid;
 media_id := (to_jsonb(NEW)->>'cover_media_id')::uuid;
 IF media_id IS NOT NULL THEN SELECT tenant_id INTO media_tenant FROM media_assets WHERE id=media_id; IF media_tenant IS DISTINCT FROM content_tenant THEN RAISE EXCEPTION 'cross-tenant cover media'; END IF; END IF;
 IF TG_TABLE_NAME='exercises' THEN media_id := (to_jsonb(NEW)->>'demo_media_id')::uuid; IF media_id IS NOT NULL THEN SELECT tenant_id INTO media_tenant FROM media_assets WHERE id=media_id; IF media_tenant IS DISTINCT FROM content_tenant THEN RAISE EXCEPTION 'cross-tenant demo media'; END IF; END IF; END IF;
 RETURN NEW;
END $$;
CREATE TRIGGER lessons_media_tenant_guard BEFORE INSERT OR UPDATE OF tenant_id,cover_media_id ON lessons FOR EACH ROW EXECUTE FUNCTION enforce_content_media_tenant();
CREATE TRIGGER exercises_media_tenant_guard BEFORE INSERT OR UPDATE OF tenant_id,cover_media_id,demo_media_id ON exercises FOR EACH ROW EXECUTE FUNCTION enforce_content_media_tenant();
CREATE TRIGGER workouts_media_tenant_guard BEFORE INSERT OR UPDATE OF tenant_id,cover_media_id ON workouts FOR EACH ROW EXECUTE FUNCTION enforce_content_media_tenant();
CREATE TRIGGER programs_media_tenant_guard BEFORE INSERT OR UPDATE OF tenant_id,cover_media_id ON programs FOR EACH ROW EXECUTE FUNCTION enforce_content_media_tenant();
CREATE TRIGGER skills_media_tenant_guard BEFORE INSERT OR UPDATE OF tenant_id,cover_media_id ON skills FOR EACH ROW EXECUTE FUNCTION enforce_content_media_tenant();

-- coach is now a tenant membership role, never a platform privilege.
DELETE FROM admin_users WHERE role='coach';
ALTER TABLE admin_users DROP CONSTRAINT admin_users_role_check;
ALTER TABLE admin_users ADD CONSTRAINT admin_users_role_check CHECK (role IN ('admin','super_admin'));
