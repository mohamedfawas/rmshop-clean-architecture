package domain

type TopProduct struct {
	ID            int64   `json:"id"`
	Name          string  `json:"name"`
	TotalQuantity int     `json:"total_quantity_sold"`
	TotalRevenue  float64 `json:"total_revenue"`
}
