-- Create checkout_sessions table (if not exists)
CREATE TABLE IF NOT EXISTS checkout_sessions (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    total_amount DECIMAL(10, 2) NOT NULL DEFAULT 0,
    item_count INT NOT NULL DEFAULT 0,
    status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN (
        'pending', 'payment_initiated', 'payment_failed', 'completed', 'abandoned', 'expired'
    )),
    coupon_applied BOOLEAN NOT NULL DEFAULT FALSE,
    coupon_code VARCHAR(20),
    discount_amount DECIMAL(10, 2) NOT NULL DEFAULT 0 CHECK (discount_amount >= 0),
    final_amount DECIMAL(10, 2) NOT NULL DEFAULT 0 CHECK (final_amount >= 0 AND final_amount <= total_amount),
    shipping_address_id BIGINT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_checkout_sessions_user FOREIGN KEY (user_id)
        REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT fk_checkout_sessions_shipping_address FOREIGN KEY (shipping_address_id)
        REFERENCES shipping_addresses(id),
    CONSTRAINT check_final_amount_non_negative
         CHECK (final_amount >= 0),
    CONSTRAINT check_discount_amount_non_negative 
        CHECK (discount_amount >= 0),
    CONSTRAINT check_final_amount_lte_total_amount 
        CHECK (final_amount <= total_amount)
);

-- Create indexes for faster lookups
CREATE INDEX IF NOT EXISTS idx_checkout_sessions_user_id ON checkout_sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_checkout_sessions_status ON checkout_sessions(status);