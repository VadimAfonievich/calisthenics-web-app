ALTER TABLE program_levels
  ADD COLUMN mastery_type text NOT NULL DEFAULT 'manual' CHECK (mastery_type IN ('duration','reps','manual')),
  ADD COLUMN mastery_value integer,
  ADD COLUMN mastery_name text NOT NULL DEFAULT '',
  ADD COLUMN mastery_description text NOT NULL DEFAULT 'Подтвердите, что вы освоили навык этого этапа.',
  ADD CONSTRAINT program_levels_mastery_value CHECK (
    (mastery_type='manual' AND mastery_value IS NULL) OR
    (mastery_type IN ('duration','reps') AND mastery_value IS NOT NULL AND mastery_value>0)
  );

UPDATE program_levels SET mastery_name=title WHERE mastery_name='';
UPDATE program_levels
SET mastery_type='duration',mastery_value=15,mastery_name='Накат',
    mastery_description='Удержи Накат не менее 15 секунд с правильной техникой.'
WHERE lower(trim(title))='накат';

CREATE TABLE user_program_level_mastery (
  user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  tenant_id uuid NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  program_id uuid NOT NULL REFERENCES programs(id) ON DELETE CASCADE,
  program_level_id uuid NOT NULL REFERENCES program_levels(id) ON DELETE CASCADE,
  mastered_at timestamptz NOT NULL DEFAULT NOW(),
  PRIMARY KEY(user_id,tenant_id,program_id,program_level_id)
);
CREATE INDEX idx_program_level_mastery_progress ON user_program_level_mastery(user_id,tenant_id,program_id,mastered_at);

-- Legacy current_level/status was inferred only from completed workouts. Under
-- mastery semantics no stage is confirmed yet, so safely return each enrolled
-- student to the first stage without touching workout history or statistics.
UPDATE user_program_progress upp
SET current_level=first_level.level_number,status='active',completed_at=NULL
FROM (SELECT program_id,min(level_number) level_number FROM program_levels GROUP BY program_id) first_level
WHERE upp.program_id=first_level.program_id;

CREATE FUNCTION enforce_program_level_mastery_tenant() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
  IF NOT EXISTS(
    SELECT 1 FROM program_levels pl JOIN programs p ON p.id=pl.program_id
    WHERE pl.id=NEW.program_level_id AND p.id=NEW.program_id AND p.tenant_id=NEW.tenant_id
  ) THEN RAISE EXCEPTION 'cross-tenant program stage mastery'; END IF;
  RETURN NEW;
END $$;
CREATE TRIGGER program_level_mastery_tenant_guard BEFORE INSERT OR UPDATE ON user_program_level_mastery FOR EACH ROW EXECUTE FUNCTION enforce_program_level_mastery_tenant();
