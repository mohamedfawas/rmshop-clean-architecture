-- Alter checkout_sessions table to add new columns
ALTER TABLE checkout_sessions
ADD COLUMN discount_amount DECIMAL(10, 2) NOT NULL DEFAULT 0,
ADD COLUMN final_amount DECIMAL(10, 2) NOT NULL DEFAULT 0,
ADD COLUMN coupon_code VARCHAR(20);

-- Update the existing rows to set final_amount equal to total_amount
UPDATE checkout_sessions SET final_amount = total_amount;

-- Add a check constraint to ensure final_amount is not negative
ALTER TABLE checkout_sessions
ADD CONSTRAINT check_final_amount_non_negative CHECK (final_amount >= 0);

-- Add a check constraint to ensure discount_amount is not negative
ALTER TABLE checkout_sessions
ADD CONSTRAINT check_discount_amount_non_negative CHECK (discount_amount >= 0);

-- Add a check constraint to ensure final_amount is less than or equal to total_amount
ALTER TABLE checkout_sessions
ADD CONSTRAINT check_final_amount_lte_total_amount CHECK (final_amount <= total_amount);