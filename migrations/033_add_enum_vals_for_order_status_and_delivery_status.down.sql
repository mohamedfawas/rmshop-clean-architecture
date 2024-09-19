-- Remove indexes
DROP INDEX IF EXISTS idx_orders_order_status;
DROP INDEX IF EXISTS idx_orders_delivery_status;

-- Remove the columns
ALTER TABLE orders
DROP COLUMN IF EXISTS order_status;

ALTER TABLE orders
DROP COLUMN IF EXISTS delivery_status;
