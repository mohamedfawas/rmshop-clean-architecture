-- Create checkout_sessions table
CREATE TABLE IF NOT EXISTS checkout_sessions (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    total_amount DECIMAL(10, 2) NOT NULL DEFAULT 0,
    item_count INT NOT NULL DEFAULT 0,
    status VARCHAR(50) NOT NULL,
    coupon_applied BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_checkout_sessions_user
        FOREIGN KEY (user_id)
        REFERENCES users(id)
        ON DELETE CASCADE
);

-- Create index on user_id for faster lookups
CREATE INDEX idx_checkout_sessions_user_id ON checkout_sessions(user_id);

-- Create checkout_items table
CREATE TABLE IF NOT EXISTS checkout_items (
    id BIGSERIAL PRIMARY KEY,
    session_id BIGINT NOT NULL,
    product_id BIGINT NOT NULL,
    quantity INT NOT NULL,
    price DECIMAL(10, 2) NOT NULL,
    subtotal DECIMAL(10, 2) NOT NULL,
    CONSTRAINT fk_checkout_items_session
        FOREIGN KEY (session_id)
        REFERENCES checkout_sessions(id)
        ON DELETE CASCADE,
    CONSTRAINT fk_checkout_items_product
        FOREIGN KEY (product_id)
        REFERENCES products(id)
        ON DELETE RESTRICT
);

-- Create index on session_id for faster lookups
CREATE INDEX idx_checkout_items_session_id ON checkout_items(session_id);

-- Create a unique constraint to prevent duplicate products in a session
CREATE UNIQUE INDEX idx_unique_session_product ON checkout_items(session_id, product_id);