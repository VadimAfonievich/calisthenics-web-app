ALTER TABLE workouts ADD COLUMN category text NOT NULL DEFAULT 'strength'
  CHECK (category IN ('warmup','morning','strength','skill'));

UPDATE workouts w SET category=CASE p.category
  WHEN 'WARMUP' THEN 'warmup'
  WHEN 'MORNING_ROUTINE' THEN 'morning'
  WHEN 'SKILL' THEN 'skill'
  ELSE 'strength'
END FROM programs p WHERE p.id=w.program_id;

ALTER TABLE skills ADD COLUMN hidden boolean NOT NULL DEFAULT false;
UPDATE skills SET sort_order=sort_order+1 WHERE sort_order>0;
UPDATE skills SET hidden=true WHERE code IN ('PULL_UP_BASE','DIP_BASE');

-- The two technical base nodes used to gate MUSCLE_UP. Their role is now
-- represented by the single, user-facing CALISTHENICS_BASE progression.
DELETE FROM skill_requirements
WHERE required_skill_id IN (
  SELECT id FROM skills WHERE code IN ('PULL_UP_BASE','DIP_BASE')
);

INSERT INTO skills(id,code,name,description,category,difficulty,icon,xp_reward,final_criterion_type,final_criterion_value,sort_order,owner_user_id,status,published_at)
VALUES('65000000-0000-0000-0000-000000000009','CALISTHENICS_BASE','Базовая физическая подготовка','Фундаментальная сила и контроль тела перед освоением элементов.','BASE_STRENGTH','beginner','🛡️',200,'manual_confirmation',1,1,NULL,'published',NOW())
ON CONFLICT(code) DO UPDATE SET hidden=false,sort_order=1,status='published';

CREATE TABLE skill_criteria (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  skill_id uuid NOT NULL REFERENCES skills(id) ON DELETE CASCADE,
  code text NOT NULL,
  title text NOT NULL,
  criterion_type text NOT NULL CHECK (criterion_type IN ('repetitions','duration_seconds','manual_confirmation')),
  target_value integer NOT NULL CHECK (target_value>0),
  sort_order integer NOT NULL CHECK (sort_order>=0),
  UNIQUE(skill_id,code), UNIQUE(skill_id,sort_order)
);

CREATE TABLE user_skill_criteria (
  user_id uuid NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
  criterion_id uuid NOT NULL REFERENCES skill_criteria(id) ON DELETE RESTRICT,
  confirmed_at timestamptz NOT NULL DEFAULT NOW(),
  PRIMARY KEY(user_id,criterion_id)
);

INSERT INTO skill_criteria(id,skill_id,code,title,criterion_type,target_value,sort_order) VALUES
('68000000-0000-0000-0000-000000000001','65000000-0000-0000-0000-000000000009','PUSH_UP_20','20 отжиманий от пола','repetitions',20,1),
('68000000-0000-0000-0000-000000000002','65000000-0000-0000-0000-000000000009','DIP_20','20 отжиманий на брусьях','repetitions',20,2),
('68000000-0000-0000-0000-000000000003','65000000-0000-0000-0000-000000000009','PULL_UP_15','15 подтягиваний','repetitions',15,3),
('68000000-0000-0000-0000-000000000004','65000000-0000-0000-0000-000000000009','HANGING_LEG_RAISE_15','15 подъёмов ног в висе','repetitions',15,4),
('68000000-0000-0000-0000-000000000005','65000000-0000-0000-0000-000000000009','L_SIT_30','Уголок — 30 секунд','duration_seconds',30,5),
('68000000-0000-0000-0000-000000000006','65000000-0000-0000-0000-000000000009','SQUAT_50','50 приседаний','repetitions',50,6),
('68000000-0000-0000-0000-000000000007','65000000-0000-0000-0000-000000000009','PLANK_60','Планка — 60 секунд','duration_seconds',60,7),
('68000000-0000-0000-0000-000000000008','65000000-0000-0000-0000-000000000009','DEAD_HANG_60','Вис на турнике — 60 секунд','duration_seconds',60,8);

INSERT INTO skill_requirements(skill_id,required_skill_id,requirement_type,requirement_value)
SELECT id,'65000000-0000-0000-0000-000000000009','skill_mastered',0 FROM skills WHERE code IN ('HANDSTAND','MUSCLE_UP','PLANCHE')
ON CONFLICT DO NOTHING;

ALTER TABLE achievements DROP CONSTRAINT achievements_condition_type_check;
ALTER TABLE achievements ADD CONSTRAINT achievements_condition_type_check
  CHECK (condition_type IN ('workouts_completed','streak_days','exercise_completed','exercise_max_reps','skill_mastered'));

INSERT INTO achievements(id,code,title,description,icon,xp_reward,condition_type,condition_value)
VALUES('67000000-0000-0000-0000-000000000009','CALISTHENICS_BASE_READY','База готова','Выполнены восемь критериев базовой физической подготовки.','🛡️',0,'skill_mastered',1)
ON CONFLICT(code) DO NOTHING;
