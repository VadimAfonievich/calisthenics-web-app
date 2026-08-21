-- The old schema cannot represent standalone workouts. Refuse rollback until
-- they are assigned through Program Editor instead of deleting user content.
DO $$ BEGIN
  IF EXISTS(SELECT 1 FROM workouts WHERE (program_id IS NULL OR day_number IS NULL) AND id<>'59000000-0000-0000-0000-000000000001') THEN
    RAISE EXCEPTION 'assign standalone workouts to programs before rolling back migration 000009';
  END IF;
  IF EXISTS(SELECT 1 FROM workout_sessions WHERE workout_id='59000000-0000-0000-0000-000000000001')
    OR EXISTS(SELECT 1 FROM user_training_schedules WHERE workout_id='59000000-0000-0000-0000-000000000001')
    OR EXISTS(SELECT 1 FROM user_planned_workouts WHERE workout_id='59000000-0000-0000-0000-000000000001') THEN
    RAISE EXCEPTION 'standard warmup has user history and cannot be removed safely';
  END IF;
END $$;
DELETE FROM workout_exercises WHERE workout_id='59000000-0000-0000-0000-000000000001';
DELETE FROM workouts WHERE id='59000000-0000-0000-0000-000000000001';
DROP INDEX workouts_one_default_warmup;
ALTER TABLE workouts DROP COLUMN is_default_warmup;
ALTER TABLE workouts DROP COLUMN warmup_enabled;
ALTER TABLE workouts DROP COLUMN difficulty;
ALTER TABLE workouts ALTER COLUMN day_number SET NOT NULL;
ALTER TABLE workouts ALTER COLUMN program_id SET NOT NULL;
