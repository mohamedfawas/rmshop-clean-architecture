-- 002_add_last_login_to_users.up.sql
ALTER TABLE users ADD COLUMN last_login TIMESTAMP WITHOUT TIME ZONE;