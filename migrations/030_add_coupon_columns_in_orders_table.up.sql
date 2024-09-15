ALTER TABLE orders
ADD COLUMN discount_amount DECIMAL(10, 2) DEFAULT 0.00,
ADD COLUMN final_amount DECIMAL(10, 2) DEFAULT 0.00,
ADD COLUMN coupon_applied BOOLEAN DEFAULT FALSE;

-- Add a check constraint to ensure final_amount is not greater than total_amount
ALTER TABLE orders
ADD CONSTRAINT check_final_amount
CHECK (final_amount <= total_amount);