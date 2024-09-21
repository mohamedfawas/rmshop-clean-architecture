ALTER TABLE checkout_sessions DROP CONSTRAINT IF EXISTS fk_checkout_sessions_user;
ALTER TABLE checkout_sessions DROP CONSTRAINT IF EXISTS fk_checkout_sessions_shipping_address;
ALTER TABLE checkout_sessions DROP CONSTRAINT IF EXISTS check_final_amount_non_negative;
ALTER TABLE checkout_sessions DROP CONSTRAINT IF EXISTS check_discount_amount_non_negative;
ALTER TABLE checkout_sessions DROP CONSTRAINT IF EXISTS check_final_amount_lte_total_amount;

DROP INDEX IF EXISTS idx_checkout_sessions_user_id;
DROP INDEX IF EXISTS idx_checkout_sessions_status;

DROP TABLE IF EXISTS checkout_sessions;