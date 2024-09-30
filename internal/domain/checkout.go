package domain

import "time"

type CheckoutSession struct {
	ID                int64            `json:"id"`
	UserID            int64            `json:"user_id"`
	TotalAmount       float64          `json:"total_amount"`
	DiscountAmount    float64          `json:"discount_amount"`
	FinalAmount       float64          `json:"final_amount"`
	ItemCount         int              `json:"item_count"`
	CreatedAt         time.Time        `json:"created_at"`
	UpdatedAt         time.Time        `json:"updated_at"`
	Status            string           `json:"status"`
	CouponCode        string           `json:"coupon_code,omitempty"`
	CouponApplied     bool             `json:"coupon_applied"`
	ShippingAddressID int64            `json:"shipping_address_id,omitempty"`
	ShippingAddress   *ShippingAddress `json:"shipping_address,omitempty"`
}

type CheckoutItem struct {
	ID        int64   `json:"id"`
	SessionID int64   `json:"session_id"`
	ProductID int64   `json:"product_id"`
	Quantity  int     `json:"quantity"`
	Price     float64 `json:"price"`
	Subtotal  float64 `json:"subtotal"`
}

type ApplyCouponInput struct {
	CouponCode string `json:"coupon_code"`
}

type ApplyCouponResponse struct {
	CheckoutSession
	Message string `json:"message,omitempty"`
}

type AddressInput struct {
	AddressID  int64        `json:"address_id,omitempty"`
	NewAddress *UserAddress `json:"new_address,omitempty"`
}

type CheckoutSummary struct {
	ID             int64                                     `json:"id"`
	UserID         int64                                     `json:"user_id"`
	TotalAmount    float64                                   `json:"total_amount"`
	DiscountAmount float64                                   `json:"discount_amount"`
	FinalAmount    float64                                   `json:"final_amount"`
	ItemCount      int                                       `json:"item_count"`
	Status         string                                    `json:"status"`
	CouponCode     string                                    `json:"coupon_code,omitempty"`
	CouponApplied  bool                                      `json:"coupon_applied"`
	Address        *ShippingAddressResponseInCheckoutSummary `json:"address,omitempty"`
	Items          []*CheckoutItemDetail                     `json:"items"`
}

type CheckoutItemDetail struct {
	ID        int64   `json:"id"`
	ProductID int64   `json:"product_id"`
	Name      string  `json:"name"`
	Quantity  int     `json:"quantity"`
	Price     float64 `json:"price"`
	Subtotal  float64 `json:"subtotal"`
}
