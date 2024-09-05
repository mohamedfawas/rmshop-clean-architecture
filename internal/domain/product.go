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

type ProductQueryParams struct {
	Page          int
	Limit         int
	Sort          string
	Order         string
	Category      string
	Subcategory   string
	Search        string
	MinPrice      float64
	MaxPrice      float64
	InStock       bool
	Brand         string
	Color         string
	Size          string
	MinRating     float64
	MinDiscount   float64
	MaxDiscount   float64
	CreatedAfter  string
	CreatedBefore string
	UpdatedAfter  string
	UpdatedBefore string
	Categories    []string
}

type PublicProduct struct {
	ID              int64     `json:"id"`
	Name            string    `json:"name"`
	Slug            string    `json:"slug"`
	Description     string    `json:"description"`
	Price           float64   `json:"price"`
	StockQuantity   int       `json:"stock_quantity"`
	CategoryName    string    `json:"category_name"`
	SubcategoryName string    `json:"subcategory_name"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
	Images          []string  `json:"images"`
}
