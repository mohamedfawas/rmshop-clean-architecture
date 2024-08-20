CREATE TABLE IF NOT EXISTS product_images (
    id SERIAL PRIMARY KEY,
    product_id BIGINT NOT NULL,
    image_url TEXT NOT NULL,
    is_primary BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_product_images_product FOREIGN KEY (product_id) REFERENCES products (id) ON DELETE CASCADE
);

ALTER TABLE products
ADD CONSTRAINT fk_products_primary_image
FOREIGN KEY (primary_image_id)
REFERENCES product_images (id);

CREATE INDEX idx_product_images_product_id ON product_images (product_id);
