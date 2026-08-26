ALTER TABLE exercises ADD COLUMN standard_key text;
ALTER TABLE exercises ADD CONSTRAINT exercises_standard_key_format CHECK (standard_key IS NULL OR standard_key ~ '^[a-z0-9]+(-[a-z0-9]+)*$');
ALTER TABLE exercises ADD CONSTRAINT exercises_standard_key_system_owned CHECK (standard_key IS NULL OR owner_user_id IS NULL);
CREATE UNIQUE INDEX exercises_standard_key_unique ON exercises(standard_key) WHERE standard_key IS NOT NULL;
