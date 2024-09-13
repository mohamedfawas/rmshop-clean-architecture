CREATE TABLE IF NOT EXISTS orders (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    total_amount DECIMAL(10, 2) NOT NULL,
    order_status VARCHAR(50) NOT NULL DEFAULT 'pending', -- e.g., 'pending', 'shipped', 'delivered'
    refund_status VARCHAR(50),  -- e.g., 'requested', 'approved', 'rejected'
    delivery_status VARCHAR(50) NOT NULL,
    shipping_address_id BIGINT, -- CHANGE THIS LOGIC LATER
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_orders_user FOREIGN KEY (user_id) REFERENCES users(id),
    CONSTRAINT fk_orders_address FOREIGN KEY (shipping_address_id) REFERENCES user_address(id)
);

CREATE INDEX idx_orders_user_id ON orders(user_id);