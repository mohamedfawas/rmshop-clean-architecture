package postgres

import (
	"context"
	"database/sql"
	"log"
	"time"

	"github.com/mohamedfawas/rmshop-clean-architecture/internal/domain"
)

type salesRepository struct {
	db *sql.DB
}

func NewSalesRepository(db *sql.DB) *salesRepository {
	return &salesRepository{db: db}
}

func (r *salesRepository) GetDailySalesData(ctx context.Context, date time.Time) ([]domain.DailySales, error) {
	query := `
		SELECT 
			DATE(o.created_at) as date,
			COUNT(DISTINCT o.id) as order_count,
			SUM(o.final_amount) as total_amount,
            -- Counts the distinct number of orders where a coupon was applied (if 'coupon_applied' is true)
			COUNT(DISTINCT CASE WHEN o.coupon_applied THEN o.id ELSE NULL END) as coupon_order_count
		FROM 
            -- Refers to the 'orders' table with an alias 'o'
			orders o
		WHERE 
            -- Filters the results to only include orders created on a specific date (provided via parameter)
			DATE(o.created_at) = $1
		GROUP BY 
            -- Groups the results by the 'created_at' date (since we're interested in aggregating by date)
			DATE(o.created_at)
	`

	var salesData domain.DailySales
	err := r.db.QueryRowContext(ctx, query, date).Scan(
		&salesData.Date,
		&salesData.OrderCount,
		&salesData.TotalAmount,
		&salesData.CouponOrderCount,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		log.Printf("error while retrieving Get Daily Sales data from db  : %v", err)
		return nil, err
	}

	return []domain.DailySales{salesData}, nil
}

func (r *salesRepository) GetWeeklySalesData(ctx context.Context, startDate time.Time) ([]domain.DailySales, error) {
	endDate := startDate.AddDate(0, 0, 6) // 7 days including start date
	query := `
        SELECT 
            DATE(o.created_at) as date,
            COUNT(DISTINCT o.id) as order_count,
            SUM(o.final_amount) as total_amount,
            COUNT(DISTINCT CASE WHEN o.coupon_applied THEN o.id ELSE NULL END) as coupon_order_count
        FROM 
            orders o
        WHERE 
            DATE(o.created_at) BETWEEN $1 AND $2
        GROUP BY 
            DATE(o.created_at)
        ORDER BY
            DATE(o.created_at)
    `

	rows, err := r.db.QueryContext(ctx, query, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var salesData []domain.DailySales
	for rows.Next() {
		var sale domain.DailySales
		err := rows.Scan(
			&sale.Date,
			&sale.OrderCount,
			&sale.TotalAmount,
			&sale.CouponOrderCount,
		)
		if err != nil {
			return nil, err
		}
		salesData = append(salesData, sale)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return salesData, nil
}

func (r *salesRepository) GetMonthlySalesData(ctx context.Context, year int, month time.Month) ([]domain.DailySales, error) {
	startDate := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)
	endDate := startDate.AddDate(0, 1, -1)

	query := `
        SELECT 
            DATE(o.created_at) as date,
            COUNT(DISTINCT o.id) as order_count,
            SUM(o.final_amount) as total_amount,
            COUNT(DISTINCT CASE WHEN o.coupon_applied THEN o.id ELSE NULL END) as coupon_order_count
        FROM 
            orders o
        WHERE 
            DATE(o.created_at) BETWEEN $1 AND $2
        GROUP BY 
            DATE(o.created_at)
        ORDER BY
            DATE(o.created_at)
    `

	rows, err := r.db.QueryContext(ctx, query, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var salesData []domain.DailySales
	for rows.Next() {
		var sale domain.DailySales
		err := rows.Scan(
			&sale.Date,
			&sale.OrderCount,
			&sale.TotalAmount,
			&sale.CouponOrderCount,
		)
		if err != nil {
			return nil, err
		}
		salesData = append(salesData, sale)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return salesData, nil
}

func (r *salesRepository) GetCustomSalesData(ctx context.Context, startDate, endDate time.Time) ([]domain.DailySales, error) {
	query := `
        SELECT 
            DATE(o.created_at) as date,
            COUNT(DISTINCT o.id) as order_count,
            SUM(o.final_amount) as total_amount,
            COUNT(DISTINCT CASE WHEN o.coupon_applied THEN o.id ELSE NULL END) as coupon_order_count
        FROM 
            orders o
        WHERE 
            DATE(o.created_at) BETWEEN $1 AND $2
        GROUP BY 
            DATE(o.created_at)
        ORDER BY
            DATE(o.created_at)
    `

	rows, err := r.db.QueryContext(ctx, query, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var salesData []domain.DailySales
	for rows.Next() {
		var sale domain.DailySales
		err := rows.Scan(
			&sale.Date,
			&sale.OrderCount,
			&sale.TotalAmount,
			&sale.CouponOrderCount,
		)
		if err != nil {
			return nil, err
		}
		salesData = append(salesData, sale)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return salesData, nil
}
