INSERT INTO users(id,telegram_id,first_name) VALUES
('97000000-0000-0000-0000-000000000001',970000001,'Super Admin'),
('97000000-0000-0000-0000-000000000002',970000002,'Coach A'),
('97000000-0000-0000-0000-000000000003',970000003,'Coach B'),
('97000000-0000-0000-0000-000000000004',970000004,'Student 1'),
('97000000-0000-0000-0000-000000000005',970000005,'Student 2'),
('97000000-0000-0000-0000-000000000006',970000006,'Student 3');
INSERT INTO profiles(user_id,display_name) SELECT id,first_name FROM users WHERE telegram_id BETWEEN 970000001 AND 970000006;
INSERT INTO user_progress(user_id) SELECT id FROM users WHERE telegram_id BETWEEN 970000001 AND 970000006;
INSERT INTO admin_users(user_id,role) VALUES
('97000000-0000-0000-0000-000000000001','super_admin'),
('97000000-0000-0000-0000-000000000002','coach'),
('97000000-0000-0000-0000-000000000003','coach');

INSERT INTO media_assets(id,owner_user_id,type,status,storage_provider,storage_key,url,original_filename,mime_type,size_bytes) VALUES
('97000000-0000-0000-0000-000000000011','97000000-0000-0000-0000-000000000002','image','ready','fixture','coach-a.jpg','https://fixture/a.jpg','a.jpg','image/jpeg',100),
('97000000-0000-0000-0000-000000000012','97000000-0000-0000-0000-000000000003','image','ready','fixture','coach-b.jpg','https://fixture/b.jpg','b.jpg','image/jpeg',100);
INSERT INTO exercises(id,name,slug,description,instructions,common_mistakes,difficulty,muscle_groups,owner_user_id,status,standard_key) VALUES
('97000000-0000-0000-0000-000000000021','Coach A Exercise','fixture-coach-a-exercise','A','A','A','beginner',ARRAY['core'],'97000000-0000-0000-0000-000000000002','published',NULL),
('97000000-0000-0000-0000-000000000022','Coach B Exercise','fixture-coach-b-exercise','B','B','B','beginner',ARRAY['core'],'97000000-0000-0000-0000-000000000003','published',NULL);
INSERT INTO lessons(id,category_id,title,slug,short_description,content,difficulty,duration_minutes,published,owner_user_id,status) SELECT
'97000000-0000-0000-0000-000000000031',id,'Coach A Lesson','fixture-coach-a-lesson','A','A','beginner',5,true,'97000000-0000-0000-0000-000000000002','published' FROM lesson_categories ORDER BY sort_order LIMIT 1;
INSERT INTO lessons(id,category_id,title,slug,short_description,content,difficulty,duration_minutes,published,owner_user_id,status) SELECT
'97000000-0000-0000-0000-000000000032',id,'Coach B Lesson','fixture-coach-b-lesson','B','B','beginner',5,true,'97000000-0000-0000-0000-000000000003','published' FROM lesson_categories ORDER BY sort_order LIMIT 1;
INSERT INTO programs(id,name,slug,description,difficulty,duration_weeks,published,category,owner_user_id,status) VALUES
('97000000-0000-0000-0000-000000000041','Coach A Program','fixture-coach-a-program','A','beginner',2,true,'SKILL','97000000-0000-0000-0000-000000000002','published'),
('97000000-0000-0000-0000-000000000042','Coach B Program','fixture-coach-b-program','B','beginner',2,true,'SKILL','97000000-0000-0000-0000-000000000003','published');
INSERT INTO program_levels(id,program_id,level_number,title,description,difficulty,unlock_rule_type,sort_order) VALUES
('97000000-0000-0000-0000-000000000051','97000000-0000-0000-0000-000000000041',1,'A level','A','beginner','none',1),
('97000000-0000-0000-0000-000000000052','97000000-0000-0000-0000-000000000042',1,'B level','B','beginner','none',1);
INSERT INTO workouts(id,program_id,program_level_id,day_number,title,description,estimated_minutes,owner_user_id,status,category) VALUES
('97000000-0000-0000-0000-000000000061','97000000-0000-0000-0000-000000000041','97000000-0000-0000-0000-000000000051',1,'Coach A Workout','A',10,'97000000-0000-0000-0000-000000000002','published','strength'),
('97000000-0000-0000-0000-000000000062','97000000-0000-0000-0000-000000000042','97000000-0000-0000-0000-000000000052',1,'Coach B Workout','B',10,'97000000-0000-0000-0000-000000000003','published','strength');
INSERT INTO workout_exercises(workout_id,exercise_id,sets,target_reps,rest_seconds) VALUES
('97000000-0000-0000-0000-000000000061','97000000-0000-0000-0000-000000000021',1,5,30),
('97000000-0000-0000-0000-000000000062','97000000-0000-0000-0000-000000000022',1,5,30);
INSERT INTO skills(id,code,name,description,category,difficulty,icon,xp_reward,final_criterion_type,final_criterion_value,owner_user_id,status) VALUES
('97000000-0000-0000-0000-000000000071','FIXTURE_SKILL_A','Coach A Skill','A','SKILL','beginner','A',10,'repetitions',1,'97000000-0000-0000-0000-000000000002','published'),
('97000000-0000-0000-0000-000000000072','FIXTURE_SKILL_B','Coach B Skill','B','SKILL','beginner','B',10,'repetitions',1,'97000000-0000-0000-0000-000000000003','published');
INSERT INTO skill_levels(id,skill_id,level_number,name,description,criterion_type,criterion_value,sort_order) VALUES
('97000000-0000-0000-0000-000000000081','97000000-0000-0000-0000-000000000071',1,'A','A','repetitions',1,1),
('97000000-0000-0000-0000-000000000082','97000000-0000-0000-0000-000000000072',1,'B','B','repetitions',1,1);

