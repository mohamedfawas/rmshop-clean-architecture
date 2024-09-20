ALTER TABLE wallet_transactions DROP CONSTRAINT IF EXISTS fk_wallet_transactions_user;
DROP INDEX IF EXISTS idx_wallet_transactions_user_id;
DROP INDEX IF EXISTS idx_wallet_transactions_created_at;
DROP TABLE IF EXISTS wallet_transactions;
