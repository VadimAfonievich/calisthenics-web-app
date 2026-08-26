ALTER TABLE exercises ADD COLUMN tags text[] NOT NULL DEFAULT '{}'::text[];
CREATE INDEX idx_exercises_library_filters ON exercises(status,difficulty,movement_type,owner_user_id);
CREATE INDEX idx_exercises_muscle_groups_gin ON exercises USING gin(muscle_groups);
CREATE INDEX idx_exercises_equipment_gin ON exercises USING gin(equipment);
CREATE INDEX idx_exercises_tags_gin ON exercises USING gin(tags);
