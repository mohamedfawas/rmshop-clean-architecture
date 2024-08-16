-- 002_add_last_login_to_users.down.sql
ALTER TABLE users DROP COLUMN IF EXISTS last_login;