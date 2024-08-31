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
