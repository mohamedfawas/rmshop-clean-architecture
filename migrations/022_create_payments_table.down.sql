ALTER TABLE payments DROP CONSTRAINT IF EXISTS fk_payments_order;
DROP INDEX IF EXISTS idx_payments_razorpay_order_id;
DROP INDEX IF EXISTS idx_payments_order_id;
DROP TABLE IF EXISTS payments;