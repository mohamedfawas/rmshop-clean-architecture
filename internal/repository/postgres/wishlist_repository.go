package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/mohamedfawas/rmshop-clean-architecture/internal/domain"
)

type wishlistRepository struct {
	db *sql.DB
}

func NewWishlistRepository(db *sql.DB) *wishlistRepository {
	return &wishlistRepository{db: db}
}

func (r *wishlistRepository) AddItem(ctx context.Context, item *domain.WishlistItem) error {
	query := `
		INSERT INTO wishlist_items (user_id, product_id, is_available, price)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at
	`
	return r.db.QueryRowContext(ctx, query, item.UserID, item.ProductID, item.IsAvailable, item.Price).Scan(&item.ID, &item.CreatedAt)
}

func (r *wishlistRepository) ItemExists(ctx context.Context, userID, productID int64) (bool, error) {
	query := `
		SELECT EXISTS(SELECT 1 FROM wishlist_items WHERE user_id = $1 AND product_id = $2)
	`
	var exists bool
	err := r.db.QueryRowContext(ctx, query, userID, productID).Scan(&exists)
	return exists, err
}

func (r *wishlistRepository) GetWishlistItemCount(ctx context.Context, userID int64) (int, error) {
	query := `
		SELECT COUNT(*) FROM wishlist_items WHERE user_id = $1
	`
	var count int
	err := r.db.QueryRowContext(ctx, query, userID).Scan(&count)
	return count, err
}

func (r *wishlistRepository) RemoveItem(ctx context.Context, userID, productID int64) error {
	query := `
        DELETE FROM wishlist_items WHERE user_id = $1 AND product_id = $2
    `
	_, err := r.db.ExecContext(ctx, query, userID, productID)
	return err
}

func (r *wishlistRepository) GetUserWishlistItems(ctx context.Context, userID int64, page, limit int, sortBy, order string) ([]*domain.WishlistItem, int64, error) {
	offset := (page - 1) * limit

	// Count total items
	var totalCount int64
	countErr := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM wishlist_items WHERE user_id = $1", userID).Scan(&totalCount)
	if countErr != nil {
		return nil, 0, countErr
	}

	// Fetch items
	query := fmt.Sprintf(`
        SELECT wi.id, wi.user_id, wi.product_id, wi.is_available, wi.price, wi.created_at, p.name
        FROM wishlist_items wi
        JOIN products p ON wi.product_id = p.id
        WHERE wi.user_id = $1
        ORDER BY %s %s
        LIMIT $2 OFFSET $3
    `, sortBy, order)

	rows, err := r.db.QueryContext(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var items []*domain.WishlistItem
	for rows.Next() {
		var item domain.WishlistItem
		err := rows.Scan(&item.ID, &item.UserID, &item.ProductID, &item.IsAvailable, &item.Price, &item.CreatedAt, &item.ProductName)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, &item)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, err
	}

	return items, totalCount, nil
}
