ALTER TABLE return_requests DROP COLUMN IF EXISTS refund_completed;


ALTER TABLE return_requests
ADD COLUMN order_returned_to_seller_at TIMESTAMP WITH TIME ZONE,
ADD COLUMN is_order_reached_the_seller BOOLEAN DEFAULT FALSE,
ADD COLUMN is_stock_updated BOOLEAN NOT NULL DEFAULT FALSE;


CREATE INDEX IF NOT EXISTS idx_return_requests_is_order_reached_the_seller ON return_requests(is_order_reached_the_seller);
CREATE INDEX IF NOT EXISTS idx_return_requests_is_stock_updated ON return_requests(is_stock_updated);