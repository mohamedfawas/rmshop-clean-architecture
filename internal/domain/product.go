package domain

import (
	"time"
)

type Product struct {
	ID            int64      `json:"id"`
	Name          string     `json:"name"`
	Description   string     `json:"description"`
	Price         float64    `json:"price"`
	StockQuantity int        `json:"stock_quantity"`
	CategoryID    int        `json:"category_id"`
	SubCategoryID int        `json:"sub_category_id"`
	ImageURL      string     `json:"image_url"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
	DeletedAt     *time.Time `json:"deleted_at,omitempty"`
}
