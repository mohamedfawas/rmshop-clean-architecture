ALTER TABLE checkout_sessions
DROP CONSTRAINT IF EXISTS fk_checkout_sessions_address;

ALTER TABLE checkout_sessions
DROP COLUMN IF EXISTS address_id;