INSERT INTO workout_sessions(id,user_id,workout_id,status,completed_at,duration_seconds) VALUES
('97000000-0000-0000-0000-000000000091','97000000-0000-0000-0000-000000000004','97000000-0000-0000-0000-000000000061','completed',NOW(),600),
('97000000-0000-0000-0000-000000000092','97000000-0000-0000-0000-000000000004','97000000-0000-0000-0000-000000000062','completed',NOW(),700),
('97000000-0000-0000-0000-000000000093','97000000-0000-0000-0000-000000000005','97000000-0000-0000-0000-000000000061','completed',NOW(),500);
INSERT INTO exercise_sets(session_id,exercise_id,set_number,reps,completed) VALUES
('97000000-0000-0000-0000-000000000091','97000000-0000-0000-0000-000000000021',1,5,true),
('97000000-0000-0000-0000-000000000092','97000000-0000-0000-0000-000000000022',1,5,true);
INSERT INTO user_program_progress(user_id,program_id) VALUES
('97000000-0000-0000-0000-000000000004','97000000-0000-0000-0000-000000000041'),
('97000000-0000-0000-0000-000000000004','97000000-0000-0000-0000-000000000042');
INSERT INTO user_skill_progress(user_id,skill_id,current_level,status,started_at) VALUES
('97000000-0000-0000-0000-000000000004','97000000-0000-0000-0000-000000000071',1,'in_progress',NOW()),
('97000000-0000-0000-0000-000000000004','97000000-0000-0000-0000-000000000072',1,'in_progress',NOW());
INSERT INTO user_skill_level_progress(user_id,skill_level_id,status,progress_value) VALUES
('97000000-0000-0000-0000-000000000004','97000000-0000-0000-0000-000000000081','in_progress',1);
INSERT INTO user_lesson_progress(user_id,lesson_id,completed,progress_percent,completed_at) VALUES
('97000000-0000-0000-0000-000000000004','97000000-0000-0000-0000-000000000031',true,100,NOW()),
('97000000-0000-0000-0000-000000000006','97000000-0000-0000-0000-000000000032',true,100,NOW());
INSERT INTO user_training_schedules(id,user_id,workout_id,timezone,start_date) VALUES
('97000000-0000-0000-0000-0000000000a1','97000000-0000-0000-0000-000000000004','97000000-0000-0000-0000-000000000061','UTC',CURRENT_DATE),
('97000000-0000-0000-0000-0000000000a2','97000000-0000-0000-0000-000000000006','97000000-0000-0000-0000-000000000062','UTC',CURRENT_DATE);
INSERT INTO user_planned_workouts(user_id,workout_id,scheduled_date,timezone) VALUES
('97000000-0000-0000-0000-000000000004','97000000-0000-0000-0000-000000000061',CURRENT_DATE,'UTC'),
('97000000-0000-0000-0000-000000000006','97000000-0000-0000-0000-000000000062',CURRENT_DATE,'UTC');
INSERT INTO user_exercise_stats(user_id,exercise_id,total_sets,total_reps,max_reps,last_performed_at) VALUES
('97000000-0000-0000-0000-000000000004','97000000-0000-0000-0000-000000000021',1,5,5,NOW()),
('97000000-0000-0000-0000-000000000004','97000000-0000-0000-0000-000000000022',1,5,5,NOW());
