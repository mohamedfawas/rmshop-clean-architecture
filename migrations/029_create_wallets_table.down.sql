ALTER TABLE wallets DROP CONSTRAINT IF EXISTS fk_wallets_user;
DROP INDEX IF EXISTS idx_wallets_user_id;
DROP TABLE IF EXISTS wallets;