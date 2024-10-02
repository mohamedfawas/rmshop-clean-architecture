ALTER TABLE return_requests ADD COLUMN IF NOT EXISTS refund_completed BOOLEAN NOT NULL DEFAULT FALSE;

ALTER TABLE return_requests
DROP COLUMN IF EXISTS order_returned_to_seller_at,
DROP COLUMN IF EXISTS is_order_reached_the_seller,
DROP COLUMN IF EXISTS  is_stock_updated;

DROP INDEX IF EXISTS idx_return_requests_is_order_reached_the_seller;
DROP INDEX IF EXISTS idx_return_requests_is_stock_updated;