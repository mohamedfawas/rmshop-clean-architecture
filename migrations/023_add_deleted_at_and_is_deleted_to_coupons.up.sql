ALTER TABLE coupons
ADD COLUMN deleted_at TIMESTAMP WITH TIME ZONE,
ADD COLUMN is_deleted BOOLEAN NOT NULL DEFAULT FALSE;

-- Update existing rows to set is_deleted based on is_active
UPDATE coupons SET is_deleted = NOT is_active;