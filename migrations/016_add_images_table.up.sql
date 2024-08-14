-- Modify the products table
ALTER TABLE products
ADD COLUMN primary_image_id BIGINT;

-- drop image_url column from products table
ALTER TABLE products
DROP COLUMN image_url;

CREATE TABLE product_images (
    id SERIAL PRIMARY KEY,
    product_id BIGINT NOT NULL,
    image_url TEXT NOT NULL,
    created_at TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_product_images_product FOREIGN KEY (product_id) REFERENCES products (id) ON DELETE CASCADE
);

-- Add foreign key constraint to products table
ALTER TABLE products
ADD CONSTRAINT fk_products_primary_image
FOREIGN KEY (primary_image_id)
REFERENCES product_images (id);

-- Create index on product_id in product_images table
CREATE INDEX idx_product_images_product_id ON product_images (product_id);