package postgres

import (
	"context"
	"database/sql"
	"fmt"
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
        SELECT id, user_id, total_amount, payment_method, payment_status, delivery_status, 
               order_status, refund_status, address_id, created_at, updated_at
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
		&order.OrderStatus,
		&order.RefundStatus,
		&order.AddressID,
		&order.CreatedAt,
		&order.UpdatedAt,
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

func (r *orderRepository) GetUserOrders(ctx context.Context, userID int64, page, limit int, sortBy, order, status string) ([]*domain.Order, int64, error) {
	offset := (page - 1) * limit

	// Build the base query
	query := `
		SELECT o.id, o.total_amount, o.payment_method, o.payment_status, o.delivery_status, o.address_id, o.created_at
		FROM orders o
		WHERE o.user_id = $1
	`
	countQuery := `SELECT COUNT(*) FROM orders o WHERE o.user_id = $1`
	args := []interface{}{userID}

	// Add status filter if provided
	if status != "" {
		query += " AND o.delivery_status = $2"
		countQuery += " AND o.delivery_status = $2"
		args = append(args, status)
	}

	// Add sorting
	query += fmt.Sprintf(" ORDER BY o.%s %s", sortBy, order)

	// Add pagination
	query += " LIMIT $2 OFFSET $3"
	args = append(args, limit, offset)

	// Execute the count query
	var totalCount int64
	err := r.db.QueryRowContext(ctx, countQuery, args[:len(args)-2]...).Scan(&totalCount)
	if err != nil {
		log.Printf("Error counting orders: %v", err)
		return nil, 0, err
	}

	// Execute the main query
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		log.Printf("Error querying orders: %v", err)
		return nil, 0, err
	}
	defer rows.Close()

	var orders []*domain.Order
	for rows.Next() {
		var o domain.Order
		err := rows.Scan(&o.ID, &o.TotalAmount, &o.PaymentMethod, &o.PaymentStatus, &o.DeliveryStatus, &o.AddressID, &o.CreatedAt)
		if err != nil {
			log.Printf("Error scanning order row: %v", err)
			return nil, 0, err
		}
		orders = append(orders, &o)
	}

	if err = rows.Err(); err != nil {
		log.Printf("Error iterating order rows: %v", err)
		return nil, 0, err
	}

	return orders, totalCount, nil
}

func (r *orderRepository) UpdateOrderStatus(ctx context.Context, orderID int64, status string) error {
	query := `
        UPDATE orders
        SET order_status = $1, updated_at = NOW()
        WHERE id = $2
    `
	_, err := r.db.ExecContext(ctx, query, status, orderID)
	if err != nil {
		log.Printf("Error updating order status: %v", err)
		return err
	}
	return nil
}

func (r *orderRepository) UpdateRefundStatus(ctx context.Context, orderID int64, refundStatus sql.NullString) error {
	query := `
		UPDATE orders
		SET refund_status = $1, updated_at = NOW()
		WHERE id = $2
	`
	_, err := r.db.ExecContext(ctx, query, refundStatus, orderID)
	if err != nil {
		log.Printf("Error updating refund status: %v", err)
		return err
	}
	return nil
}
