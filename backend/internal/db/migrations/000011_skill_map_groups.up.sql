ALTER TABLE skills
ADD COLUMN map_group text NOT NULL DEFAULT 'basic'
CHECK (map_group IN ('basic','floor','bar','parallel_bars'));

UPDATE skills SET map_group=CASE
  WHEN code IN ('HANDSTAND','PLANCHE','L_SIT','HANDSTAND_PUSHUP') THEN 'floor'
  WHEN code IN ('PULL_UP_BASE','MUSCLE_UP','FRONT_LEVER') THEN 'bar'
  WHEN code IN ('DIP_BASE') THEN 'parallel_bars'
  ELSE 'basic'
END;

CREATE INDEX idx_skills_map_group_order ON skills(map_group,sort_order);
