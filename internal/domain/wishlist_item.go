package domain

import (
	"time"
)

type WishlistItem struct {
	ID          int64     `json:"id"`
	UserID      int64     `json:"user_id"`
	ProductID   int64     `json:"product_id"`
	IsAvailable bool      `json:"is_available"`
	Price       float64   `json:"price"`
	CreatedAt   time.Time `json:"created_at"`
	ProductName string    `json:"product_name,omitempty"`
}
