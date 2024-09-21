ALTER TABLE wishlist_items DROP CONSTRAINT IF EXISTS fk_wishlist_user;
ALTER TABLE wishlist_items DROP CONSTRAINT IF EXISTS fk_wishlist_product;
ALTER TABLE wishlist_items DROP CONSTRAINT IF EXISTS uq_wishlist_user_product;

DROP INDEX IF EXISTS idx_wishlist_user_id;

DROP TABLE IF EXISTS wishlist_items;
