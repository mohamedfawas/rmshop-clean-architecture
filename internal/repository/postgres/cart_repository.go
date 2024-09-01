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

func (r *cartRepository) GetCartByUserID(ctx context.Context, userID int64) ([]*domain.CartItemWithProduct, error) {
	query := `
		SELECT ci.id, ci.user_id, ci.product_id, ci.quantity, ci.created_at, ci.updated_at,
			   p.name, p.price
		FROM cart_items ci
		JOIN products p ON ci.product_id = p.id
		WHERE ci.user_id = $1 AND p.deleted_at IS NULL
		ORDER BY ci.created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		log.Printf("error while retrieving cart using userID : %v", err)
		return nil, err
	}
	defer rows.Close()

	var cartItems []*domain.CartItemWithProduct
	for rows.Next() {
		var ci domain.CartItemWithProduct
		err := rows.Scan(
			&ci.ID, &ci.UserID, &ci.ProductID, &ci.Quantity, &ci.CreatedAt, &ci.UpdatedAt,
			&ci.ProductName, &ci.ProductPrice,
		)
		if err != nil {
			log.Printf("db error : %v", err)
			return nil, err
		}
		cartItems = append(cartItems, &ci)
	}

	if err = rows.Err(); err != nil {
		log.Printf("db error : %v", err)
		return nil, err
	}

	return cartItems, nil
}

func (r *cartRepository) UpdateCartItemQuantity(ctx context.Context, userID, itemID int64, quantity int) error {
	query := `UPDATE cart_items
				SET quantity=$1, updated_at = NOW()
				WHERE id=$2 AND user_id=$3`
	result, err := r.db.ExecContext(ctx, query, quantity, itemID, userID)
	if err != nil {
		log.Printf("error while updating the cart_items : %v", err)
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("erorr while checking the rows affected : %v", err)
		return err
	}

	if rowsAffected == 0 {
		return utils.ErrCartItemNotFound
	}

	return nil
}

func (r *cartRepository) DeleteCartItem(ctx context.Context, itemID int64) error {
	query := `
        DELETE FROM cart_items
        WHERE id = $1
    `
	result, err := r.db.ExecContext(ctx, query, itemID)
	if err != nil {
		log.Printf("error while deleting the cart item : %v", err)
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("error while checking the rows affected : %v", err)
		return err
	}

	if rowsAffected == 0 {
		return utils.ErrCartItemNotFound
	}

	return nil
}

func (r *cartRepository) GetCartItemByID(ctx context.Context, itemID int64) (*domain.CartItem, error) {
	query := `
        SELECT id, user_id, product_id, quantity, created_at, updated_at
        FROM cart_items
        WHERE id = $1
    `
	var item domain.CartItem
	err := r.db.QueryRowContext(ctx, query, itemID).Scan(
		&item.ID, &item.UserID, &item.ProductID, &item.Quantity, &item.CreatedAt, &item.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, utils.ErrCartItemNotFound
	}
	if err != nil {
		log.Printf("error while retrieiving cart item details : %v", err)
		return nil, err
	}
	return &item, nil
}
