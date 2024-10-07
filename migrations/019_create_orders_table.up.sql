-- Create orders table
CREATE TABLE IF NOT EXISTS orders (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    total_amount DECIMAL(10, 2) NOT NULL,
    discount_amount DECIMAL(10, 2) NOT NULL DEFAULT 0.00,
    final_amount DECIMAL(10, 2) NOT NULL,
    shipping_address_id BIGINT,
    coupon_applied BOOLEAN NOT NULL DEFAULT FALSE,
    has_return_request BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    delivered_at TIMESTAMP WITH TIME ZONE,
    order_status VARCHAR(20) CHECK (order_status IN (
        'pending_payment',
        'processing',
        'shipped',
        'completed',
        'cancelled',
        'refunded',
        'confirmed',
        'pending_cancellation',
        'return_approved'
    )) DEFAULT 'pending_payment',
    delivery_status VARCHAR(20) CHECK (delivery_status IN (
        'pending',
        'in_transit',
        'out_for_delivery',
        'delivered',
        'failed_attempt',
        'returned_to_sender'
    )) DEFAULT 'pending',
    is_cancelled BOOLEAN NOT NULL DEFAULT FALSE,
    CONSTRAINT fk_orders_user FOREIGN KEY (user_id) REFERENCES users(id),
    CONSTRAINT fk_orders_address FOREIGN KEY (shipping_address_id) REFERENCES shipping_addresses(id),
    CONSTRAINT check_final_amount_non_negative CHECK (final_amount >= 0),
    CONSTRAINT check_discount_amount_non_negative CHECK (discount_amount >= 0),
    CONSTRAINT check_final_amount_lte_total_amount CHECK (final_amount <= total_amount)
);

-- Create indexes for faster queries
CREATE INDEX idx_orders_user_id ON orders(user_id);
CREATE INDEX idx_orders_order_status ON orders(order_status);
CREATE INDEX idx_orders_delivery_status ON orders(delivery_status);
CREATE INDEX idx_orders_created_at ON orders(created_at);
CREATE INDEX idx_orders_is_cancelled ON orders(is_cancelled);