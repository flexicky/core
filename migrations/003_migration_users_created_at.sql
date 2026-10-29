-- created_at is read into a non-nullable time.Time, so make the column enforce
-- it. Existing rows without a value get the current timestamp.
UPDATE users SET created_at = CURRENT_TIMESTAMP WHERE created_at IS NULL;

ALTER TABLE users ALTER COLUMN created_at SET NOT NULL;

ALTER TABLE users ALTER COLUMN created_at SET DEFAULT CURRENT_TIMESTAMP;

-- idx_users_name_email duplicates the index that the UNIQUE constraint on
-- users.email already provides.
DROP INDEX IF EXISTS idx_users_name_email;
