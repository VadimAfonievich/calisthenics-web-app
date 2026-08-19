ALTER TABLE programs
  ADD COLUMN category text NOT NULL DEFAULT 'OTHER'
  CHECK (category IN ('MORNING_ROUTINE','WARMUP','BASE_STRENGTH','SKILL','MOBILITY','OTHER'));

UPDATE programs SET category = 'BASE_STRENGTH' WHERE slug IN ('start-s-nulya','baza-kalisteniki');
UPDATE programs SET category = 'SKILL' WHERE slug = 'handstand-foundations';

CREATE TABLE program_levels (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  program_id uuid NOT NULL REFERENCES programs(id) ON DELETE CASCADE,
  level_number integer NOT NULL CHECK (level_number > 0),
  title text NOT NULL,
  description text NOT NULL,
  difficulty text NOT NULL CHECK (difficulty IN ('beginner','intermediate','advanced')),
  unlock_rule_type text NOT NULL CHECK (unlock_rule_type IN ('none','previous_level','workouts_completed','criterion')),
  unlock_rule_value integer NOT NULL DEFAULT 0 CHECK (unlock_rule_value >= 0),
  sort_order integer NOT NULL DEFAULT 0 CHECK (sort_order >= 0),
  created_at timestamptz NOT NULL DEFAULT NOW(),
  updated_at timestamptz NOT NULL DEFAULT NOW(),
  UNIQUE(program_id,level_number)
);

ALTER TABLE workouts ADD COLUMN program_level_id uuid REFERENCES program_levels(id) ON DELETE SET NULL;
CREATE INDEX idx_workouts_program_level ON workouts(program_level_id);

CREATE TABLE skills (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  code text NOT NULL UNIQUE,
  name text NOT NULL,
  description text NOT NULL,
  category text NOT NULL CHECK (category IN ('MORNING_ROUTINE','WARMUP','BASE_STRENGTH','SKILL','MOBILITY','OTHER')),
  difficulty text NOT NULL CHECK (difficulty IN ('beginner','intermediate','advanced')),
  icon text NOT NULL,
  xp_reward integer NOT NULL DEFAULT 0 CHECK (xp_reward >= 0),
  final_criterion_type text NOT NULL CHECK (final_criterion_type IN ('duration_seconds','repetitions','manual_confirmation')),
  final_criterion_value integer NOT NULL CHECK (final_criterion_value > 0),
  sort_order integer NOT NULL DEFAULT 0 CHECK (sort_order >= 0),
  created_at timestamptz NOT NULL DEFAULT NOW(),
  updated_at timestamptz NOT NULL DEFAULT NOW()
);

CREATE TABLE skill_levels (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  skill_id uuid NOT NULL REFERENCES skills(id) ON DELETE CASCADE,
  level_number integer NOT NULL CHECK (level_number > 0),
  name text NOT NULL,
  description text NOT NULL,
  program_level_id uuid REFERENCES program_levels(id) ON DELETE SET NULL,
  criterion_type text NOT NULL CHECK (criterion_type IN ('workout_completed','duration_seconds','repetitions','manual_confirmation')),
  criterion_value integer NOT NULL CHECK (criterion_value > 0),
  sort_order integer NOT NULL DEFAULT 0 CHECK (sort_order >= 0),
  created_at timestamptz NOT NULL DEFAULT NOW(),
  updated_at timestamptz NOT NULL DEFAULT NOW(),
  UNIQUE(skill_id,level_number)
);

CREATE TABLE skill_requirements (
  skill_id uuid NOT NULL REFERENCES skills(id) ON DELETE CASCADE,
  required_skill_id uuid NOT NULL REFERENCES skills(id) ON DELETE CASCADE,
  requirement_type text NOT NULL CHECK (requirement_type IN ('skill_mastered','skill_level')),
  requirement_value integer NOT NULL DEFAULT 0 CHECK (requirement_value >= 0),
  PRIMARY KEY(skill_id,required_skill_id),
  CHECK(skill_id <> required_skill_id)
);

CREATE TABLE user_skill_progress (
  user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  skill_id uuid NOT NULL REFERENCES skills(id) ON DELETE CASCADE,
  current_level integer NOT NULL DEFAULT 1 CHECK (current_level > 0),
  status text NOT NULL CHECK (status IN ('locked','available','in_progress','mastered')),
  started_at timestamptz,
  completed_at timestamptz,
  updated_at timestamptz NOT NULL DEFAULT NOW(),
  PRIMARY KEY(user_id,skill_id)
);

CREATE TABLE user_skill_level_progress (
  user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  skill_level_id uuid NOT NULL REFERENCES skill_levels(id) ON DELETE CASCADE,
  status text NOT NULL CHECK (status IN ('locked','available','in_progress','completed')),
  progress_value integer NOT NULL DEFAULT 0 CHECK (progress_value >= 0),
  completed_at timestamptz,
  updated_at timestamptz NOT NULL DEFAULT NOW(),
  PRIMARY KEY(user_id,skill_level_id)
);

