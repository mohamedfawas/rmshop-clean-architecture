package postgres

import (
	"context"
	"database/sql"
	"log"

	"github.com/mohamedfawas/rmshop-clean-architecture/internal/domain"
)

type orderRepository struct {
	db *sql.DB
}

func NewOrderRepository(db *sql.DB) *orderRepository {
	return &orderRepository{db: db}
}

func (r *orderRepository) CreateOrder(ctx context.Context, tx *sql.Tx, order *domain.Order) error {
	query := `
		INSERT INTO orders (user_id, total_amount, payment_method, payment_status, delivery_status, address_id)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at
	`
	err := tx.QueryRowContext(ctx, query,
		order.UserID, order.TotalAmount, order.PaymentMethod, order.PaymentStatus, order.DeliveryStatus, order.AddressID,
	).Scan(&order.ID, &order.CreatedAt)
	if err != nil {
		log.Printf("error while adding the order entry : %v", err)
	}
	return err
}

func (r *orderRepository) AddOrderItem(ctx context.Context, tx *sql.Tx, item *domain.OrderItem) error {
	query := `
		INSERT INTO order_items (order_id, product_id, quantity, price)
		VALUES ($1, $2, $3, $4)
	`
	_, err := tx.ExecContext(ctx, query, item.OrderID, item.ProductID, item.Quantity, item.Price)
	if err != nil {
		log.Printf("error while adding order item entry : %v", err)
	}
	return err
}
