ALTER TABLE sub_categories DROP CONSTRAINT IF EXISTS fk_parent_category;
DROP INDEX IF EXISTS idx_sub_categories_slug;
DROP INDEX IF EXISTS idx_sub_categories_parent;
DROP TABLE IF EXISTS sub_categories;
