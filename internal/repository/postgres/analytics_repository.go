package postgres

import (
	"context"
	"database/sql"
	"time"

	"github.com/mohamedfawas/rmshop-clean-architecture/internal/domain"
)

type analyticsRepository struct {
	db *sql.DB
}

func NewAnalyticsRepository(db *sql.DB) *analyticsRepository {
	return &analyticsRepository{db: db}
}

func (r *analyticsRepository) GetTopProducts(ctx context.Context, startDate, endDate time.Time, limit int, sortBy string) ([]domain.TopProduct, error) {
	query := `
		SELECT p.id, p.name, SUM(oi.quantity) as total_quantity, SUM(oi.quantity * oi.price) as total_revenue
		FROM order_items oi
		JOIN orders o ON oi.order_id = o.id
		JOIN products p ON oi.product_id = p.id
		WHERE o.created_at BETWEEN $1 AND $2
		GROUP BY p.id, p.name
		ORDER BY 
	`

	if sortBy == "revenue" {
		query += "total_revenue DESC"
	} else {
		query += "total_quantity DESC"
	}

	query += " LIMIT $3"

	rows, err := r.db.QueryContext(ctx, query, startDate, endDate, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var topProducts []domain.TopProduct
	for rows.Next() {
		var p domain.TopProduct
		if err := rows.Scan(&p.ID, &p.Name, &p.TotalQuantity, &p.TotalRevenue); err != nil {
			return nil, err
		}
		topProducts = append(topProducts, p)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return topProducts, nil
}
