ALTER TABLE user_address DROP CONSTRAINT IF EXISTS fk_user_address_user;
DROP INDEX IF EXISTS idx_user_address_user_id;
DROP INDEX IF EXISTS idx_unique_user_address;
DROP TABLE IF EXISTS user_address;