ALTER TABLE cancellation_requests DROP CONSTRAINT IF EXISTS fk_cancellation_requests_order;
ALTER TABLE cancellation_requests DROP CONSTRAINT IF EXISTS fk_cancellation_requests_user;
DROP INDEX IF EXISTS idx_cancellation_requests_order_id;
DROP INDEX IF EXISTS idx_cancellation_requests_user_id;
DROP INDEX IF EXISTS idx_cancellation_requests_status;
DROP TABLE IF EXISTS cancellation_requests;

CREATE TABLE cancellation_requests (
    id SERIAL PRIMARY KEY,
    order_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    cancellation_status VARCHAR(20) NOT NULL,
    CONSTRAINT fk_order
        FOREIGN KEY(order_id) 
        REFERENCES orders(id),
    CONSTRAINT fk_user
        FOREIGN KEY(user_id) 
        REFERENCES users(id),
    CONSTRAINT uq_order_cancellation
        UNIQUE (order_id)
);
