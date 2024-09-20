ALTER TABLE checkout_sessions DROP COLUMN IF EXISTS status;
DROP INDEX IF EXISTS idx_checkout_sessions_status;