-- Add is_deleted to users table
ALTER TABLE users ADD COLUMN is_deleted BOOLEAN NOT NULL DEFAULT FALSE;

-- Add is_deleted to categories table
ALTER TABLE categories ADD COLUMN is_deleted BOOLEAN NOT NULL DEFAULT FALSE;

-- Add is_deleted to sub_categories table
ALTER TABLE sub_categories ADD COLUMN is_deleted BOOLEAN NOT NULL DEFAULT FALSE;

-- Add is_deleted to products table
ALTER TABLE products ADD COLUMN is_deleted BOOLEAN NOT NULL DEFAULT FALSE;

-- Add is_deleted to user_address table
ALTER TABLE user_address ADD COLUMN is_deleted BOOLEAN NOT NULL DEFAULT FALSE;

-- Update existing records
UPDATE users SET is_deleted = (deleted_at IS NOT NULL);
UPDATE categories SET is_deleted = (deleted_at IS NOT NULL);
UPDATE sub_categories SET is_deleted = (deleted_at IS NOT NULL);
UPDATE products SET is_deleted = (deleted_at IS NOT NULL);
UPDATE user_address SET is_deleted = (deleted_at IS NOT NULL);