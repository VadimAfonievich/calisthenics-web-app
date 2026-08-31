DROP TRIGGER IF EXISTS program_level_mastery_tenant_guard ON user_program_level_mastery;
DROP FUNCTION IF EXISTS enforce_program_level_mastery_tenant();
DROP INDEX IF EXISTS idx_program_level_mastery_progress;
DROP TABLE IF EXISTS user_program_level_mastery;
ALTER TABLE program_levels DROP CONSTRAINT IF EXISTS program_levels_mastery_value;
ALTER TABLE program_levels
  DROP COLUMN mastery_description,
  DROP COLUMN mastery_name,
  DROP COLUMN mastery_value,
  DROP COLUMN mastery_type;
