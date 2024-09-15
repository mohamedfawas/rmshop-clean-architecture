ALTER TABLE orders DROP COLUMN IF EXISTS has_return_request;
ALTER TABLE orders ADD COLUMN IF NOT EXISTS refund_status VARCHAR(20);
DROP INDEX IF EXISTS idx_orders_has_return_request;