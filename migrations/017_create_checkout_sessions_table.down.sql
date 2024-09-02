ALTER TABLE checkout_sessions DROP CONSTRAINT IF EXISTS fk_checkout_sessions_user;
ALTER TABLE checkout_items DROP CONSTRAINT IF EXISTS fk_checkout_items_session;
ALTER TABLE checkout_items DROP CONSTRAINT IF EXISTS fk_checkout_items_product;
DROP INDEX IF EXISTS idx_checkout_sessions_user_id;
DROP INDEX IF EXISTS idx_checkout_items_session_id;
DROP INDEX IF EXISTS idx_unique_session_product;

-- Drop checkout_items table
DROP TABLE IF EXISTS checkout_items;

-- Drop checkout_sessions table
DROP TABLE IF EXISTS checkout_sessions;