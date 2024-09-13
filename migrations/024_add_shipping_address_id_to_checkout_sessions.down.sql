ALTER TABLE checkout_sessions
DROP CONSTRAINT fk_checkout_sessions_shipping_address;

ALTER TABLE checkout_sessions
DROP COLUMN shipping_address_id;

ALTER TABLE checkout_sessions
ADD COLUMN IF NOT EXISTS address_id;