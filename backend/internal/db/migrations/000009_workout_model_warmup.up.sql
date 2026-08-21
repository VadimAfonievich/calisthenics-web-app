ALTER TABLE workouts ALTER COLUMN program_id DROP NOT NULL;
ALTER TABLE workouts ALTER COLUMN day_number DROP NOT NULL;
ALTER TABLE workouts ADD COLUMN difficulty text NOT NULL DEFAULT 'beginner'
  CHECK (difficulty IN ('beginner','intermediate','advanced'));
UPDATE workouts w SET difficulty=p.difficulty FROM programs p WHERE p.id=w.program_id;
ALTER TABLE workouts ADD COLUMN warmup_enabled boolean NOT NULL DEFAULT true;
ALTER TABLE workouts ADD COLUMN is_default_warmup boolean NOT NULL DEFAULT false;
UPDATE workouts SET warmup_enabled=false WHERE category='warmup';
CREATE UNIQUE INDEX workouts_one_default_warmup
  ON workouts (is_default_warmup) WHERE is_default_warmup;

-- Deterministic initial default without relying on a localized title.
UPDATE workouts SET is_default_warmup=true
WHERE id=(SELECT id FROM workouts WHERE category='warmup' AND status='published' ORDER BY sort_order,id LIMIT 1);

INSERT INTO workouts(id,title,description,difficulty,estimated_minutes,sort_order,category,warmup_enabled,is_default_warmup,status,published_at)
VALUES(
  '59000000-0000-0000-0000-000000000001',
  'Полная разминка',
  'Подготовьте суставы, корпус и основные мышечные группы к тренировке.',
  'beginner',12,0,'warmup',false,
  NOT EXISTS(SELECT 1 FROM workouts WHERE is_default_warmup),
  'published',NOW()
)
ON CONFLICT(id) DO NOTHING;

INSERT INTO workout_exercises(workout_id,exercise_id,sets,target_reps,rest_seconds,sort_order,notes) VALUES
('59000000-0000-0000-0000-000000000001','34000000-0000-0000-0000-000000000001',1,12,15,0,'Двигайтесь плавно, без боли.'),
('59000000-0000-0000-0000-000000000001','30000000-0000-0000-0000-000000000010',2,10,20,1,'Сохраняйте контроль лопаток.'),
('59000000-0000-0000-0000-000000000001','30000000-0000-0000-0000-000000000002',2,12,20,2,'Разогрейте колени и тазобедренные суставы.'),
('59000000-0000-0000-0000-000000000001','30000000-0000-0000-0000-000000000003',1,10,15,3,'Активируйте корпус перед основной нагрузкой.')
ON CONFLICT(workout_id,exercise_id) DO NOTHING;
