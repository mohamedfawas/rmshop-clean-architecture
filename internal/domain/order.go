package domain

import (
	"database/sql"
	"time"
)

type Order struct {
	ID             int64          `json:"id"`
	UserID         int64          `json:"user_id"`
	TotalAmount    float64        `json:"total_amount"`
	PaymentMethod  string         `json:"payment_method"`
	PaymentStatus  string         `json:"payment_status"`
	DeliveryStatus string         `json:"delivery_status"`
	OrderStatus    string         `json:"order_status"`
	RefundStatus   sql.NullString `json:"refund_status"`
	AddressID      int64          `json:"address_id"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	Items          []OrderItem    `json:"items"`
}

type OrderItem struct {
	ID        int64   `json:"id"`
	OrderID   int64   `json:"order_id"`
	ProductID int64   `json:"product_id"`
	Quantity  int     `json:"quantity"`
	Price     float64 `json:"price"`
}

type OrderCancellationResult struct {
	OrderID      int64  `json:"order_id"`
	RefundStatus string `json:"refund_status"`
}
