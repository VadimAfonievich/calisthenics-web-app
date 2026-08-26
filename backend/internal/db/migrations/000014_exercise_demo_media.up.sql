ALTER TABLE exercises ADD COLUMN demo_media_id uuid REFERENCES media_assets(id) ON DELETE SET NULL;
CREATE INDEX idx_exercises_demo_media ON exercises(demo_media_id) WHERE demo_media_id IS NOT NULL;
