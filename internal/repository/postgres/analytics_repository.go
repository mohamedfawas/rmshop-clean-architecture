package postgres

import (
	"context"
	"database/sql"
	"strconv"
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

func (r *analyticsRepository) GetTopCategories(ctx context.Context, params domain.TopCategoriesParams) ([]domain.TopCategory, error) {
	query := `
        SELECT c.id, c.name, SUM(oi.quantity * oi.price) as total_sales, COUNT(DISTINCT o.id) as order_count
        FROM categories c
        JOIN sub_categories sc ON c.id = sc.parent_category_id
        JOIN products p ON sc.id = p.sub_category_id
        JOIN order_items oi ON p.id = oi.product_id
        JOIN orders o ON oi.order_id = o.id
        WHERE o.created_at BETWEEN $1 AND $2
        GROUP BY c.id, c.name
        ORDER BY total_sales ` + params.SortOrder + `
        LIMIT $3
    `

	rows, err := r.db.QueryContext(ctx, query, params.StartDate, params.EndDate, params.Limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []domain.TopCategory
	for rows.Next() {
		var cat domain.TopCategory
		if err := rows.Scan(&cat.ID, &cat.Name, &cat.TotalSales, &cat.OrderCount); err != nil {
			return nil, err
		}
		categories = append(categories, cat)
	}

	return categories, nil
}

func (r *analyticsRepository) GetTopSubcategories(ctx context.Context, params domain.SubcategoryAnalyticsParams) ([]domain.TopSubcategory, error) {
	query := `
        SELECT 
            sc.id, 
            sc.name, 
            c.id AS category_id,
            c.name AS category_name,
            COUNT(oi.id) AS total_orders,
            SUM(oi.quantity) AS total_quantity,
            SUM(oi.quantity * oi.price) AS total_revenue
        FROM 
            sub_categories sc
        JOIN 
            categories c ON sc.parent_category_id = c.id
        JOIN 
            products p ON p.sub_category_id = sc.id
        JOIN 
            order_items oi ON oi.product_id = p.id
        JOIN 
            orders o ON o.id = oi.order_id
        WHERE 
            o.created_at BETWEEN $1 AND $2
    `

	args := []interface{}{params.StartDate, params.EndDate}
	argCount := 2

	if params.CategoryID != 0 {
		query += " AND c.id = $" + strconv.Itoa(argCount+1)
		args = append(args, params.CategoryID)
		argCount++
	}

	query += `
        GROUP BY 
            sc.id, sc.name, c.id, c.name
    `

	if params.SortBy == "revenue" {
		query += " ORDER BY total_revenue DESC"
	} else {
		query += " ORDER BY total_quantity DESC"
	}

	query += " LIMIT $" + strconv.Itoa(argCount+1) + " OFFSET $" + strconv.Itoa(argCount+2)
	args = append(args, params.Limit, (params.Page-1)*params.Limit)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var subcategories []domain.TopSubcategory
	for rows.Next() {
		var sc domain.TopSubcategory
		err := rows.Scan(
			&sc.ID, &sc.Name, &sc.CategoryID, &sc.CategoryName,
			&sc.TotalOrders, &sc.TotalQuantity, &sc.TotalRevenue,
		)
		if err != nil {
			return nil, err
		}
		subcategories = append(subcategories, sc)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return subcategories, nil
}
