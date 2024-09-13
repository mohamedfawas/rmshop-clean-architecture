ALTER TABLE shipping_addresses DROP CONSTRAINT IF EXISTS fk_shipping_addresses_user;
ALTER TABLE shipping_addresses DROP CONSTRAINT IF EXISTS fk_shipping_addresses_address;
ALTER TABLE shipping_addresses DROP CONSTRAINT IF EXISTS unique_user_address;
-- Drop the indexes first
DROP INDEX IF EXISTS idx_shipping_addresses_user_id;
DROP INDEX IF EXISTS idx_shipping_addresses_address_id;

-- Drop the table after the indexes are removed
DROP TABLE IF EXISTS shipping_addresses;
