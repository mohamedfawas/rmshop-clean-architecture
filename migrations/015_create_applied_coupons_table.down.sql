ALTER TABLE applied_coupons DROP CONSTRAINT IF EXISTS applied_coupons_user_id_fkey;
ALTER TABLE applied_coupons DROP CONSTRAINT IF EXISTS applied_coupons_coupon_id_fkey;
DROP TABLE IF EXISTS applied_coupons;