-- Remove is_deleted from users table
ALTER TABLE users DROP COLUMN IF EXISTS is_deleted;

-- Remove is_deleted from categories table
ALTER TABLE categories DROP COLUMN IF EXISTS is_deleted;

-- Remove is_deleted from sub_categories table
ALTER TABLE sub_categories DROP COLUMN IF EXISTS is_deleted;

-- Remove is_deleted from products table
ALTER TABLE products DROP COLUMN IF EXISTS is_deleted;

-- Remove is_deleted from user_address table
ALTER TABLE user_address DROP COLUMN IF EXISTS is_deleted;