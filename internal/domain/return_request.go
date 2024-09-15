package domain

import "time"

type ReturnRequest struct {
	ID              int64      `json:"id"`
	OrderID         int64      `json:"order_id"`
	UserID          int64      `json:"user_id"`
	ReturnReason    string     `json:"return_reason"`
	IsApproved      bool       `json:"is_approved"`
	RequestedDate   time.Time  `json:"requested_date"`
	ApprovedAt      *time.Time `json:"approved_at,omitempty"`
	RejectedAt      *time.Time `json:"rejected_at,omitempty"`
	RefundInitiated bool       `json:"refund_initiated"`
	RefundCompleted bool       `json:"refund_completed"`
	RefundAmount    *float64   `json:"refund_amount,omitempty"`
}
