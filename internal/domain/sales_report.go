package domain

import "time"

type SalesReport struct {
	ReportType            string              `json:"report_type"`
	StartDate             time.Time           `json:"start_date"`
	EndDate               time.Time           `json:"end_date"`
	CouponApplied         bool                `json:"coupon_applied"`
	DailyData             []DailySalesData    `json:"daily_data,omitempty"`
	WeeklyData            []WeeklySalesData   `json:"weekly_data,omitempty"`
	MonthlyData           []MonthlySalesData  `json:"monthly_data,omitempty"`
	YearlyData            []YearlySalesData   `json:"yearly_data,omitempty"`
	CustomData            CustomSalesData     `json:"custom_data,omitempty"`
	TotalOrderCount       int                 `json:"total_order_count"`
	TotalAmount           float64             `json:"total_amount"`
	TotalCouponOrderCount int                 `json:"total_coupon_order_count"`
	AverageOrderValue     float64             `json:"average_order_value,omitempty"`
	TopSellingProducts    []TopSellingProduct `json:"top_selling_products,omitempty"`
}

type DailySalesData struct {
	Date             time.Time `json:"date"`
	OrderCount       int       `json:"order_count"`
	TotalAmount      float64   `json:"total_amount"`
	CouponOrderCount int       `json:"coupon_order_count"`
}

type WeeklySalesData struct {
	WeekNumber       int       `json:"week_number"`
	StartDate        time.Time `json:"start_date"`
	EndDate          time.Time `json:"end_date"`
	OrderCount       int       `json:"order_count"`
	TotalAmount      float64   `json:"total_amount"`
	CouponOrderCount int       `json:"coupon_order_count"`
}

type MonthlySalesData struct {
	Year             int        `json:"year"`
	Month            time.Month `json:"month"`
	OrderCount       int        `json:"order_count"`
	TotalAmount      float64    `json:"total_amount"`
	CouponOrderCount int        `json:"coupon_order_count"`
}

type YearlySalesData struct {
	Year             int     `json:"year"`
	OrderCount       int     `json:"order_count"`
	TotalAmount      float64 `json:"total_amount"`
	CouponOrderCount int     `json:"coupon_order_count"`
}

type CustomSalesData struct {
	StartDate        time.Time `json:"start_date"`
	EndDate          time.Time `json:"end_date"`
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
