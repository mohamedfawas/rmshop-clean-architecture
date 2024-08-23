ALTER TABLE products DROP CONSTRAINT IF EXISTS fk_sub_category;
DROP INDEX IF EXISTS idx_products_sub_category;
DROP INDEX IF EXISTS idx_products_slug;
DROP TABLE IF EXISTS products;
