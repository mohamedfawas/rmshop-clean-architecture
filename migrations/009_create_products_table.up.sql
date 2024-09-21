CREATE TABLE IF NOT EXISTS products (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    price DECIMAL(10, 2) NOT NULL,
    stock_quantity INTEGER NOT NULL,
    sub_category_id INTEGER NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    slug VARCHAR(255) UNIQUE NOT NULL,
    primary_image_id INTEGER,
    is_deleted BOOLEAN NOT NULL DEFAULT FALSE,
    CONSTRAINT fk_sub_category
        FOREIGN KEY (sub_category_id)
        REFERENCES sub_categories(id)
        ON DELETE CASCADE
);

CREATE INDEX idx_products_sub_category ON products(sub_category_id);
CREATE INDEX idx_products_slug ON products(slug);
