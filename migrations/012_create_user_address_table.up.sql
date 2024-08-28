CREATE TABLE IF NOT EXISTS user_address (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    address_line1 VARCHAR(255) NOT NULL,
    address_line2 VARCHAR(255), 
    state VARCHAR(100) NOT NULL,
    city VARCHAR(100) NOT NULL,
    pincode VARCHAR(20) NOT NULL,
    landmark VARCHAR(255),
    phone_number VARCHAR(15) NOT NULL,
    created_at TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITHOUT TIME ZONE,
    CONSTRAINT fk_user_address_user
        FOREIGN KEY (user_id)
        REFERENCES users(id)
        ON DELETE CASCADE
);

-- Create an index on user_id for faster lookups
CREATE INDEX idx_user_address_user_id ON user_address(user_id);

-- create a unique constraint to prevent duplicate addresses for user
CREATE UNIQUE INDEX idx_unique_user_address ON user_address(user_id, address_line1, city, state, pincode) WHERE deleted_at IS NULL;