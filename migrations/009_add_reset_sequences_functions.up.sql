-- Function to reset categories sequence
CREATE OR REPLACE FUNCTION reset_categories_id_seq()
RETURNS void AS $$
DECLARE
    max_id INT;
BEGIN
    SELECT COALESCE(MAX(id), 0) INTO max_id FROM categories;
    EXECUTE 'ALTER SEQUENCE categories_id_seq RESTART WITH ' || (max_id + 1)::TEXT;
END;
$$ LANGUAGE plpgsql;

-- Function to reset users sequence
CREATE OR REPLACE FUNCTION reset_users_id_seq()
RETURNS void AS $$
DECLARE
    max_id BIGINT;
BEGIN
    SELECT COALESCE(MAX(id), 0) INTO max_id FROM users;
    EXECUTE 'ALTER SEQUENCE users_id_seq RESTART WITH ' || (max_id + 1)::TEXT;
END;
$$ LANGUAGE plpgsql;