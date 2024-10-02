package domain

import "time"

type ReturnRequest struct {
	ID                      int64      `json:"id"`
	OrderID                 int64      `json:"order_id"`
	UserID                  int64      `json:"user_id"`
	ReturnReason            string     `json:"return_reason"`
	IsApproved              bool       `json:"is_approved"`
	RequestedDate           time.Time  `json:"requested_date"`
	ApprovedAt              *time.Time `json:"approved_at,omitempty"`
	RejectedAt              *time.Time `json:"rejected_at,omitempty"`
	RefundInitiated         bool       `json:"refund_initiated"`
	RefundAmount            *float64   `json:"refund_amount,omitempty"`
	OrderReturnedToSellerAt *time.Time `json:"order_returned_to_seller_at,omitempty"`
	IsOrderReachedTheSeller bool       `json:"is_order_reached_the_seller"`
	IsStockUpdated          bool       `json:"is_stock_updated"`
}
