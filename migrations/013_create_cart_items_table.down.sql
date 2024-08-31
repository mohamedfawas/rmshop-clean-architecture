ALTER TABLE cart_items DROP CONSTRAINT IF EXISTS fk_cart_items_product;
ALTER TABLE cart_items DROP CONSTRAINT IF EXISTS fk_cart_items_user;
ALTER TABLE cart_items DROP CONSTRAINT IF EXISTS unique_user_product;
DROP INDEX IF EXISTS idx_cart_items_user_id;
DROP TABLE IF EXISTS cart_items;