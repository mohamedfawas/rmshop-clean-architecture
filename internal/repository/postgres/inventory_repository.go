package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/mohamedfawas/rmshop-clean-architecture/internal/domain"
)

type inventoryRepository struct {
	db *sql.DB
}

func NewInventoryRepository(db *sql.DB) *inventoryRepository {
	return &inventoryRepository{db: db}
}

func (r *inventoryRepository) GetInventory(ctx context.Context, params domain.InventoryQueryParams) ([]*domain.InventoryItem, int64, error) {
	query := `
        SELECT p.id, p.name, p.price, p.stock_quantity, c.id, c.name
        FROM products p
        JOIN sub_categories sc ON p.sub_category_id = sc.id
        JOIN categories c ON sc.parent_category_id = c.id
        WHERE 1=1
    `

	var conditions []string
	var args []interface{}
	argCount := 1

	if params.ProductID != 0 {
		conditions = append(conditions, fmt.Sprintf("p.id = $%d", argCount))
		args = append(args, params.ProductID)
		argCount++
	}

	if params.ProductName != "" {
		conditions = append(conditions, fmt.Sprintf("LOWER(p.name) LIKE LOWER($%d)", argCount))
		args = append(args, "%"+params.ProductName+"%")
		argCount++
	}

	if params.CategoryID != 0 {
		conditions = append(conditions, fmt.Sprintf("c.id = $%d", argCount))
		args = append(args, params.CategoryID)
		argCount++
	}

	if params.CategoryName != "" {
		conditions = append(conditions, fmt.Sprintf("LOWER(c.name) LIKE LOWER($%d)", argCount))
		args = append(args, "%"+params.CategoryName+"%")
		argCount++
	}

	if params.StockLessThan != nil {
		conditions = append(conditions, fmt.Sprintf("p.stock_quantity < $%d", argCount))
		args = append(args, *params.StockLessThan)
		argCount++
	}

	if params.StockMoreThan != nil {
		conditions = append(conditions, fmt.Sprintf("p.stock_quantity > $%d", argCount))
		args = append(args, *params.StockMoreThan)
		argCount++
	}

	if len(conditions) > 0 {
		query += " AND " + strings.Join(conditions, " AND ")
	}

	// Count total before applying pagination
	countQuery := "SELECT COUNT(*) FROM (" + query + ") AS count_query"
	var total int64
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Apply sorting
	if params.SortBy != "" {
		query += fmt.Sprintf(" ORDER BY %s", params.SortBy)
		if params.SortOrder != "" {
			query += " " + params.SortOrder
		}
	}

	// Apply pagination
	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argCount, argCount+1)
	args = append(args, params.Limit, (params.Page-1)*params.Limit)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var items []*domain.InventoryItem
	for rows.Next() {
		var item domain.InventoryItem
		err := rows.Scan(&item.ProductID, &item.ProductName, &item.Price, &item.StockQuantity, &item.CategoryID, &item.CategoryName)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, &item)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, err
	}

	return items, total, nil
}
