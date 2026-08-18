CREATE TABLE levels (
  level integer PRIMARY KEY CHECK (level > 0),
  name text NOT NULL UNIQUE,
  min_xp integer NOT NULL UNIQUE CHECK (min_xp >= 0)
);

INSERT INTO levels (level, name, min_xp) VALUES
  (1, 'Beginner', 0), (2, 'Novice', 500), (3, 'Intermediate', 1000),
  (4, 'Advanced', 1500), (5, 'Athlete', 2000);

ALTER TABLE profiles ADD COLUMN last_workout_date date;
