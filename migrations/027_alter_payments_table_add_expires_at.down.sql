ALTER TABLE payments
DROP COLUMN IF EXISTS expires_at;

DROP INDEX IF EXISTS idx_payments_expires_at;