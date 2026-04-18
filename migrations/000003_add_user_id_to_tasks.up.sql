ALTER TABLE tasks
  ADD COLUMN user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE;

CREATE INDEX IF NOT EXISTS tasks_user_id_created_at_idx
  ON tasks (user_id, created_at DESC);
