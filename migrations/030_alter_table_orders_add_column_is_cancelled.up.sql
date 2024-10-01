ALTER TABLE orders
ADD COLUMN is_cancelled BOOLEAN NOT NULL DEFAULT FALSE;

CREATE INDEX idx_orders_is_cancelled ON orders(is_cancelled);