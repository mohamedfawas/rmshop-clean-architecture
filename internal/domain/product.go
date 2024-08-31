package domain

import (
	"time"
)

type Product struct {
	ID             int64      `json:"id"`
	Name           string     `json:"name"`
	Slug           string     `json:"slug"`
	Description    string     `json:"description"`
	Price          float64    `json:"price"`
	StockQuantity  int        `json:"stock_quantity"`
	SubCategoryID  int        `json:"sub_category_id"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
	DeletedAt      *time.Time `json:"deleted_at,omitempty"`
	PrimaryImageID *int64     `json:"primary_image_id,omitempty"`
	IsDeleted      bool       `json:"is_deleted"`
}
