ALTER TABLE orders DROP CONSTRAINT IF EXISTS fk_orders_user;
ALTER TABLE orders DROP CONSTRAINT IF EXISTS fk_orders_address;
ALTER TABLE orders DROP CONSTRAINT IF EXISTS check_final_amount_non_negative;
ALTER TABLE orders DROP CONSTRAINT IF EXISTS check_discount_amount_non_negative;
ALTER TABLE orders DROP CONSTRAINT IF EXISTS check_final_amount_lte_total_amount;

DROP INDEX IF EXISTS idx_orders_user_id;
DROP INDEX IF EXISTS idx_orders_order_status;
DROP INDEX IF EXISTS idx_orders_delivery_status;
DROP INDEX IF EXISTS idx_orders_created_at;

DROP TABLE IF EXISTS orders;