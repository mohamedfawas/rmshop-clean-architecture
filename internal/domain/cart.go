package domain

import "time"

type CartItem struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	ProductID int64     `json:"product_id"`
	Quantity  int       `json:"quantity"`
	Price     float64   `json:"price"`
	Subtotal  float64   `json:"subtotal"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type AddToCartInput struct {
	ProductID int64 `json:"product_id"`
	Quantity  int   `json:"quantity"`
}

type CartItemWithProduct struct {
	CartItem
	ProductName  string  `json:"product_name"`
	ProductPrice float64 `json:"product_price"`
}

type CartResponse struct {
	Items      []*CartItem `json:"cart_items"`
	TotalValue float64     `json:"total_value"`
}

type UpdatedCartItemResponse struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	ProductID int64     `json:"product_id"`
	Quantity  int       `json:"quantity"`
	Price     float64   `json:"price"`
	Subtotal  float64   `json:"subtotal"`
	UpdatedAt time.Time `json:"updated_at"`
}
