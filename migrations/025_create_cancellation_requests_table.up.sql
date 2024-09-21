CREATE TABLE IF NOT EXISTS cancellation_requests (
    id BIGSERIAL PRIMARY KEY,
    order_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    status VARCHAR(50) NOT NULL, --  The current status of the cancellation request (e.g., 'pending_review', 'approved', 'rejected').
    processed_at TIMESTAMP WITH TIME ZONE,
    CONSTRAINT fk_cancellation_requests_order
        FOREIGN KEY (order_id)
        REFERENCES orders(id)
        ON DELETE CASCADE,
    CONSTRAINT fk_cancellation_requests_user
        FOREIGN KEY (user_id)
        REFERENCES users(id)
        ON DELETE CASCADE
);

CREATE INDEX idx_cancellation_requests_order_id ON cancellation_requests(order_id);
CREATE INDEX idx_cancellation_requests_user_id ON cancellation_requests(user_id);
CREATE INDEX idx_cancellation_requests_status ON cancellation_requests(status);
