ALTER TABLE tenants ADD COLUMN avatar_media_id uuid REFERENCES media_assets(id) ON DELETE SET NULL;
ALTER TABLE workouts ADD COLUMN standard_key text;
ALTER TABLE workouts ADD CONSTRAINT workouts_standard_key_format CHECK (standard_key IS NULL OR standard_key ~ '^[a-z0-9]+(-[a-z0-9]+)*$');
ALTER TABLE workouts ADD CONSTRAINT workouts_standard_key_system_owned CHECK (standard_key IS NULL OR (tenant_id IS NULL AND owner_user_id IS NULL));
CREATE UNIQUE INDEX workouts_standard_key_unique ON workouts(standard_key) WHERE standard_key IS NOT NULL;

UPDATE workouts SET standard_key='system-warmup-standard',title='Стандартная разминка'
WHERE id=(SELECT id FROM workouts WHERE is_default_warmup AND tenant_id IS NULL AND owner_user_id IS NULL ORDER BY id LIMIT 1);

CREATE OR REPLACE FUNCTION enforce_tenant_references() RETURNS trigger LANGUAGE plpgsql AS $$
DECLARE parent_tenant uuid; child_tenant uuid; parent_standard text;
BEGIN
 IF TG_TABLE_NAME='workout_exercises' THEN
   SELECT tenant_id,standard_key INTO parent_tenant,parent_standard FROM workouts WHERE id=NEW.workout_id;
   SELECT tenant_id INTO child_tenant FROM exercises WHERE id=NEW.exercise_id;
   IF (parent_standard IS NOT NULL AND child_tenant IS NOT NULL) OR (parent_tenant IS NOT NULL AND child_tenant IS NOT NULL AND child_tenant<>parent_tenant) THEN RAISE EXCEPTION 'cross-tenant workout exercise'; END IF;
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
   SELECT tenant_id,standard_key INTO parent_tenant,parent_standard FROM workouts WHERE id=NEW.workout_id;
   IF NEW.tenant_id IS NULL OR (parent_standard IS NULL AND NEW.tenant_id IS DISTINCT FROM parent_tenant) THEN RAISE EXCEPTION 'cross-tenant workout session'; END IF;
 ELSIF TG_TABLE_NAME='user_training_schedules' OR TG_TABLE_NAME='user_planned_workouts' THEN
   SELECT tenant_id,standard_key INTO parent_tenant,parent_standard FROM workouts WHERE id=NEW.workout_id;
   IF NEW.tenant_id IS NULL OR (parent_standard IS NULL AND NEW.tenant_id IS DISTINCT FROM parent_tenant) THEN RAISE EXCEPTION 'cross-tenant calendar workout'; END IF;
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

CREATE OR REPLACE FUNCTION enforce_tenant_avatar_media() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
 IF NEW.avatar_media_id IS NOT NULL AND NOT EXISTS(SELECT 1 FROM media_assets m WHERE m.id=NEW.avatar_media_id AND m.tenant_id=NEW.id AND m.type='image' AND m.mime_type IN ('image/jpeg','image/png','image/webp') AND m.status='ready') THEN
   RAISE EXCEPTION 'cross-tenant or invalid tenant avatar media';
 END IF;
 RETURN NEW;
END $$;
CREATE TRIGGER tenants_avatar_media_guard BEFORE INSERT OR UPDATE OF avatar_media_id ON tenants FOR EACH ROW EXECUTE FUNCTION enforce_tenant_avatar_media();
