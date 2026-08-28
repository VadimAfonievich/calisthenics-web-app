CREATE TABLE user_program_progress (
  user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  program_id uuid NOT NULL REFERENCES programs(id) ON DELETE CASCADE,
  status text NOT NULL DEFAULT 'active' CHECK (status IN ('active','completed')),
  current_level integer NOT NULL DEFAULT 1 CHECK (current_level > 0),
  started_at timestamptz NOT NULL DEFAULT NOW(),
  completed_at timestamptz,
  updated_at timestamptz NOT NULL DEFAULT NOW(),
  PRIMARY KEY(user_id,program_id),
  CHECK ((status='completed' AND completed_at IS NOT NULL) OR (status='active' AND completed_at IS NULL))
);

CREATE INDEX idx_user_program_progress_user_status ON user_program_progress(user_id,status,updated_at DESC);
CREATE TRIGGER user_program_progress_set_updated_at BEFORE UPDATE ON user_program_progress FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TABLE role_change_audit (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  actor_user_id uuid NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
  target_user_id uuid NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
  old_role text NOT NULL,
  new_role text NOT NULL CHECK (new_role IN ('user','coach')),
  created_at timestamptz NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_role_change_audit_target ON role_change_audit(target_user_id,created_at DESC);
