package postgres

import (
	"context"
	"database/sql"
	"log"

	"github.com/mohamedfawas/rmshop-clean-architecture/internal/domain"
	"github.com/mohamedfawas/rmshop-clean-architecture/pkg/utils"
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

func (r *orderRepository) GetByID(ctx context.Context, id int64) (*domain.Order, error) {
	query := `
        SELECT id, user_id, total_amount, payment_method, payment_status, delivery_status, address_id, created_at
        FROM orders
        WHERE id = $1
    `
	var order domain.Order
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&order.ID,
		&order.UserID,
		&order.TotalAmount,
		&order.PaymentMethod,
		&order.PaymentStatus,
		&order.DeliveryStatus,
		&order.AddressID,
		&order.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, utils.ErrOrderNotFound
		}
		log.Printf("Error retrieving order: %v", err)
		return nil, err
	}

	// Fetch order items
	itemsQuery := `
        SELECT id, product_id, quantity, price
        FROM order_items
        WHERE order_id = $1
    `
	rows, err := r.db.QueryContext(ctx, itemsQuery, id)
	if err != nil {
		log.Printf("Error retrieving order items: %v", err)
		return nil, err
	}
	defer rows.Close()

	var items []domain.OrderItem
	for rows.Next() {
		var item domain.OrderItem
		err := rows.Scan(&item.ID, &item.ProductID, &item.Quantity, &item.Price)
		if err != nil {
			log.Printf("Error scanning order item: %v", err)
			return nil, err
		}
		items = append(items, item)
	}
	if err = rows.Err(); err != nil {
		log.Printf("Error iterating order items: %v", err)
		return nil, err
	}

	order.Items = items
	return &order, nil
}
