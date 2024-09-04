-- Remove trigger
DROP TRIGGER IF EXISTS update_orders_updated_at ON orders;

-- Remove function
DROP FUNCTION IF EXISTS update_orders_updated_at();

-- Remove columns
ALTER TABLE orders
DROP COLUMN IF EXISTS order_status,
DROP COLUMN IF EXISTS refund_status,
DROP COLUMN IF EXISTS updated_at;