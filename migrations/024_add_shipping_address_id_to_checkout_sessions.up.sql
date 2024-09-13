ALTER TABLE checkout_sessions
ADD COLUMN shipping_address_id BIGINT;

ALTER TABLE checkout_sessions
ADD CONSTRAINT fk_checkout_sessions_shipping_address
FOREIGN KEY (shipping_address_id) 
REFERENCES shipping_addresses(id);

ALTER TABLE checkout_sessions
DROP COLUMN IF EXISTS address_id;