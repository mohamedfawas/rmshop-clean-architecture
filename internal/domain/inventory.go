package domain

type InventoryItem struct {
	ID            int64   `json:"id,omitempty"`
	ProductID     int64   `json:"product_id"`
	ProductName   string  `json:"product_name"`
	CategoryID    int64   `json:"category_id"`
	CategoryName  string  `json:"category_name"`
	StockQuantity int     `json:"stock_quantity"`
	Price         float64 `json:"price"`
}

type InventoryQueryParams struct {
	ProductID     int64
	ProductName   string
	CategoryID    int64
	CategoryName  string
	StockLessThan *int
	StockMoreThan *int
	Page          int
	Limit         int
	SortBy        string
	SortOrder     string
}
