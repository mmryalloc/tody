DROP INDEX IF EXISTS users_active_email_idx;
DROP INDEX IF EXISTS users_active_id_idx;

ALTER TABLE users
  DROP COLUMN IF EXISTS deleted_at,
  DROP COLUMN IF EXISTS name;
