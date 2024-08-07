-- 011_add_updated_at_to_categories.up.sql
ALTER TABLE categories 
ADD COLUMN updated_at TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP;

-- Update existing rows to set updated_at equal to created_at
UPDATE categories SET updated_at = created_at;