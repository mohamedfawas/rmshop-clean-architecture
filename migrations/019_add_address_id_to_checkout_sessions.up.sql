ALTER TABLE checkout_sessions
ADD COLUMN address_id BIGINT;

ALTER TABLE checkout_sessions
ADD CONSTRAINT fk_checkout_sessions_address
FOREIGN KEY (address_id) 
REFERENCES user_address(id);