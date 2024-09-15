-- add has_return_request column
ALTER TABLE orders
ADD COLUMN IF NOT EXISTS has_return_request BOOLEAN NOT NULL DEFAULT FALSE;

-- remove refund_status column
ALTER TABLE orders 
DROP COLUMN IF EXISTS refund_status;

-- an index on the has_return_request column for faster queries
CREATE INDEX idx_orders_has_return_request ON orders(has_return_request);