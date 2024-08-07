DROP INDEX IF EXISTS idx_products_deleted_at;
ALTER TABLE products DROP COLUMN IF EXISTS deleted_at;