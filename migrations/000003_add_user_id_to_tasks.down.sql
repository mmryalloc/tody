DROP INDEX IF EXISTS tasks_user_id_created_at_idx;
ALTER TABLE tasks DROP COLUMN IF EXISTS user_id;
