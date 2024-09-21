-- Create checkout_items table (if not exists)
CREATE TABLE IF NOT EXISTS checkout_items (
    id BIGSERIAL PRIMARY KEY,
    session_id BIGINT NOT NULL,
    product_id BIGINT NOT NULL,
    quantity INT NOT NULL,
    price DECIMAL(10, 2) NOT NULL,
    subtotal DECIMAL(10, 2) NOT NULL,
    CONSTRAINT fk_checkout_items_session FOREIGN KEY (session_id)
        REFERENCES checkout_sessions(id) ON DELETE CASCADE,
    CONSTRAINT fk_checkout_items_product FOREIGN KEY (product_id)
        REFERENCES products(id) ON DELETE RESTRICT
);

-- Create unique index to prevent duplicate products in the same session
CREATE UNIQUE INDEX IF NOT EXISTS idx_unique_session_product ON checkout_items(session_id, product_id);
