BEGIN;
INSERT INTO users(id,telegram_id,first_name) VALUES
 ('a0000000-0000-0000-0000-000000000001',190000001,'Coach A'),
 ('b0000000-0000-0000-0000-000000000001',190000002,'Coach B'),
 ('a0000000-0000-0000-0000-000000000002',190000003,'Student A'),
 ('b0000000-0000-0000-0000-000000000002',190000004,'Student B');
INSERT INTO profiles(user_id,display_name) SELECT id,first_name FROM users WHERE telegram_id BETWEEN 190000001 AND 190000004;
INSERT INTO user_progress(user_id) SELECT id FROM users WHERE telegram_id BETWEEN 190000001 AND 190000004;
INSERT INTO tenants(id,slug,name,owner_user_id) VALUES
 ('a1000000-0000-0000-0000-000000000001','tenant-a','Tenant A','a0000000-0000-0000-0000-000000000001'),
 ('b1000000-0000-0000-0000-000000000001','tenant-b','Tenant B','b0000000-0000-0000-0000-000000000001');
INSERT INTO tenant_memberships(tenant_id,user_id,role) VALUES
 ('a1000000-0000-0000-0000-000000000001','a0000000-0000-0000-0000-000000000001','coach'),
 ('b1000000-0000-0000-0000-000000000001','b0000000-0000-0000-0000-000000000001','coach'),
 ('a1000000-0000-0000-0000-000000000001','a0000000-0000-0000-0000-000000000002','student'),
 ('b1000000-0000-0000-0000-000000000001','b0000000-0000-0000-0000-000000000002','student');
INSERT INTO exercises(id,name,slug,description,instructions,common_mistakes,difficulty,muscle_groups,owner_user_id,tenant_id,standard_key) VALUES
 ('a2000000-0000-0000-0000-000000000001','Exercise A','phase19-exercise-a','a','a','a','beginner',ARRAY['core'],'a0000000-0000-0000-0000-000000000001','a1000000-0000-0000-0000-000000000001',NULL),
 ('b2000000-0000-0000-0000-000000000001','Exercise B','phase19-exercise-b','b','b','b','beginner',ARRAY['core'],'b0000000-0000-0000-0000-000000000001','b1000000-0000-0000-0000-000000000001',NULL);
INSERT INTO workouts(id,title,description,estimated_minutes,owner_user_id,tenant_id,status) VALUES
 ('a3000000-0000-0000-0000-000000000001','Workout A','a',10,'a0000000-0000-0000-0000-000000000001','a1000000-0000-0000-0000-000000000001','published'),
 ('b3000000-0000-0000-0000-000000000001','Workout B','b',10,'b0000000-0000-0000-0000-000000000001','b1000000-0000-0000-0000-000000000001','published');
INSERT INTO workout_exercises(workout_id,exercise_id,sets,target_reps,rest_seconds) VALUES ('a3000000-0000-0000-0000-000000000001','a2000000-0000-0000-0000-000000000001',1,1,0);
DO $$ BEGIN
 BEGIN
  INSERT INTO workout_exercises(workout_id,exercise_id,sets,target_reps,rest_seconds) VALUES ('a3000000-0000-0000-0000-000000000001','b2000000-0000-0000-0000-000000000001',1,1,0);
  RAISE EXCEPTION 'cross-tenant exercise reference was accepted';
 EXCEPTION WHEN raise_exception THEN
  IF SQLERRM='cross-tenant exercise reference was accepted' THEN RAISE; END IF;
 END;
END $$;
DO $$ DECLARE a_count int; b_count int; BEGIN
 SELECT count(*) INTO a_count FROM tenant_memberships WHERE tenant_id='a1000000-0000-0000-0000-000000000001' AND role='student';
 SELECT count(*) INTO b_count FROM tenant_memberships WHERE tenant_id='b1000000-0000-0000-0000-000000000001' AND role='student';
 IF a_count<>1 OR b_count<>1 THEN RAISE EXCEPTION 'analytics membership isolation failed'; END IF;
END $$;
ROLLBACK;
