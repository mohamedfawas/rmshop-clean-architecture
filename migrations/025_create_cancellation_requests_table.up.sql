CREATE TABLE cancellation_requests (
    id SERIAL PRIMARY KEY,
    order_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    cancellation_status VARCHAR(20) NOT NULL,
    is_stock_updated BOOLEAN NOT NULL DEFAULT FALSE,
    CONSTRAINT fk_order
        FOREIGN KEY(order_id) 
        REFERENCES orders(id),
    CONSTRAINT fk_user
        FOREIGN KEY(user_id) 
        REFERENCES users(id),
    CONSTRAINT uq_order_cancellation
        UNIQUE (order_id)
);
