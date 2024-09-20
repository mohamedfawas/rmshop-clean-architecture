package domain

import "time"

type Refund struct {
	ID        int64     `json:"id"`
	OrderID   int64     `json:"order_id"`
	Amount    float64   `json:"amount"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type RefundDetails struct {
	ReturnID     int64     `json:"return_id"`
	OrderID      int64     `json:"order_id"`
	RefundAmount float64   `json:"refund_amount"`
	RefundStatus string    `json:"refund_status"`
	RefundedAt   time.Time `json:"refunded_at"`
}
