-- Create return_requests table
CREATE TABLE IF NOT EXISTS return_requests (
    id BIGSERIAL PRIMARY KEY,
    order_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    return_reason TEXT NOT NULL,
    is_approved BOOLEAN DEFAULT FALSE,
    requested_date TIMESTAMP WITH TIME ZONE NOT NULL,
    approved_at TIMESTAMP WITH TIME ZONE,
    rejected_at TIMESTAMP WITH TIME ZONE,
    refund_initiated BOOLEAN DEFAULT FALSE,
    refund_amount DECIMAL(10, 2),
    order_returned_to_seller_at TIMESTAMP WITH TIME ZONE,
    is_order_reached_the_seller BOOLEAN DEFAULT FALSE,
    is_stock_updated BOOLEAN NOT NULL DEFAULT FALSE,
    CONSTRAINT fk_return_requests_order
        FOREIGN KEY (order_id)
        REFERENCES orders(id)
        ON DELETE CASCADE,
    CONSTRAINT fk_return_requests_user
        FOREIGN KEY (user_id)
        REFERENCES users(id)
        ON DELETE CASCADE
);

-- Create indexes
CREATE INDEX idx_return_requests_order_id ON return_requests(order_id);
CREATE INDEX idx_return_requests_user_id ON return_requests(user_id);
CREATE INDEX idx_return_requests_is_approved ON return_requests(is_approved);