INSERT INTO program_levels(id,program_id,level_number,title,description,difficulty,unlock_rule_type,unlock_rule_value,sort_order) VALUES
('45000000-0000-0000-0000-000000000001','44000000-0000-0000-0000-000000000001',1,'Подготовка тела','Сила корпуса, кистей и плеч для безопасной стойки.','beginner','none',0,1),
('45000000-0000-0000-0000-000000000002','44000000-0000-0000-0000-000000000001',2,'Стойка у стены','Активные плечи и ровная линия тела в стойке лицом к стене.','beginner','previous_level',1,2),
('45000000-0000-0000-0000-000000000003','44000000-0000-0000-0000-000000000001',3,'Контроль баланса','Отрывы от стены и управляемая точка равновесия.','intermediate','previous_level',1,3),
('45000000-0000-0000-0000-000000000004','44000000-0000-0000-0000-000000000001',4,'Свободная стойка','Контролируемые выходы и первые свободные удержания.','intermediate','previous_level',1,4),
('45000000-0000-0000-0000-000000000005','44000000-0000-0000-0000-000000000001',5,'Стойка освоена','Подтверждение устойчивой свободной стойки.','advanced','previous_level',1,5);

UPDATE workouts SET program_level_id='45000000-0000-0000-0000-000000000001' WHERE id='54000000-0000-0000-0000-000000000001';

INSERT INTO exercises(id,name,slug,description,instructions,common_mistakes,difficulty,muscle_groups,equipment) VALUES
('34000000-0000-0000-0000-000000000009','Отрыв носков от стены','handstand-toe-pull','Контроль баланса при стойке спиной к стене.','Мягко отведите носки от стены за счёт напряжения корпуса и удержите линию тела.','Отталкивание ногами, прогиб в пояснице и пассивные плечи.','intermediate',ARRAY['плечи','кор','кисти'],ARRAY['стена']),
('34000000-0000-0000-0000-000000000010','Свободное удержание стойки','free-handstand-hold','Практика самостоятельного баланса в стойке на руках.','Выйдите в стойку контролируемым махом и удерживайте ровную линию, используя пальцы для коррекции.','Сильный мах, задержка дыхания и отсутствие безопасного выхода.','intermediate',ARRAY['плечи','кор','кисти'],ARRAY[]::text[]);

INSERT INTO workouts(id,program_id,program_level_id,day_number,title,description,estimated_minutes,sort_order) VALUES
('54000000-0000-0000-0000-000000000002','44000000-0000-0000-0000-000000000001','45000000-0000-0000-0000-000000000002',2,'Стойка на руках — уровень 2','Линия тела, активные плечи и уверенное удержание у стены.',25,2),
('54000000-0000-0000-0000-000000000003','44000000-0000-0000-0000-000000000001','45000000-0000-0000-0000-000000000003',3,'Стойка на руках — уровень 3','Контролируемые отрывы от стены и первые секунды баланса.',30,3),
('54000000-0000-0000-0000-000000000004','44000000-0000-0000-0000-000000000001','45000000-0000-0000-0000-000000000004',4,'Стойка на руках — уровень 4','Свободные попытки, контроль выхода и удержания 3–5 секунд.',30,4);

INSERT INTO workout_exercises(workout_id,exercise_id,sets,target_reps,target_duration_seconds,rest_seconds,sort_order) VALUES
('54000000-0000-0000-0000-000000000002','34000000-0000-0000-0000-000000000005',3,3,NULL,75,1),
('54000000-0000-0000-0000-000000000002','34000000-0000-0000-0000-000000000004',3,10,NULL,60,2),
('54000000-0000-0000-0000-000000000002','34000000-0000-0000-0000-000000000006',4,NULL,40,75,3),
('54000000-0000-0000-0000-000000000003','34000000-0000-0000-0000-000000000007',5,3,NULL,60,1),
('54000000-0000-0000-0000-000000000003','34000000-0000-0000-0000-000000000009',5,3,NULL,60,2),
('54000000-0000-0000-0000-000000000003','34000000-0000-0000-0000-000000000006',3,NULL,30,75,3),
('54000000-0000-0000-0000-000000000004','34000000-0000-0000-0000-000000000008',6,3,NULL,60,1),
('54000000-0000-0000-0000-000000000004','34000000-0000-0000-0000-000000000010',6,NULL,5,75,2);

