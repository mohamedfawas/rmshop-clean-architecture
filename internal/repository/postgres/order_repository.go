package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/mohamedfawas/rmshop-clean-architecture/internal/domain"
	"github.com/mohamedfawas/rmshop-clean-architecture/pkg/utils"
)

type orderRepository struct {
	db *sql.DB
}

func NewOrderRepository(db *sql.DB) *orderRepository {
	return &orderRepository{db: db}
}

func (r *orderRepository) CreateOrder(ctx context.Context, tx *sql.Tx, order *domain.Order) (int64, error) {
	query := `
        INSERT INTO orders (user_id, total_amount, delivery_status, shipping_address_id, order_status, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7)
        RETURNING id, created_at
    `
	var orderID int64
	err := tx.QueryRowContext(ctx, query,
		order.UserID, order.TotalAmount, order.DeliveryStatus, order.ShippingAddressID, order.OrderStatus, time.Now().UTC(), time.Now().UTC(),
	).Scan(&orderID, &order.CreatedAt)
	if err != nil {
		log.Printf("error while adding the order entry : %v", err)
		return 0, err
	}
	order.ID = orderID
	return orderID, nil
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
        SELECT id, user_id, total_amount, delivery_status, 
               order_status, refund_status, shipping_address_id, created_at, updated_at, delivered_at
        FROM orders
        WHERE id = $1
    `
	var order domain.Order
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&order.ID,
		&order.UserID,
		&order.TotalAmount,
		&order.DeliveryStatus,
		&order.OrderStatus,
		&order.RefundStatus,
		&order.ShippingAddressID,
		&order.CreatedAt,
		&order.UpdatedAt,
		&order.DeliveredAt,
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
		SELECT o.id, o.total_amount, o.delivery_status, o.address_id, o.created_at
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
		err := rows.Scan(&o.ID, &o.TotalAmount, &o.DeliveryStatus, &o.ShippingAddressID, &o.CreatedAt)
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
	result, err := r.db.ExecContext(ctx, query, status, orderID)
	if err != nil {
		log.Printf("Error updating order status: %v", err)
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("Error getting rows affected: %v", err)
		return err
	}

	if rowsAffected == 0 {
		return utils.ErrOrderNotFound
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

func (r *orderRepository) GetOrders(ctx context.Context, params domain.OrderQueryParams) ([]*domain.Order, int64, error) {
	// Start building the base query
	query := `SELECT `
	if len(params.Fields) > 0 {
		// If specific fields are requested, use them
		query += strings.Join(params.Fields, ", ")
	} else {
		// Otherwise, select all fields
		query += `o.id, o.user_id, o.total_amount, o.delivery_status, 
                  o.order_status, o.refund_status, o.address_id, o.created_at, o.updated_at`
	}
	query += ` FROM orders o WHERE 1=1` // o is an alias, 1=1 : used in dynamic querying where it is always true
	// user to simplify the process of adding additional WHERE conditions dynamically

	// Initialize slices to hold query arguments and conditions
	var args []interface{}
	var conditions []string

	// Add condition for order status if provided
	if params.Status != "" {
		conditions = append(conditions, "o.order_status = $"+strconv.Itoa(len(args)+1))
		args = append(args, params.Status)
	}

	// Add condition for customer ID if provided
	if params.CustomerID != 0 {
		conditions = append(conditions, "o.user_id = $"+strconv.Itoa(len(args)+1))
		args = append(args, params.CustomerID)
	}

	// Add condition for start date if provided
	if params.StartDate != nil {
		conditions = append(conditions, "o.created_at >= $"+strconv.Itoa(len(args)+1))
		args = append(args, params.StartDate)
	}

	// Add condition for end date if provided
	if params.EndDate != nil {
		conditions = append(conditions, "o.created_at <= $"+strconv.Itoa(len(args)+1))
		args = append(args, params.EndDate)
	}

	// Combine all conditions with AND
	if len(conditions) > 0 {
		query += " AND " + strings.Join(conditions, " AND ")
	}

	// Create a count query to get total number of matching orders
	countQuery := "SELECT COUNT(*) FROM (" + query + ") AS count_query"
	var total int64
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		log.Printf("error while get count : %v", err)
		return nil, 0, err
	}

	// Apply sorting if sort field is provided
	if params.SortBy != "" {
		query += " ORDER BY o." + params.SortBy
		if params.SortOrder != "" {
			query += " " + params.SortOrder
		}
	}

	// Apply pagination
	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", len(args)+1, len(args)+2)
	args = append(args, params.Limit, (params.Page-1)*params.Limit)

	// Execute the final query
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		log.Printf("erorr applying pagination : %v", err)
		return nil, 0, err
	}
	defer rows.Close()

	// Iterate through the result set and create Order objects
	var orders []*domain.Order
	for rows.Next() {
		var o domain.Order
		err := rows.Scan(&o.ID, &o.UserID, &o.TotalAmount,
			&o.DeliveryStatus, &o.OrderStatus, &o.RefundStatus, &o.ShippingAddressID, &o.CreatedAt, &o.UpdatedAt)
		if err != nil {
			log.Printf("db error : %v", err)
			return nil, 0, err
		}
		orders = append(orders, &o)
	}

	// Return the list of orders, total count, and nil error if successful
	return orders, total, nil
}

func (r *orderRepository) UpdateOrderPaymentStatus(ctx context.Context, orderID int64, status string, paymentID string) error {
	query := `
        UPDATE orders
        SET payment_status = $1, razorpay_payment_id = $2, updated_at = NOW()
        WHERE id = $3
    `
	_, err := r.db.ExecContext(ctx, query, status, paymentID, orderID)
	if err != nil {
		return fmt.Errorf("error updating order payment status: %w", err)
	}
	return nil
}

func (r *orderRepository) GetPaymentByOrderID(ctx context.Context, orderID int64) (*domain.Payment, error) {
	query := `
        SELECT id, order_id, amount, payment_method, payment_status, created_at, updated_at, 
               razorpay_order_id, razorpay_payment_id, razorpay_signature
        FROM payments
        WHERE order_id = $1
    `
	var payment domain.Payment
	err := r.db.QueryRowContext(ctx, query, orderID).Scan(
		&payment.ID, &payment.OrderID, &payment.Amount, &payment.PaymentMethod, &payment.Status,
		&payment.CreatedAt, &payment.UpdatedAt, &payment.RazorpayOrderID, &payment.RazorpayPaymentID,
		&payment.RazorpaySignature,
	)
	if err == sql.ErrNoRows {
		return nil, utils.ErrPaymentNotFound
	}
	if err != nil {
		return nil, err
	}
	return &payment, nil
}

func (r *orderRepository) UpdatePayment(ctx context.Context, payment *domain.Payment) error {
	query := `
        UPDATE payments
        SET payment_status = $1, razorpay_payment_id = $2, razorpay_signature = $3, updated_at = NOW()
        WHERE id = $4
    `
	_, err := r.db.ExecContext(ctx, query,
		payment.Status, payment.RazorpayPaymentID, payment.RazorpaySignature, payment.ID,
	)
	return err
}

func (r *orderRepository) GetPaymentByRazorpayOrderID(ctx context.Context, razorpayOrderID string) (*domain.Payment, error) {
	query := `
        SELECT id, order_id, amount, payment_method, payment_status, created_at, updated_at, 
               razorpay_order_id, razorpay_payment_id, razorpay_signature
        FROM payments
        WHERE razorpay_order_id = $1
    `
	var payment domain.Payment
	err := r.db.QueryRowContext(ctx, query, razorpayOrderID).Scan(
		&payment.ID, &payment.OrderID, &payment.Amount, &payment.PaymentMethod, &payment.Status,
		&payment.CreatedAt, &payment.UpdatedAt, &payment.RazorpayOrderID, &payment.RazorpayPaymentID,
		&payment.RazorpaySignature,
	)
	if err == sql.ErrNoRows {
		return nil, utils.ErrPaymentNotFound
	}
	if err != nil {
		return nil, err
	}
	return &payment, nil
}

func (r *orderRepository) CreatePayment(ctx context.Context, tx *sql.Tx, payment *domain.Payment) error {
	query := `
        INSERT INTO payments (order_id, amount, payment_method, payment_status, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6)
        RETURNING id
    `
	err := tx.QueryRowContext(ctx, query,
		payment.OrderID, payment.Amount, payment.PaymentMethod, payment.Status, time.Now().UTC(), time.Now().UTC(),
	).Scan(&payment.ID)
	if err != nil {
		return fmt.Errorf("failed to create payment: %w", err)
	}
	return nil
}

func (r *orderRepository) GetReturnRequestByOrderID(ctx context.Context, orderID int64) (*domain.ReturnRequest, error) {
	query := `
		SELECT id, order_id, reason, status, created_at, updated_at
		FROM return_requests
		WHERE order_id = $1
	`
	var returnRequest domain.ReturnRequest
	err := r.db.QueryRowContext(ctx, query, orderID).Scan(
		&returnRequest.ID,
		&returnRequest.OrderID,
		&returnRequest.Reason,
		&returnRequest.Status,
		&returnRequest.CreatedAt,
		&returnRequest.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, utils.ErrReturnRequestNotFound
	}
	if err != nil {
		return nil, err
	}
	return &returnRequest, nil
}

func (r *orderRepository) CreateReturnRequest(ctx context.Context, returnRequest *domain.ReturnRequest) error {
	query := `
		INSERT INTO return_requests (order_id, reason, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`
	err := r.db.QueryRowContext(ctx, query,
		returnRequest.OrderID,
		returnRequest.Reason,
		returnRequest.Status,
		returnRequest.CreatedAt,
		returnRequest.UpdatedAt,
	).Scan(&returnRequest.ID)
	return err
}

func (r *orderRepository) SetOrderDeliveredAt(ctx context.Context, orderID int64, deliveredAt *time.Time) error {
	query := `
        UPDATE orders
        SET delivered_at = $1, updated_at = NOW()
        WHERE id = $2
    `
	_, err := r.db.ExecContext(ctx, query, deliveredAt, orderID)
	return err
}

func (r *orderRepository) BeginTx(ctx context.Context) (*sql.Tx, error) {
	return r.db.BeginTx(ctx, nil)
}

func (r *orderRepository) CreateReturnRequestTx(ctx context.Context, tx *sql.Tx, returnRequest *domain.ReturnRequest) error {
	query := `
        INSERT INTO return_requests (order_id, reason, status, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5)
        RETURNING id
    `
	err := tx.QueryRowContext(ctx, query,
		returnRequest.OrderID,
		returnRequest.Reason,
		returnRequest.Status,
		returnRequest.CreatedAt,
		returnRequest.UpdatedAt,
	).Scan(&returnRequest.ID)
	return err
}

func (r *orderRepository) UpdateOrderStatusTx(ctx context.Context, tx *sql.Tx, orderID int64, status string) error {
	query := `
        UPDATE orders
        SET order_status = $1, updated_at = NOW()
        WHERE id = $2
    `
	_, err := tx.ExecContext(ctx, query, status, orderID)
	return err
}

func (r *orderRepository) UpdateRefundStatusTx(ctx context.Context, tx *sql.Tx, orderID int64, refundStatus sql.NullString) error {
	query := `
        UPDATE orders
        SET refund_status = $1, updated_at = NOW()
        WHERE id = $2
    `
	_, err := tx.ExecContext(ctx, query, refundStatus, orderID)
	return err
}

func (r *orderRepository) CreateRefundTx(ctx context.Context, tx *sql.Tx, refund *domain.Refund) error {
	query := `
        INSERT INTO refunds (order_id, amount, status, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5)
        RETURNING id
    `
	err := tx.QueryRowContext(ctx, query,
		refund.OrderID,
		refund.Amount,
		refund.Status,
		refund.CreatedAt,
		refund.UpdatedAt,
	).Scan(&refund.ID)
	return err
}
