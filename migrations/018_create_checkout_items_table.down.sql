ALTER TABLE checkout_items DROP CONSTRAINT IF EXISTS fk_checkout_items_session;
ALTER TABLE checkout_items DROP CONSTRAINT IF EXISTS fk_checkout_items_product;

DROP INDEX IF EXISTS idx_unique_session_product;

DROP TABLE IF EXISTS checkout_items;