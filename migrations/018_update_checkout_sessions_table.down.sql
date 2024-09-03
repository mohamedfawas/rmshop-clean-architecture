-- Remove the check constraints
ALTER TABLE checkout_sessions
DROP CONSTRAINT IF EXISTS check_final_amount_non_negative,
DROP CONSTRAINT IF EXISTS check_discount_amount_non_negative,
DROP CONSTRAINT IF EXISTS check_final_amount_lte_total_amount;

-- Remove the new columns
ALTER TABLE checkout_sessions
DROP COLUMN IF EXISTS discount_amount,
DROP COLUMN IF EXISTS final_amount,
DROP COLUMN IF EXISTS coupon_code;