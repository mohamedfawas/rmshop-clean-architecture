package domain

import "time"

type SalesReport struct {
	ReportType            string              `json:"report_type"`
	StartDate             time.Time           `json:"start_date"`
	EndDate               time.Time           `json:"end_date"`
	CouponApplied         bool                `json:"coupon_applied"`
	TotalOrderCount       int                 `json:"total_order_count"`
	TotalAmount           float64             `json:"total_amount"`
	TotalCouponOrderCount int                 `json:"total_coupon_order_count"`
	AverageOrderValue     float64             `json:"average_order_value,omitempty"`
	TopSellingProducts    []TopSellingProduct `json:"top_selling_products,omitempty"`
	PDFReport             []byte              `json:"-"`
	ExcelReport           []byte              `json:"-"`
}

type DailySales struct {
	Date             time.Time `json:"date"`
	OrderCount       int       `json:"order_count"`
	TotalAmount      float64   `json:"total_amount"`
	CouponOrderCount int       `json:"coupon_order_count"`
}

type TopSellingProduct struct {
	ID       int64   `json:"id"`
	Name     string  `json:"name"`
	Quantity int     `json:"quantity"`
	Revenue  float64 `json:"revenue"`
}

type WeeklySalesReport struct {
	StartDate             time.Time           `json:"start_date"`
	EndDate               time.Time           `json:"end_date"`
	DailySales            []DailySales        `json:"daily_sales"`
	TotalOrderCount       int                 `json:"total_order_count"`
	TotalAmount           float64             `json:"total_amount"`
	TotalCouponOrderCount int                 `json:"total_coupon_order_count"`
	AverageOrderValue     float64             `json:"average_order_value"`
	TopSellingProducts    []TopSellingProduct `json:"top_selling_products,omitempty"`
}

type MonthlySalesReport struct {
	Year                  int          `json:"year"`
	Month                 time.Month   `json:"month"`
	DailySales            []DailySales `json:"daily_sales"`
	TotalOrderCount       int          `json:"total_order_count"`
	TotalAmount           float64      `json:"total_amount"`
	TotalCouponOrderCount int          `json:"total_coupon_order_count"`
}

type CustomSalesReport struct {
	StartDate             time.Time    `json:"start_date"`
	EndDate               time.Time    `json:"end_date"`
	DailySales            []DailySales `json:"daily_sales"`
	TotalOrderCount       int          `json:"total_order_count"`
	TotalAmount           float64      `json:"total_amount"`
	TotalCouponOrderCount int          `json:"total_coupon_order_count"`
}
