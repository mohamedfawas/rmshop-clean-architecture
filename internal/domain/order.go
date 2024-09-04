package domain

import "time"

type Order struct {
	ID             int64     `json:"id"`
	UserID         int64     `json:"user_id"`
	TotalAmount    float64   `json:"total_amount"`
	PaymentMethod  string    `json:"payment_method"`
	PaymentStatus  string    `json:"payment_status"`
	DeliveryStatus string    `json:"delivery_status"`
	AddressID      int64     `json:"address_id"`
	CreatedAt      time.Time `json:"created_at"`
}

type OrderItem struct {
	ID        int64   `json:"id"`
	OrderID   int64   `json:"order_id"`
	ProductID int64   `json:"product_id"`
	Quantity  int     `json:"quantity"`
	Price     float64 `json:"price"`
}
