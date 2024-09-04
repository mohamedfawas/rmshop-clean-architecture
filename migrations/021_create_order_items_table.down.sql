ALTER TABLE order_items DROP CONSTRAINT IF EXISTS fk_order_items_order;
ALTER TABLE order_items DROP CONSTRAINT IF EXISTS fk_order_items_product;
DROP INDEX IF EXISTS idx_order_items_order_id;
DROP TABLE IF EXISTS order_items;