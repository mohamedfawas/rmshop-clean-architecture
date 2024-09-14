package postgres

import (
	"context"
	"database/sql"

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