INSERT INTO skills(id,code,name,description,category,difficulty,icon,xp_reward,final_criterion_type,final_criterion_value,sort_order) VALUES
('65000000-0000-0000-0000-000000000001','HANDSTAND','Стойка на руках','Путь от подготовки тела до устойчивой свободной стойки.','SKILL','beginner','🤸',250,'duration_seconds',10,1),
('65000000-0000-0000-0000-000000000002','MUSCLE_UP','Выход силой','Сила тяги, переход над перекладиной и первый чистый выход.','SKILL','advanced','🚀',300,'repetitions',1,2),
('65000000-0000-0000-0000-000000000003','FRONT_LEVER','Передний вис','Прогрессия горизонтального удержания на перекладине.','SKILL','advanced','🏹',300,'duration_seconds',5,3),
('65000000-0000-0000-0000-000000000004','PLANCHE','Планш','Силовой баланс на прямых руках с горизонтальным корпусом.','SKILL','advanced','⚡',350,'duration_seconds',5,4),
('65000000-0000-0000-0000-000000000005','L_SIT','Уголок','Статическое удержание ног параллельно полу в упоре.','SKILL','intermediate','📐',200,'duration_seconds',10,5),
('65000000-0000-0000-0000-000000000006','PULL_UP_BASE','База подтягиваний','Базовая тяговая сила для сложных навыков.','BASE_STRENGTH','beginner','⬆️',100,'repetitions',10,6),
('65000000-0000-0000-0000-000000000007','DIP_BASE','База на брусьях','Базовая жимовая сила для выхода силой.','BASE_STRENGTH','beginner','💪',100,'repetitions',10,7),
('65000000-0000-0000-0000-000000000008','HANDSTAND_PUSHUP','Отжимание в стойке','Вертикальный жим в контролируемой стойке на руках.','SKILL','advanced','🔻',300,'repetitions',1,8);

INSERT INTO skill_levels(id,skill_id,level_number,name,description,program_level_id,criterion_type,criterion_value,sort_order) VALUES
('66000000-0000-0000-0000-000000000001','65000000-0000-0000-0000-000000000001',1,'Подготовка тела','Подготовьте кисти, плечи и корпус.','45000000-0000-0000-0000-000000000001','workout_completed',1,1),
('66000000-0000-0000-0000-000000000002','65000000-0000-0000-0000-000000000001',2,'Стойка лицом к стене','Удерживайте активную стойку у стены 40 секунд.','45000000-0000-0000-0000-000000000002','duration_seconds',40,2),
('66000000-0000-0000-0000-000000000003','65000000-0000-0000-0000-000000000001',3,'Баланс у стены','Освойте heel pulls, toe pulls и контролируемый баланс.','45000000-0000-0000-0000-000000000003','repetitions',5,3),
('66000000-0000-0000-0000-000000000004','65000000-0000-0000-0000-000000000001',4,'Свободные попытки','Удерживайте свободную стойку 3–5 секунд.','45000000-0000-0000-0000-000000000004','duration_seconds',5,4),
('66000000-0000-0000-0000-000000000005','65000000-0000-0000-0000-000000000001',5,'Стойка освоена','Подтвердите свободную стойку не менее 10 секунд.','45000000-0000-0000-0000-000000000005','duration_seconds',10,5);

