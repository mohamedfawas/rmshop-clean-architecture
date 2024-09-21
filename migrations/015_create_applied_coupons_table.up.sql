CREATE TABLE IF NOT EXISTS applied_coupons (
    user_id BIGINT PRIMARY KEY,
    coupon_id INTEGER NOT NULL,
    applied_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT applied_coupons_user_id_fkey
        FOREIGN KEY (user_id) 
        REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT applied_coupons_coupon_id_fkey
        FOREIGN KEY (coupon_id) 
        REFERENCES coupons(id) ON DELETE CASCADE
);