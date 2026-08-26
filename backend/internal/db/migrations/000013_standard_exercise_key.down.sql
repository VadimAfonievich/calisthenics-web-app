DROP INDEX IF EXISTS exercises_standard_key_unique;
ALTER TABLE exercises DROP CONSTRAINT IF EXISTS exercises_standard_key_system_owned;
ALTER TABLE exercises DROP CONSTRAINT IF EXISTS exercises_standard_key_format;
ALTER TABLE exercises DROP COLUMN IF EXISTS standard_key;
