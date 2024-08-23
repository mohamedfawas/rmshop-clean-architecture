ALTER TABLE products DROP CONSTRAINT IF EXISTS fk_products_primary_image;
ALTER TABLE product_images DROP CONSTRAINT IF EXISTS fk_product_images_product;
ALTER TABLE product_images DROP CONSTRAINT IF EXISTS uq_product_image_url;
DROP INDEX IF EXISTS idx_product_images_product_id;
DROP TABLE IF EXISTS product_images;
