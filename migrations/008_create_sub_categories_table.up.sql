CREATE TABLE IF NOT EXISTS sub_categories (
    id SERIAL PRIMARY KEY,
    parent_category_id INTEGER NOT NULL,
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(255) NOT NULL UNIQUE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    CONSTRAINT fk_parent_category
        FOREIGN KEY (parent_category_id)
        REFERENCES categories(id)
        ON DELETE CASCADE
);

CREATE INDEX idx_sub_categories_slug ON sub_categories(slug);
CREATE INDEX idx_sub_categories_parent ON sub_categories(parent_category_id);
