ALTER TABLE orders DROP COLUMN IF EXISTS discount_amount;
ALTER TABLE orders DROP COLUMN IF EXISTS final_amount;
ALTER TABLE orders DROP COLUMN IF EXISTS coupon_applied;
ALTER TABLE orders DROP CONSTRAINT IF EXISTS check_final_amount;