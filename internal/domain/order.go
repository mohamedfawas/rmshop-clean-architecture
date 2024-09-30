package domain

import (
	"time"
)

type Order struct {
	ID                int64            `json:"id"`
	UserID            int64            `json:"user_id"`
	TotalAmount       float64          `json:"total_amount"`
	DiscountAmount    float64          `json:"discount_amount"`
	FinalAmount       float64          `json:"final_amount"`
	DeliveryStatus    string           `json:"delivery_status"`
	OrderStatus       string           `json:"order_status"`
	HasReturnRequest  bool             `json:"has_return_request"`
	ShippingAddressID int64            `json:"shipping_address_id"`
	ShippingAddress   *ShippingAddress `json:"shipping_address,omitempty"`
	CouponApplied     bool             `json:"coupon_applied"`
	CreatedAt         time.Time        `json:"created_at"`
	UpdatedAt         time.Time        `json:"updated_at"`
	DeliveredAt       *time.Time       `json:"delivered_at,omitempty"`
	Items             []OrderItem      `json:"items,omitempty"` // In some responses we don't need to display order items
	Payment           *Payment         `json:"payment,omitempty"`
}

type OrderResponse struct {
	OrderID        int64       `json:"order_id"`
	UserID         int64       `json:"user_id"`
	TotalAmount    float64     `json:"total_amount"`
	DiscountAmount float64     `json:"discount_amount"`
	FinalAmount    float64     `json:"final_amount"`
	OrderStatus    string      `json:"order_status"`
	DeliveryStatus string      `json:"delivery_status"`
	CreatedAt      time.Time   `json:"created_at"`
	UpdatedAt      time.Time   `json:"updated_at"`
	Items          []OrderItem `json:"items"`
	Payment        *Payment    `json:"payment"`
}

type OrderItem struct {
	ID          int64   `json:"id"`
	OrderID     int64   `json:"order_id"`
	ProductID   int64   `json:"product_id"`
	ProductName string  `json:"product_name"`
	Quantity    int     `json:"quantity"`
	Price       float64 `json:"price"`
}

type OrderCancellationResult struct {
	OrderID             int64
	OrderStatus         string
	RequiresAdminReview bool
	RefundInitiated     bool
}

type AdminOrderCancellationResult struct {
	OrderID             int64
	OrderStatus         string
	RequiresAdminReview bool
	RefundInitiated     bool
}

type CancellationRequest struct {
	ID                        int64
	OrderID                   int64
	UserID                    int64
	CreatedAt                 time.Time
	CancellationRequestStatus string
}

type OrderStatusUpdateResult struct {
	OrderID      int64  `json:"order_id"`
	NewStatus    string `json:"new_status"`
	RefundStatus string `json:"refund_status,omitempty"`
}

type OrderQueryParams struct {
	Page       int
	Limit      int
	SortBy     string
	SortOrder  string
	Status     string
	CustomerID int64
	StartDate  *time.Time
	EndDate    *time.Time
	Fields     []string
}

type RazorpayOrderResponse struct {
	OrderID     string `json:"order_id"`
	Amount      int64  `json:"amount"`
	Currency    string `json:"currency"`
	Description string `json:"description"`
}

type VerifyPaymentInput struct {
	OrderID           int64  `json:"order_id"`
	RazorpayPaymentID string `json:"razorpay_payment_id"`
	RazorpaySignature string `json:"razorpay_signature"`
}

type CreateOrderInput struct {
	UserID        int64       `json:"user_id"`
	Items         []OrderItem `json:"items"`
	AddressID     int64       `json:"address_id"`
	PaymentMethod string      `json:"payment_method"`
	CouponCode    string      `json:"coupon_code,omitempty"`
	Notes         string      `json:"notes,omitempty"`
}

type RazorpayPaymentInput struct {
	OrderID   string `json:"razorpay_order_id"`
	PaymentID string `json:"razorpay_payment_id"`
	Signature string `json:"razorpay_signature"`
}

type CancellationRequestParams struct {
	Page       int
	Limit      int
	SortBy     string
	SortOrder  string
	CustomerID int64
	StartDate  *time.Time
	EndDate    *time.Time
}
