package domain

import (
	"time"
)

type Product struct {
	ID             int64          `json:"id"`
	Name           string         `json:"name"`
	Description    string         `json:"description"`
	Price          float64        `json:"price"`
	StockQuantity  *int           `json:"stock_quantity"`
	CategoryID     int            `json:"category_id"`
	SubCategoryID  int            `json:"sub_category_id"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      *time.Time     `json:"deleted_at,omitempty"`
	PrimaryImageID *int64         `json:"primary_image_id,omitempty"`
	Images         []ProductImage `json:"images,omitempty"`
}

type ProductImage struct {
	ID        int64     `json:"id"`
	ProductID int64     `json:"product_id"`
	ImageURL  string    `json:"image_url"`
	IsPrimary bool      `json:"is_primary"`
	CreatedAt time.Time `json:"created_at"`
}
