package domain

import "time"

type TopProduct struct {
	ID            int64   `json:"id"`
	Name          string  `json:"name"`
	TotalQuantity int     `json:"total_quantity_sold"`
	TotalRevenue  float64 `json:"total_revenue"`
}

type TopCategory struct {
	ID         int     `json:"id"`
	Name       string  `json:"name"`
	TotalSales float64 `json:"total_sales"`
	OrderCount int     `json:"order_count"`
}

type TopCategoriesParams struct {
	StartDate time.Time
	EndDate   time.Time
	Limit     int
	SortOrder string
}

type SubcategoryAnalyticsParams struct {
	StartDate  time.Time
	EndDate    time.Time
	CategoryID int
	Page       int
	Limit      int
	SortBy     string
}

type TopSubcategory struct {
	ID            int     `json:"id"`
	Name          string  `json:"name"`
	CategoryID    int     `json:"category_id"`
	CategoryName  string  `json:"category_name"`
	TotalOrders   int     `json:"total_orders"`
	TotalQuantity int     `json:"total_quantity"`
	TotalRevenue  float64 `json:"total_revenue"`
}
