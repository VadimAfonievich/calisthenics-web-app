DROP INDEX IF EXISTS idx_exercises_tags_gin;
DROP INDEX IF EXISTS idx_exercises_equipment_gin;
DROP INDEX IF EXISTS idx_exercises_muscle_groups_gin;
DROP INDEX IF EXISTS idx_exercises_library_filters;
ALTER TABLE exercises DROP COLUMN tags;
