CREATE OR REPLACE FUNCTION reset_sub_categories_id_seq()
RETURNS void AS $$
DECLARE
    max_id INT;
BEGIN
    -- Get the maximum id from the table
    SELECT COALESCE(MAX(id), 0) INTO max_id FROM sub_categories;
    
    -- Reset the sequence to start from the next value after the maximum id
    EXECUTE 'ALTER SEQUENCE sub_categories_id_seq RESTART WITH ' || (max_id + 1)::TEXT;
END;
$$ LANGUAGE plpgsql;