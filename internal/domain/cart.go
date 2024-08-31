package domain

import "time"

type CartItem struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	ProductID int64     `json:"product_id"`
	Quantity  int       `json:"quantity"`
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

type Cart struct {
	Items      []*CartItemWithProduct `json:"items"`
	TotalValue float64                `json:"total_value"`
}
