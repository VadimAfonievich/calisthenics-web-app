DROP TRIGGER IF EXISTS tenants_avatar_media_guard ON tenants;
DROP FUNCTION IF EXISTS enforce_tenant_avatar_media();
DROP INDEX IF EXISTS workouts_standard_key_unique;
ALTER TABLE workouts DROP CONSTRAINT IF EXISTS workouts_standard_key_system_owned;
ALTER TABLE workouts DROP CONSTRAINT IF EXISTS workouts_standard_key_format;
ALTER TABLE workouts DROP COLUMN standard_key;
ALTER TABLE tenants DROP COLUMN avatar_media_id;

CREATE OR REPLACE FUNCTION enforce_tenant_references() RETURNS trigger LANGUAGE plpgsql AS $$
DECLARE parent_tenant uuid; child_tenant uuid;
BEGIN
 IF TG_TABLE_NAME='workout_exercises' THEN
   SELECT tenant_id INTO parent_tenant FROM workouts WHERE id=NEW.workout_id;
   SELECT tenant_id INTO child_tenant FROM exercises WHERE id=NEW.exercise_id;
   IF parent_tenant IS NULL OR (child_tenant IS NOT NULL AND child_tenant<>parent_tenant) THEN RAISE EXCEPTION 'cross-tenant workout exercise'; END IF;
 ELSIF TG_TABLE_NAME='workouts' THEN
   IF NEW.program_level_id IS NOT NULL THEN SELECT p.tenant_id INTO parent_tenant FROM program_levels pl JOIN programs p ON p.id=pl.program_id WHERE pl.id=NEW.program_level_id; IF NEW.tenant_id IS DISTINCT FROM parent_tenant THEN RAISE EXCEPTION 'cross-tenant program workout'; END IF; END IF;
 ELSIF TG_TABLE_NAME='skill_requirements' THEN SELECT tenant_id INTO parent_tenant FROM skills WHERE id=NEW.skill_id; SELECT tenant_id INTO child_tenant FROM skills WHERE id=NEW.required_skill_id; IF parent_tenant IS DISTINCT FROM child_tenant THEN RAISE EXCEPTION 'cross-tenant skill prerequisite'; END IF;
 ELSIF TG_TABLE_NAME='workout_sessions' THEN SELECT tenant_id INTO parent_tenant FROM workouts WHERE id=NEW.workout_id; IF NEW.tenant_id IS NULL OR NEW.tenant_id IS DISTINCT FROM parent_tenant THEN RAISE EXCEPTION 'cross-tenant workout session'; END IF;
 ELSIF TG_TABLE_NAME='user_training_schedules' OR TG_TABLE_NAME='user_planned_workouts' THEN SELECT tenant_id INTO parent_tenant FROM workouts WHERE id=NEW.workout_id; IF NEW.tenant_id IS NULL OR NEW.tenant_id IS DISTINCT FROM parent_tenant THEN RAISE EXCEPTION 'cross-tenant calendar workout'; END IF;
 ELSIF TG_TABLE_NAME='exercise_sets' THEN SELECT tenant_id INTO parent_tenant FROM workout_sessions WHERE id=NEW.session_id; SELECT tenant_id INTO child_tenant FROM exercises WHERE id=NEW.exercise_id; IF parent_tenant IS NULL OR (child_tenant IS NOT NULL AND child_tenant<>parent_tenant) OR NOT EXISTS(SELECT 1 FROM workout_sessions ws JOIN workout_exercises we ON we.workout_id=ws.workout_id WHERE ws.id=NEW.session_id AND we.exercise_id=NEW.exercise_id) THEN RAISE EXCEPTION 'cross-tenant exercise set'; END IF;
 ELSIF TG_TABLE_NAME='skill_levels' THEN IF NEW.program_level_id IS NOT NULL THEN SELECT s.tenant_id INTO parent_tenant FROM skills s WHERE s.id=NEW.skill_id; SELECT p.tenant_id INTO child_tenant FROM program_levels pl JOIN programs p ON p.id=pl.program_id WHERE pl.id=NEW.program_level_id; IF parent_tenant IS DISTINCT FROM child_tenant THEN RAISE EXCEPTION 'cross-tenant skill program'; END IF; END IF;
 END IF;
 RETURN NEW;
END $$;
