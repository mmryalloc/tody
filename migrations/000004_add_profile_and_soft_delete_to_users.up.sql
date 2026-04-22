ALTER TABLE users
  ADD COLUMN name VARCHAR(255) NOT NULL DEFAULT '',
  ADD COLUMN deleted_at TIMESTAMP NULL;

CREATE INDEX IF NOT EXISTS users_active_id_idx
  ON users (id)
  WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS users_active_email_idx
  ON users (email)
  WHERE deleted_at IS NULL;
