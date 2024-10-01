DROP INDEX IF EXISTS idx_orders_is_cancelled;

ALTER TABLE orders DROP COLUMN IF EXISTS is_cancelled;

