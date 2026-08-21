DELETE FROM user_achievements WHERE achievement_id='67000000-0000-0000-0000-000000000009';
DELETE FROM achievements WHERE id='67000000-0000-0000-0000-000000000009';
ALTER TABLE achievements DROP CONSTRAINT achievements_condition_type_check;
ALTER TABLE achievements ADD CONSTRAINT achievements_condition_type_check
  CHECK (condition_type IN ('workouts_completed','streak_days','exercise_completed','exercise_max_reps'));
DELETE FROM skill_requirements WHERE required_skill_id='65000000-0000-0000-0000-000000000009';
DROP TABLE user_skill_criteria;
DROP TABLE skill_criteria;
DELETE FROM user_skill_progress WHERE skill_id='65000000-0000-0000-0000-000000000009';
DELETE FROM skills WHERE id='65000000-0000-0000-0000-000000000009';
UPDATE skills SET hidden=false WHERE code IN ('PULL_UP_BASE','DIP_BASE');
INSERT INTO skill_requirements(skill_id,required_skill_id,requirement_type,requirement_value) VALUES
('65000000-0000-0000-0000-000000000002','65000000-0000-0000-0000-000000000006','skill_mastered',0),
('65000000-0000-0000-0000-000000000002','65000000-0000-0000-0000-000000000007','skill_mastered',0)
ON CONFLICT DO NOTHING;
UPDATE skills SET sort_order=sort_order-1 WHERE sort_order>1;
ALTER TABLE skills DROP COLUMN hidden;
ALTER TABLE workouts DROP COLUMN category;
