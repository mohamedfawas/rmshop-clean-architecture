package domain

import "time"

type CheckoutSession struct {
	ID            int64     `json:"id"`
	UserID        int64     `json:"user_id"`
	TotalAmount   float64   `json:"total_amount"`
	ItemCount     int       `json:"item_count"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	Status        string    `json:"status"`
	CouponApplied bool      `json:"coupon_applied"`
}

type CheckoutItem struct {
	ID        int64   `json:"id"`
	SessionID int64   `json:"session_id"`
	ProductID int64   `json:"product_id"`
	Quantity  int     `json:"quantity"`
	Price     float64 `json:"price"`
	Subtotal  float64 `json:"subtotal"`
}
