ALTER TABLE payments DROP COLUMN IF EXISTS payment_status;
DROP INDEX IF EXISTS idx_payments_payment_status;