ALTER TABLE wishlist_items
ADD COLUMN is_available BOOLEAN NOT NULL DEFAULT true,
ADD COLUMN price DECIMAL(10, 2) NOT NULL DEFAULT 0.00;