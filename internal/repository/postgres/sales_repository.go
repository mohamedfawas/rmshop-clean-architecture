package postgres

import (
	"context"
	"database/sql"
	"time"

	"github.com/mohamedfawas/rmshop-clean-architecture/internal/domain"
)

type salesRepository struct {
	db *sql.DB
}

func NewSalesRepository(db *sql.DB) *salesRepository {
	return &salesRepository{db: db}
}

func (r *salesRepository) GetSalesData(ctx context.Context, startDate, endDate time.Time, couponApplied bool) ([]domain.DailySalesData, error) {
	query := `
        SELECT 
            DATE(o.created_at) as date,
            COUNT(DISTINCT o.id) as order_count,
            SUM(o.final_amount) as total_amount,
            COUNT(DISTINCT CASE WHEN o.coupon_applied THEN o.id ELSE NULL END) as coupon_order_count
        FROM 
            orders o
        WHERE 
            o.created_at BETWEEN $1 AND $2
            AND ($3 = false OR o.coupon_applied = true)
        GROUP BY 
            DATE(o.created_at)
        ORDER BY 
            date
    `

	rows, err := r.db.QueryContext(ctx, query, startDate, endDate, couponApplied)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var dailyData []domain.DailySalesData
	for rows.Next() {
		var data domain.DailySalesData
		err := rows.Scan(&data.Date, &data.OrderCount, &data.TotalAmount, &data.CouponOrderCount)
		if err != nil {
			return nil, err
		}
		dailyData = append(dailyData, data)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return dailyData, nil
}

func (r *salesRepository) GetTopSellingProducts(ctx context.Context, startDate, endDate time.Time, limit int) ([]domain.TopSellingProduct, error) {
	query := `
        SELECT 
            p.id,
            p.name,
            SUM(oi.quantity) as total_quantity,
            SUM(oi.quantity * oi.price) as total_revenue
        FROM 
            order_items oi
        JOIN 
            orders o ON oi.order_id = o.id
        JOIN 
            products p ON oi.product_id = p.id
        WHERE 
            o.created_at BETWEEN $1 AND $2
        GROUP BY 
            p.id, p.name
        ORDER BY 
            total_quantity DESC
        LIMIT $3
    `

	rows, err := r.db.QueryContext(ctx, query, startDate, endDate, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []domain.TopSellingProduct
	for rows.Next() {
		var product domain.TopSellingProduct
		err := rows.Scan(&product.ID, &product.Name, &product.Quantity, &product.Revenue)
		if err != nil {
			return nil, err
		}
		products = append(products, product)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return products, nil
}
