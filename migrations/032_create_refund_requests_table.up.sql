CREATE TABLE return_requests (
    id BIGSERIAL PRIMARY KEY,
    order_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    return_reason TEXT NOT NULL,
    is_approved BOOLEAN DEFAULT FALSE,
    requested_date TIMESTAMP WITH TIME ZONE NOT NULL,
    approved_at TIMESTAMP WITH TIME ZONE,
    rejected_at TIMESTAMP WITH TIME ZONE,
    refund_initiated BOOLEAN DEFAULT FALSE,
    refund_completed BOOLEAN DEFAULT FALSE,
    refund_amount DECIMAL(10, 2),
    CONSTRAINT fk_return_requests_order
        FOREIGN KEY (order_id)
        REFERENCES orders(id)
        ON DELETE CASCADE,
    CONSTRAINT fk_return_requests_user
        FOREIGN KEY (user_id)
        REFERENCES users(id)
        ON DELETE CASCADE
);

-- Indexes
CREATE INDEX idx_return_requests_order_id ON return_requests(order_id);
CREATE INDEX idx_return_requests_user_id ON return_requests(user_id);
CREATE INDEX idx_return_requests_is_approved ON return_requests(is_approved);