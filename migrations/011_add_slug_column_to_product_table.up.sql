ALTER TABLE products ADD COLUMN slug VARCHAR(255) UNIQUE NOT NULL;
CREATE INDEX idx_products_slug ON products(slug);