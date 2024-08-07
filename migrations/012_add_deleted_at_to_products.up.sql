ALTER TABLE products
ADD COLUMN deleted_at TIMESTAMP WITHOUT TIME ZONE;

-- Add an index to improve query performance on the deleted_at column
CREATE INDEX idx_products_deleted_at ON products(deleted_at);