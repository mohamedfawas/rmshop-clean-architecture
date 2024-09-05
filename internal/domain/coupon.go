package domain

import "time"

type Coupon struct {
	ID                 int64      `json:"id"`
	Code               string     `json:"code"`
	DiscountPercentage float64    `json:"discount_percentage"`
	MinOrderAmount     float64    `json:"min_order_amount"`
	IsActive           bool       `json:"is_active"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
	ExpiresAt          *time.Time `json:"expires_at,omitempty"`
	DeletedAt          *time.Time `json:"deleted_at,omitempty"`
	IsDeleted          bool       `json:"is_deleted"`
}

type CreateCouponInput struct {
	Code               string  `json:"code"`
	DiscountPercentage float64 `json:"discount_percentage"`
	MinOrderAmount     float64 `json:"min_order_amount"`
	ExpiresAt          string  `json:"expires_at,omitempty"`
}

type CouponQueryParams struct {
	Page        int
	Limit       int
	Sort        string
	Order       string
	Status      string
	Search      string
	MinDiscount *float64
	MaxDiscount *float64
	CurrentTime time.Time
}

type CouponUpdateInput struct {
	Code               *string  `json:"code"`
	DiscountPercentage *float64 `json:"discount_percentage"`
	MinOrderAmount     *float64 `json:"min_order_amount"`
	ExpiresAt          *string  `json:"expires_at"`
}
