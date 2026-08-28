DROP INDEX IF EXISTS idx_role_change_audit_target;
DROP TABLE IF EXISTS role_change_audit;
DROP TRIGGER IF EXISTS user_program_progress_set_updated_at ON user_program_progress;
DROP INDEX IF EXISTS idx_user_program_progress_user_status;
DROP TABLE IF EXISTS user_program_progress;
