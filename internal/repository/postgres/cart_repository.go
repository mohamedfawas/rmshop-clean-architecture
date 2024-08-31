package postgres

import (
	"context"
	"database/sql"
	"log"
	"time"

	"github.com/mohamedfawas/rmshop-clean-architecture/internal/domain"
	"github.com/mohamedfawas/rmshop-clean-architecture/pkg/utils"
)

type cartRepository struct {
	db *sql.DB
}

func NewCartRepository(db *sql.DB) *cartRepository {
	return &cartRepository{db: db}
}

func (r *cartRepository) AddCartItem(ctx context.Context, item *domain.CartItem) error {
	query := `
		INSERT INTO cart_items (user_id, product_id, quantity, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`
	now := time.Now().UTC()
	err := r.db.QueryRowContext(ctx, query, item.UserID, item.ProductID, item.Quantity, now, now).Scan(&item.ID)
	if err != nil {
		log.Printf("error while adding item to cart : %v", err)
	}
	return err
}

func (r *cartRepository) GetCartItemByProductID(ctx context.Context, userID, productID int64) (*domain.CartItem, error) {
	query := `
		SELECT id, user_id, product_id, quantity, created_at, updated_at
		FROM cart_items
		WHERE user_id = $1 AND product_id = $2
	`
	var item domain.CartItem
	err := r.db.QueryRowContext(ctx, query, userID, productID).Scan(
		&item.ID, &item.UserID, &item.ProductID, &item.Quantity, &item.CreatedAt, &item.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, utils.ErrCartItemNotFound
	}
	if err != nil {
		log.Printf("eror while retrieving cart item details using id : %v", err)
		return nil, err
	}
	return &item, nil
}

func (r *cartRepository) UpdateCartItem(ctx context.Context, item *domain.CartItem) error {
	query := `
		UPDATE cart_items
		SET quantity = $1, updated_at = $2
		WHERE id = $3
	`
	_, err := r.db.ExecContext(ctx, query, item.Quantity, time.Now(), item.ID)
	if err != nil {
		log.Printf("error while updating cart item details : %v", err)
	}
	return err
}
