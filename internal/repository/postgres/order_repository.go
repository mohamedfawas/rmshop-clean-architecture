package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"log"
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

/*
GetUserOrders:
  - Main query : retrieve order details from orders table
    -sub query : count number of orders made by the user, used for recording total count in handler side
*/
func (r *orderRepository) GetUserOrders(ctx context.Context, userID int64, page int) ([]*domain.Order, int64, error) {
	// select values after the offset value , if offset is 10, values after first 10 after chosen
	offset := (page - 1) * 10

	query := `
        SELECT id, user_id, total_amount, discount_amount, final_amount, shipping_address_id, 
               coupon_applied, has_return_request, created_at, updated_at, delivered_at, 
               order_status, delivery_status
        FROM orders
        WHERE user_id = $1
        ORDER BY created_at DESC
        LIMIT 10 OFFSET $2
    `

	// Count total number of orders made by the given user
	countQuery := `SELECT COUNT(*) FROM orders WHERE user_id = $1`

	// total count of user orders
	var totalCount int64
	// Get total count of user's orders
	err := r.db.QueryRowContext(ctx, countQuery, userID).Scan(&totalCount)
	if err != nil {
		log.Printf("error counting user's orders: %v", err)
		return nil, 0, err
	}

	// Execute main query : get the user's orders based on the calculated offset
	rows, err := r.db.QueryContext(ctx, query, userID, offset)
	if err != nil {
		log.Printf("error querying orders from orders table: %v", err)
		return nil, 0, err
	}
	defer rows.Close() // Ensure rows are closed after processing

	var orders []*domain.Order
	// Iterate through query results
	for rows.Next() {
		var o domain.Order
		err := rows.Scan(
			&o.ID,
			&o.UserID,
			&o.TotalAmount,
			&o.DiscountAmount,
			&o.FinalAmount,
			&o.ShippingAddressID,
			&o.CouponApplied,
			&o.HasReturnRequest,
			&o.CreatedAt,
			&o.UpdatedAt,
			&o.DeliveredAt,
			&o.OrderStatus,
			&o.DeliveryStatus,
		)
		if err != nil {
			log.Printf("error scanning order row: %v", err)
			return nil, 0, err
		}
		orders = append(orders, &o)
	}

	// error captured during iteration
	if err = rows.Err(); err != nil {
		log.Printf("error iterating order rows: %v", err)
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
	log.Printf("Executing update order status query for order ID %d with status %s", orderID, status)
	result, err := r.db.ExecContext(ctx, query, status, orderID)
	if err != nil {
		log.Printf("Error executing update order status query: %v", err)
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("Error getting rows affected: %v", err)
		return err
	}

	log.Printf("Rows affected by update order status: %d", rowsAffected)
	if rowsAffected == 0 {
		return fmt.Errorf("order not found or not updated: ID %d", orderID)
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

/*
GetOrders:
- Get orders based on on offset and limit
- Used to display all the orders for admin side
*/
func (r *orderRepository) GetOrders(ctx context.Context, limit, offset int) ([]*domain.Order, int64, error) {
	// Query to get total count of orders
	var totalOrders int64
	countErr := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM orders").Scan(&totalOrders)
	if countErr != nil {
		log.Printf("error while counting orders")
		return nil, 0, countErr
	}

	// Query to get paginated orders
	query := `
        SELECT id, user_id, total_amount, discount_amount, final_amount, shipping_address_id, 
               coupon_applied, has_return_request, created_at, updated_at, delivered_at, 
               order_status, delivery_status
        FROM orders
        ORDER BY created_at DESC
        LIMIT $1 OFFSET $2
    `
	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		log.Printf("error while fetching orders from orders table : %v", err)
		return nil, 0, err
	}
	defer rows.Close()

	var orders []*domain.Order
	for rows.Next() {
		var o domain.Order
		err := rows.Scan(
			&o.ID,
			&o.UserID,
			&o.TotalAmount,
			&o.DiscountAmount,
			&o.FinalAmount,
			&o.ShippingAddressID,
			&o.CouponApplied,
			&o.HasReturnRequest,
			&o.CreatedAt,
			&o.UpdatedAt,
			&o.DeliveredAt,
			&o.OrderStatus,
			&o.DeliveryStatus,
		)
		if err != nil {
			log.Printf("error while scanning rows : %v", err)
			return nil, 0, err
		}
		orders = append(orders, &o)
	}

	// capture the errors we didn't got before
	if err = rows.Err(); err != nil {
		log.Printf("error while iterating over the rows of orders : %v", err)
		return nil, 0, err
	}

	return orders, totalOrders, nil
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

func (r *orderRepository) UpdateOrderStatusTx(ctx context.Context, tx *sql.Tx, orderID int64, status string) error {
	query := `
        UPDATE orders
        SET order_status = $1, updated_at = NOW()
        WHERE id = $2
    `
	_, err := tx.ExecContext(ctx, query, status, orderID)
	if err != nil {
		log.Printf("error while updating order status : %v", err)
	}
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

/*
CreateOrder:
  - Create order entry in the "orders" table
  - user_id, total_amount, discount_amount, final_amount, delivery_status,
    shipping_address_id, order_status, coupon_applied
*/
func (r *orderRepository) CreateOrder(ctx context.Context, tx *sql.Tx, order *domain.Order) (int64, error) {
	query := `
        INSERT INTO orders (user_id, total_amount, discount_amount, final_amount, delivery_status, 
                            shipping_address_id, order_status, coupon_applied, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
        RETURNING id
    `
	var orderID int64
	err := tx.QueryRowContext(ctx, query,
		order.UserID,
		order.TotalAmount,
		order.DiscountAmount,
		order.FinalAmount,
		order.DeliveryStatus,
		order.ShippingAddressID,
		order.OrderStatus,
		order.CouponApplied,
		order.CreatedAt,
		order.UpdatedAt,
	).Scan(&orderID)
	if err != nil {
		log.Printf("error while adding the order entry in the orders: %v", err)
		return 0, err
	}
	return orderID, nil
}

/*
AddOrderItem:
- Add order item entry in order_items table
- order_id, product_id, quantity, price
*/
func (r *orderRepository) AddOrderItem(ctx context.Context, tx *sql.Tx, item *domain.OrderItem) error {
	query := `
        INSERT INTO order_items (order_id, product_id, quantity, price)
        VALUES ($1, $2, $3, $4)
    `
	_, err := tx.ExecContext(ctx, query, item.OrderID, item.ProductID, item.Quantity, item.Price)
	if err != nil {
		log.Printf("error while adding order item entry : %v", err)
		return err
	}
	return nil
}

func (r *orderRepository) UpdateOrderRazorpayID(ctx context.Context, orderID int64, razorpayOrderID string) error {
	query := `
        UPDATE payments
        SET razorpay_order_id = $1
        WHERE order_id = $2
    `
	_, err := r.db.ExecContext(ctx, query, razorpayOrderID, orderID)
	if err != nil {
		log.Printf("failed to update razorpay_order_id in payments table : %v", err)
	}
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
	var razorpayPaymentID, razorpaySignature sql.NullString
	err := r.db.QueryRowContext(ctx, query, razorpayOrderID).Scan(
		&payment.ID, &payment.OrderID, &payment.Amount, &payment.PaymentMethod, &payment.Status,
		&payment.CreatedAt, &payment.UpdatedAt, &payment.RazorpayOrderID,
		&razorpayPaymentID, &razorpaySignature,
	)
	if err == sql.ErrNoRows {
		log.Printf("No payment found for Razorpay order ID: %s", razorpayOrderID)
		return nil, utils.ErrPaymentNotFound
	}
	if err != nil {
		log.Printf("Error querying payment by Razorpay order ID %s: %v", razorpayOrderID, err)
		return nil, fmt.Errorf("database error: %w", err)
	}

	if razorpayPaymentID.Valid {
		payment.RazorpayPaymentID = razorpayPaymentID.String
	}
	if razorpaySignature.Valid {
		payment.RazorpaySignature = razorpaySignature.String
	}

	return &payment, nil
}

/*
GetByID:
- Fetch order details and order items details
-Calls associated methods 'GetOrderDetails' and 'GetOrderItems'
*/
func (r *orderRepository) GetByID(ctx context.Context, id int64) (*domain.Order, error) {
	// Fetch the main order details from orders table
	order, err := r.GetOrderDetails(ctx, id)
	if err != nil {
		log.Printf("failed to get order details from orders table : %v", err)
		return nil, err
	}

	// Fetch order items from order_items table
	items, err := r.GetOrderItems(ctx, id)
	if err != nil {
		log.Printf("failed to get order item details from order_items table : %v", err)
		return nil, err
	}

	// Combine order details and items
	order.Items = items

	return order, nil
}

/*
GetOrderDetails:
- Fetch order details from orders table
*/
func (r *orderRepository) GetOrderDetails(ctx context.Context, id int64) (*domain.Order, error) {
	query := `
        SELECT id, user_id, total_amount, discount_amount, final_amount, delivery_status, 
               order_status, has_return_request, shipping_address_id, coupon_applied,
               created_at, updated_at, delivered_at
        FROM orders
        WHERE id = $1
    `
	var order domain.Order
	var deliveredAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&order.ID,
		&order.UserID,
		&order.TotalAmount,
		&order.DiscountAmount,
		&order.FinalAmount,
		&order.DeliveryStatus,
		&order.OrderStatus,
		&order.HasReturnRequest,
		&order.ShippingAddressID,
		&order.CouponApplied,
		&order.CreatedAt,
		&order.UpdatedAt,
		&deliveredAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, utils.ErrOrderNotFound
		}
		log.Printf("Error retrieving order details from orders table : %v", err)
		return nil, err
	}

	// If order is delivered
	if deliveredAt.Valid {
		order.DeliveredAt = &deliveredAt.Time
	}

	return &order, nil
}

/*
GetOrderItems:
- Fetch order items details from order_items table
*/
func (r *orderRepository) GetOrderItems(ctx context.Context, orderID int64) ([]domain.OrderItem, error) {
	query := `
        SELECT id, order_id, product_id, quantity, price
        FROM order_items
        WHERE order_id = $1
    `
	rows, err := r.db.QueryContext(ctx, query, orderID)
	if err != nil {
		log.Printf("Error retrieving order items: %v", err)
		return nil, err
	}
	defer rows.Close()

	var items []domain.OrderItem
	// Iterate through each rows
	for rows.Next() {
		var item domain.OrderItem
		err := rows.Scan(&item.ID,
			&item.OrderID,
			&item.ProductID,
			&item.Quantity,
			&item.Price)
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

	return items, nil
}

func (r *orderRepository) CreateReturnRequestTx(ctx context.Context, tx *sql.Tx, returnRequest *domain.ReturnRequest) error {
	query := `
		INSERT INTO return_requests (order_id, user_id, return_reason, is_approved, requested_date)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`
	err := tx.QueryRowContext(ctx, query,
		returnRequest.OrderID,
		returnRequest.UserID,
		returnRequest.ReturnReason,
		returnRequest.IsApproved,
		returnRequest.RequestedDate,
	).Scan(&returnRequest.ID)
	return err
}

func (r *orderRepository) UpdateOrderHasReturnRequestTx(ctx context.Context, tx *sql.Tx, orderID int64, hasReturnRequest bool) error {
	query := `
		UPDATE orders
		SET has_return_request = $1
		WHERE id = $2
	`
	_, err := tx.ExecContext(ctx, query, hasReturnRequest, orderID)
	return err
}

func (r *orderRepository) GetOrderWithItems(ctx context.Context, orderID int64) (*domain.Order, error) {
	query := `
        SELECT o.id, o.user_id, o.total_amount, o.discount_amount, o.final_amount, 
               o.delivery_status, o.order_status, o.has_return_request, 
               o.shipping_address_id, o.coupon_applied, o.created_at, o.updated_at, o.delivered_at,
               oi.id, oi.product_id, oi.quantity, oi.price,
               p.name AS product_name,
               sa.address_line1, sa.address_line2, sa.city, sa.state, sa.pincode, sa.phone_number
        FROM orders o
        LEFT JOIN order_items oi ON o.id = oi.order_id
        LEFT JOIN products p ON oi.product_id = p.id
        LEFT JOIN shipping_addresses sa ON o.shipping_address_id = sa.id
        WHERE o.id = $1
    `
	rows, err := r.db.QueryContext(ctx, query, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var order *domain.Order
	itemMap := make(map[int64]*domain.OrderItem)

	for rows.Next() {
		if order == nil {
			order = &domain.Order{}
		}
		var item domain.OrderItem
		var productName string
		var deliveredAt sql.NullTime
		var addressLine1, addressLine2, city, state, pincode, phoneNumber sql.NullString
		err := rows.Scan(
			&order.ID, &order.UserID, &order.TotalAmount, &order.DiscountAmount, &order.FinalAmount,
			&order.DeliveryStatus, &order.OrderStatus, &order.HasReturnRequest,
			&order.ShippingAddressID, &order.CouponApplied, &order.CreatedAt, &order.UpdatedAt, &deliveredAt,
			&item.ID, &item.ProductID, &item.Quantity, &item.Price,
			&productName,
			&addressLine1, &addressLine2, &city, &state, &pincode, &phoneNumber,
		)
		if err != nil {
			return nil, err
		}
		if deliveredAt.Valid {
			order.DeliveredAt = &deliveredAt.Time
		}
		item.OrderID = order.ID
		item.ProductName = productName
		itemMap[item.ID] = &item

		if order.ShippingAddress == nil && addressLine1.Valid {
			order.ShippingAddress = &domain.ShippingAddress{
				AddressLine1: addressLine1.String,
				AddressLine2: addressLine2.String,
				City:         city.String,
				State:        state.String,
				PinCode:      pincode.String,
				PhoneNumber:  phoneNumber.String,
			}
		}
	}

	if order == nil {
		return nil, utils.ErrOrderNotFound
	}

	order.Items = make([]domain.OrderItem, 0, len(itemMap))
	for _, item := range itemMap {
		order.Items = append(order.Items, *item)
	}

	return order, nil
}

func (r *orderRepository) UpdateOrderHasReturnRequest(ctx context.Context, orderID int64, hasReturnRequest bool) error {
	query := `
		UPDATE orders
		SET has_return_request = $1, updated_at = NOW()
		WHERE id = $2
	`
	_, err := r.db.ExecContext(ctx, query, hasReturnRequest, orderID)
	if err != nil {
		return err
	}
	return nil
}

/*
UpdateOrderDeliveryStatus:
- updates delivery_status, order_status, delivered_at of a given order
*/
func (r *orderRepository) UpdateOrderDeliveryStatus(ctx context.Context, tx *sql.Tx, orderID int64, deliveryStatus, orderStatus string, deliveredAt *time.Time) error {
	query := `
        UPDATE orders
        SET delivery_status = $1, order_status = $2, delivered_at = $3, updated_at = NOW()
        WHERE id = $4
    `
	_, err := tx.ExecContext(ctx, query, deliveryStatus, orderStatus, deliveredAt, orderID)
	if err != nil {
		log.Printf("error while updating order status and delivery status for the given order: %v", err)
		return err
	}
	return nil
}

/*
IsOrderDelivered:
- Used to check if order exists and is not delivered
*/
func (r *orderRepository) IsOrderDelivered(ctx context.Context, orderID int64) (bool, error) {
	query := `SELECT delivery_status = 'delivered' FROM orders WHERE id = $1`
	var isDelivered bool
	err := r.db.QueryRowContext(ctx, query, orderID).Scan(&isDelivered)
	if err == sql.ErrNoRows {
		// If order not found , reply with false and error
		return false, utils.ErrOrderNotFound
	}
	return isDelivered, err
}

func (r *orderRepository) GetOrderByID(ctx context.Context, id int64) (*domain.Order, error) {
	query := `
        SELECT id, user_id, total_amount, discount_amount, final_amount, delivery_status, 
               order_status, has_return_request, shipping_address_id, coupon_applied, 
               created_at, updated_at, delivered_at
        FROM orders
        WHERE id = $1
    `
	var order domain.Order
	var deliveredAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&order.ID, &order.UserID, &order.TotalAmount, &order.DiscountAmount, &order.FinalAmount,
		&order.DeliveryStatus, &order.OrderStatus, &order.HasReturnRequest, &order.ShippingAddressID,
		&order.CouponApplied, &order.CreatedAt, &order.UpdatedAt, &deliveredAt,
	)

	if err == sql.ErrNoRows {
		return nil, utils.ErrOrderNotFound
	}
	if err != nil {
		return nil, err
	}

	if deliveredAt.Valid {
		order.DeliveredAt = &deliveredAt.Time
	}

	return &order, nil
}

/*
CreatePayment:
- Create payment record in the payments table
*/
func (r *orderRepository) CreatePayment(ctx context.Context, tx *sql.Tx, payment *domain.Payment) error {
	query := `
        INSERT INTO payments (order_id, amount, payment_method, payment_status, created_at, updated_at, razorpay_order_id)
        VALUES ($1, $2, $3, $4, $5, $6, $7)
        RETURNING id
    `
	err := tx.QueryRowContext(ctx, query,
		payment.OrderID,
		payment.Amount,
		payment.PaymentMethod,
		payment.Status,
		payment.CreatedAt,
		payment.UpdatedAt,
		payment.RazorpayOrderID,
	).Scan(&payment.ID)
	if err != nil {
		log.Printf("failed to create payment record in the payments table : %v", err)
		return err
	}
	return nil
}

/*
GetPaymentByOrderID:
- Get payment details from payments table using order id
*/
func (r *orderRepository) GetPaymentByOrderID(ctx context.Context, orderID int64) (*domain.Payment, error) {
	query := `
        SELECT id, order_id, amount, payment_method, payment_status, created_at, updated_at, 
               razorpay_order_id, razorpay_payment_id, razorpay_signature
        FROM payments
        WHERE order_id = $1
    `
	var payment domain.Payment
	var rzpPaymentID, rzpSignature sql.NullString
	err := r.db.QueryRowContext(ctx, query, orderID).Scan(
		&payment.ID,
		&payment.OrderID,
		&payment.Amount,
		&payment.PaymentMethod,
		&payment.Status,
		&payment.CreatedAt,
		&payment.UpdatedAt,
		&payment.RazorpayOrderID,
		&rzpPaymentID,
		&rzpSignature,
	)

	if rzpPaymentID.Valid {
		payment.RazorpayPaymentID = rzpPaymentID.String
	}
	if rzpSignature.Valid {
		payment.RazorpaySignature = rzpSignature.String
	}

	if err == sql.ErrNoRows {
		return nil, utils.ErrPaymentNotFound
	}
	if err != nil {
		return nil, err
	}
	return &payment, nil
}

func (r *orderRepository) CreateCancellationRequest(ctx context.Context, orderID, userID int64) error {
	query := `
        INSERT INTO cancellation_requests (order_id, user_id, created_at, cancellation_status)
        VALUES ($1, $2, $3, $4)
    `

	now := time.Now().UTC()

	_, err := r.db.ExecContext(ctx, query, orderID, userID, now, "pending_review")
	if err != nil {
		// Check for unique constraint violation
		if utils.IsDuplicateKeyError(err) {
			return utils.ErrCancellationRequestExists
		}
		log.Printf("Error creating cancellation request: %v", err)
		return err
	}

	return nil
}

/*
GetByIDTx:
- get order details from orders table using order id
*/
func (r *orderRepository) GetByIDTx(ctx context.Context, tx *sql.Tx, id int64) (*domain.Order, error) {
	query := `
        SELECT id, user_id, total_amount, discount_amount, final_amount, delivery_status, 
               order_status, has_return_request, shipping_address_id, coupon_applied, 
               created_at, updated_at, delivered_at, is_cancelled
        FROM orders
        WHERE id = $1
    `
	var order domain.Order
	var deliveredAt sql.NullTime

	err := tx.QueryRowContext(ctx, query, id).Scan(
		&order.ID,
		&order.UserID,
		&order.TotalAmount,
		&order.DiscountAmount,
		&order.FinalAmount,
		&order.DeliveryStatus,
		&order.OrderStatus,
		&order.HasReturnRequest,
		&order.ShippingAddressID,
		&order.CouponApplied,
		&order.CreatedAt,
		&order.UpdatedAt,
		&deliveredAt,
		&order.IsCancelled,
	)

	if err == sql.ErrNoRows {
		return nil, utils.ErrOrderNotFound
	}
	if err != nil {
		log.Printf("failed to get order: %v", err)
		return nil, err
	}

	if deliveredAt.Valid {
		order.DeliveredAt = &deliveredAt.Time
	}

	return &order, nil
}

func (r *orderRepository) GetOrderItemsTx(ctx context.Context, tx *sql.Tx, orderID int64) ([]*domain.OrderItem, error) {
	query := `
        SELECT id, order_id, product_id, quantity, price
        FROM order_items
        WHERE order_id = $1
    `
	rows, err := tx.QueryContext(ctx, query, orderID)
	if err != nil {
		return nil, fmt.Errorf("failed to query order items: %w", err)
	}
	defer rows.Close()

	var items []*domain.OrderItem
	for rows.Next() {
		item := &domain.OrderItem{}
		if err := rows.Scan(&item.ID, &item.OrderID, &item.ProductID, &item.Quantity, &item.Price); err != nil {
			return nil, fmt.Errorf("failed to scan order item: %w", err)
		}
		items = append(items, item)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating order items: %w", err)
	}

	return items, nil
}

/*
GetCancellationRequests:
- Get order details and cancellation request details
- orders with order status = 'pending_cancellation' is fetched
- orders and cancellation_requests tables joined using order id
*/
func (r *orderRepository) GetCancellationRequests(ctx context.Context, params domain.CancellationRequestParams) ([]*domain.CancellationRequest, int64, error) {
	query := `
        SELECT cr.id, cr.order_id, cr.user_id, cr.created_at, cr.cancellation_status, cr.is_stock_updated
        FROM cancellation_requests cr
        JOIN orders o ON cr.order_id = o.id
        WHERE o.order_status = 'pending_cancellation'
        ORDER BY cr.created_at DESC
        LIMIT $1 OFFSET $2
    `
	countQuery := `
        SELECT COUNT(*)
        FROM cancellation_requests cr
        JOIN orders o ON cr.order_id = o.id
        WHERE o.order_status = 'pending_cancellation'
    `

	// Calculate offset
	offset := (params.Page - 1) * params.Limit

	// Get total count
	var totalCount int64
	err := r.db.QueryRowContext(ctx, countQuery).Scan(&totalCount)
	if err != nil {
		log.Printf("Error getting total count: %v", err)
		return nil, 0, err
	}

	// Execute main query
	rows, err := r.db.QueryContext(ctx, query, params.Limit, offset)
	if err != nil {
		log.Printf("Error executing main query: %v", err)
		return nil, 0, err
	}
	defer rows.Close()

	var requests []*domain.CancellationRequest
	for rows.Next() {
		var req domain.CancellationRequest
		err := rows.Scan(&req.ID,
			&req.OrderID,
			&req.UserID,
			&req.CreatedAt,
			&req.CancellationRequestStatus,
			&req.IsStockUpdated)
		if err != nil {
			log.Printf("Error scanning row: %v", err)
			return nil, 0, err
		}
		requests = append(requests, &req)
	}

	if err = rows.Err(); err != nil {
		log.Printf("Error iterating over rows: %v", err)
		return nil, 0, err
	}

	return requests, totalCount, nil
}

/*
GetShippingAddress:
- Get shipping address related values from shipping_addresses table
*/
func (r *orderRepository) GetShippingAddress(ctx context.Context, addressID int64) (*domain.ShippingAddress, error) {
	query := `
        SELECT user_id,address_id, address_line1, address_line2, city, state,landmark, pincode, phone_number
        FROM shipping_addresses
        WHERE id = $1
    `
	var address domain.ShippingAddress
	var addressLine2, landmark sql.NullString
	err := r.db.QueryRowContext(ctx, query, addressID).Scan(
		&address.UserID,
		&address.AddressID,
		&address.AddressLine1,
		&addressLine2,
		&address.City,
		&address.State,
		&landmark,
		&address.PinCode,
		&address.PhoneNumber,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, utils.ErrAddressNotFound
		}
		log.Printf("failed to get shipping address details from shipping addresses table : %v", err)
		return nil, err
	}
	if addressLine2.Valid {
		address.AddressLine2 = addressLine2.String
	}
	if landmark.Valid {
		address.Landmark = landmark.String
	}
	return &address, nil
}

/*
UpdateOrderStatusAndSetCancelledTx:
- Update orders table
- order_status, is_cancelled
*/
func (r *orderRepository) UpdateOrderStatusAndSetCancelledTx(ctx context.Context, tx *sql.Tx, orderID int64, order_status, delivery_status string, isCancelled bool) error {
	query := `
		UPDATE orders
		SET order_status = $1, delivery_status = $2, is_cancelled = $3, updated_at = NOW()
		WHERE id = $4
	`
	_, err := tx.ExecContext(ctx, query,
		order_status,
		delivery_status,
		isCancelled,
		orderID)
	if err != nil {
		log.Printf("failed to update order status and is_cancelled: %v", err)
		return err
	}
	return nil
}

/*
CreateCancellationRequestTx:
- Create cancellation request entry in cancellation_requests table
*/
func (r *orderRepository) CreateCancellationRequestTx(ctx context.Context, tx *sql.Tx, request *domain.CancellationRequest) error {
	query := `
        INSERT INTO cancellation_requests (order_id, user_id, created_at, cancellation_status, is_stock_updated)
        VALUES ($1, $2, $3, $4, $5)
        RETURNING id
    `
	err := tx.QueryRowContext(ctx, query,
		request.OrderID,
		request.UserID,
		request.CreatedAt,
		request.CancellationRequestStatus,
		request.IsStockUpdated,
	).Scan(&request.ID)

	if err != nil {
		log.Printf("failed to create cancellation request: %v", err)
		return err
	}

	return nil
}

/*
UpdateCancellationRequestTx:
- Update is_stock_updated column in cancellation_requests table
*/
func (r *orderRepository) UpdateCancellationRequestTx(ctx context.Context, tx *sql.Tx, request *domain.CancellationRequest) error {
	query := `
        UPDATE cancellation_requests
        SET is_stock_updated = $1, cancellation_status = $2
        WHERE id = $3
    `
	_, err := tx.ExecContext(ctx, query, request.IsStockUpdated, request.CancellationRequestStatus, request.ID)
	if err != nil {
		log.Printf("failed to update cancellation request: %v", err)
	}

	return err
}

/*
GetCancellationRequestByOrderIDTx:
- get cancellation request details using order id
*/
func (r *orderRepository) GetCancellationRequestByOrderIDTx(ctx context.Context, tx *sql.Tx, orderID int64) (*domain.CancellationRequest, error) {
	query := `
        SELECT id, order_id, user_id, created_at, cancellation_status, is_stock_updated
        FROM cancellation_requests
        WHERE order_id = $1
    `
	var request domain.CancellationRequest
	err := tx.QueryRowContext(ctx, query, orderID).Scan(
		&request.ID,
		&request.OrderID,
		&request.UserID,
		&request.CreatedAt,
		&request.CancellationRequestStatus,
		&request.IsStockUpdated,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, utils.ErrCancellationRequestNotFound
		}
		log.Printf("error while getting cancellation request details : %v", err)
		return nil, err
	}
	return &request, nil
}
