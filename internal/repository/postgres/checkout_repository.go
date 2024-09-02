package postgres

import (
	"context"
	"database/sql"
	"log"

	"github.com/mohamedfawas/rmshop-clean-architecture/internal/domain"
)

type checkoutRepository struct {
	db *sql.DB
}

func NewCheckoutRepository(db *sql.DB) *checkoutRepository {
	return &checkoutRepository{db: db}
}

func (r *checkoutRepository) CreateCheckoutSession(ctx context.Context, userID int64) (*domain.CheckoutSession, error) {
	query := `
        INSERT INTO checkout_sessions (user_id, status, created_at, updated_at)
        VALUES ($1, 'pending', NOW(), NOW())
        RETURNING id, created_at
    `
	var session domain.CheckoutSession
	err := r.db.QueryRowContext(ctx, query, userID).Scan(&session.ID, &session.CreatedAt)
	if err != nil {
		log.Printf("error while creating checkout session entry : %v", err)
		return nil, err
	}
	session.UserID = userID
	session.Status = "pending"
	return &session, nil
}

func (r *checkoutRepository) AddCheckoutItems(ctx context.Context, sessionID int64, items []*domain.CheckoutItem) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `
        INSERT INTO checkout_items (session_id, product_id, quantity, price, subtotal)
        VALUES ($1, $2, $3, $4, $5)
    `
	for _, item := range items {
		_, err := tx.ExecContext(ctx, query, sessionID, item.ProductID, item.Quantity, item.Price, item.Subtotal)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *checkoutRepository) GetCartItems(ctx context.Context, userID int64) ([]*domain.CartItemWithProduct, error) {
	query := `
        SELECT ci.id, ci.product_id, ci.quantity, p.name, p.price
        FROM cart_items ci
        JOIN products p ON ci.product_id = p.id
        WHERE ci.user_id = $1
    `
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*domain.CartItemWithProduct
	for rows.Next() {
		var item domain.CartItemWithProduct
		err := rows.Scan(&item.ID, &item.ProductID, &item.Quantity, &item.ProductName, &item.ProductPrice)
		if err != nil {
			return nil, err
		}
		items = append(items, &item)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}