INSERT INTO skill_levels(id,skill_id,level_number,name,description,criterion_type,criterion_value,sort_order) VALUES
('66000000-0000-0000-0000-000000000011','65000000-0000-0000-0000-000000000002',1,'Тяговая база','Уверенные строгие подтягивания.','repetitions',10,1),
('66000000-0000-0000-0000-000000000012','65000000-0000-0000-0000-000000000002',2,'Взрывная тяга','Подтягивание к груди с высокой траекторией.','repetitions',5,2),
('66000000-0000-0000-0000-000000000013','65000000-0000-0000-0000-000000000002',3,'Жим над перекладиной','Уверенный straight bar dip.','repetitions',5,3),
('66000000-0000-0000-0000-000000000014','65000000-0000-0000-0000-000000000002',4,'Переход','Контролируемый переход над перекладиной.','repetitions',3,4),
('66000000-0000-0000-0000-000000000015','65000000-0000-0000-0000-000000000002',5,'Выход с поддержкой','Выполните assisted muscle-up.','repetitions',3,5),
('66000000-0000-0000-0000-000000000016','65000000-0000-0000-0000-000000000002',6,'Первый выход силой','Выполните первый чистый muscle-up.','repetitions',1,6),
('66000000-0000-0000-0000-000000000021','65000000-0000-0000-0000-000000000003',1,'Активный вис','Удерживайте активные плечи в висе.','duration_seconds',20,1),
('66000000-0000-0000-0000-000000000022','65000000-0000-0000-0000-000000000003',2,'Tuck','Удерживайте tuck front lever.','duration_seconds',10,2),
('66000000-0000-0000-0000-000000000023','65000000-0000-0000-0000-000000000003',3,'Advanced Tuck','Удерживайте advanced tuck.','duration_seconds',8,3),
('66000000-0000-0000-0000-000000000024','65000000-0000-0000-0000-000000000003',4,'One Leg','Удерживайте вариант с одной ногой.','duration_seconds',5,4),
('66000000-0000-0000-0000-000000000025','65000000-0000-0000-0000-000000000003',5,'Straddle','Удерживайте straddle front lever.','duration_seconds',5,5),
('66000000-0000-0000-0000-000000000026','65000000-0000-0000-0000-000000000003',6,'Full Front Lever','Удерживайте полный передний вис.','duration_seconds',5,6),
('66000000-0000-0000-0000-000000000031','65000000-0000-0000-0000-000000000004',1,'Planche Lean','Контролируйте перенос плеч вперёд.','duration_seconds',20,1),
('66000000-0000-0000-0000-000000000032','65000000-0000-0000-0000-000000000004',2,'Frog Stand','Освойте базовый баланс на руках.','duration_seconds',15,2),
('66000000-0000-0000-0000-000000000033','65000000-0000-0000-0000-000000000004',3,'Tuck Planche','Удерживайте tuck planche.','duration_seconds',8,3),
('66000000-0000-0000-0000-000000000034','65000000-0000-0000-0000-000000000004',4,'Advanced Tuck','Удерживайте advanced tuck planche.','duration_seconds',5,4),
('66000000-0000-0000-0000-000000000035','65000000-0000-0000-0000-000000000004',5,'Straddle','Удерживайте straddle planche.','duration_seconds',5,5),
('66000000-0000-0000-0000-000000000036','65000000-0000-0000-0000-000000000004',6,'Full Planche','Удерживайте полный планш.','duration_seconds',5,6),
('66000000-0000-0000-0000-000000000041','65000000-0000-0000-0000-000000000005',1,'Support Hold','Уверенно удерживайте упор.','duration_seconds',20,1),
('66000000-0000-0000-0000-000000000042','65000000-0000-0000-0000-000000000005',2,'Knee Raises','Поднимайте колени в упоре.','repetitions',10,2),
('66000000-0000-0000-0000-000000000043','65000000-0000-0000-0000-000000000005',3,'Tuck Sit','Удерживайте tuck sit.','duration_seconds',15,3),
('66000000-0000-0000-0000-000000000044','65000000-0000-0000-0000-000000000005',4,'One Leg','Выпрямите одну ногу в упоре.','duration_seconds',10,4),
('66000000-0000-0000-0000-000000000045','65000000-0000-0000-0000-000000000005',5,'Full L-Sit','Удерживайте полный уголок.','duration_seconds',10,5),
('66000000-0000-0000-0000-000000000061','65000000-0000-0000-0000-000000000006',1,'10 подтягиваний','Выполните десять строгих подтягиваний.','repetitions',10,1),
('66000000-0000-0000-0000-000000000071','65000000-0000-0000-0000-000000000007',1,'10 отжиманий на брусьях','Выполните десять строгих повторений.','repetitions',10,1),
('66000000-0000-0000-0000-000000000081','65000000-0000-0000-0000-000000000008',1,'Силовая база в стойке','Освойте стойку и вертикальный жим.','manual_confirmation',1,1);

INSERT INTO skill_requirements(skill_id,required_skill_id,requirement_type,requirement_value) VALUES
('65000000-0000-0000-0000-000000000002','65000000-0000-0000-0000-000000000006','skill_mastered',0),
('65000000-0000-0000-0000-000000000002','65000000-0000-0000-0000-000000000007','skill_mastered',0),
('65000000-0000-0000-0000-000000000008','65000000-0000-0000-0000-000000000001','skill_mastered',0);

INSERT INTO achievements(id,code,title,description,icon,xp_reward,condition_type,condition_value) VALUES
('67000000-0000-0000-0000-000000000001','HANDSTAND_MASTERED','Стойка освоена','Освойте свободную стойку на руках.','🤸',0,'exercise_completed',1),
('67000000-0000-0000-0000-000000000002','MUSCLE_UP_MASTERED','Выход силой освоен','Выполните первый чистый выход силой.','🚀',0,'exercise_completed',1),
('67000000-0000-0000-0000-000000000003','FRONT_LEVER_MASTERED','Передний вис освоен','Освойте полный передний вис.','🏹',0,'exercise_completed',1),
('67000000-0000-0000-0000-000000000004','PLANCHE_MASTERED','Планш освоен','Освойте полный планш.','⚡',0,'exercise_completed',1),
('67000000-0000-0000-0000-000000000005','L_SIT_MASTERED','Уголок освоен','Освойте полный L-Sit.','📐',0,'exercise_completed',1);
