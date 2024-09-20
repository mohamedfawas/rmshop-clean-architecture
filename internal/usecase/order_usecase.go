package usecase

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/jung-kurt/gofpdf"
	"github.com/mohamedfawas/rmshop-clean-architecture/internal/domain"
	"github.com/mohamedfawas/rmshop-clean-architecture/internal/repository"
	"github.com/mohamedfawas/rmshop-clean-architecture/pkg/payment/razorpay"
	"github.com/mohamedfawas/rmshop-clean-architecture/pkg/utils"
)

type OrderUseCase interface {
	GetOrderByID(ctx context.Context, userID, orderID int64) (*domain.Order, error)
	GetUserOrders(ctx context.Context, userID int64, page, limit int, sortBy, order, status string) ([]*domain.Order, int64, error)
	CancelOrder(ctx context.Context, userID, orderID int64) (*domain.OrderCancellationResult, error)
	GetOrders(ctx context.Context, params domain.OrderQueryParams) ([]*domain.Order, int64, error)
	UpdateOrderStatus(ctx context.Context, orderID int64, newStatus string) (*domain.OrderStatusUpdateResult, error)
	GetPaymentByOrderID(ctx context.Context, orderID int64) (*domain.Payment, error)
	CreatePayment(ctx context.Context, tx *sql.Tx, payment *domain.Payment) error
	UpdatePayment(ctx context.Context, payment *domain.Payment) error
	ProcessPayment(ctx context.Context, tx *sql.Tx, orderID int64, paymentMethod string, amount float64) (*domain.Payment, error)
	VerifyAndUpdateRazorpayPayment(ctx context.Context, input domain.RazorpayPaymentInput) error
	PlaceOrderRazorpay(ctx context.Context, userID, checkoutID int64) (*domain.Order, error)
	UpdateOrderRazorpayID(ctx context.Context, orderID int64, razorpayOrderID string) error
	InitiateReturn(ctx context.Context, userID, orderID int64, reason string) (*domain.ReturnRequest, error)
	PlaceOrderCOD(ctx context.Context, userID, checkoutID int64) (*domain.Order, error)
	GenerateInvoice(ctx context.Context, userID, orderID int64) ([]byte, error)
}

type orderUseCase struct {
	orderRepo       repository.OrderRepository
	checkoutRepo    repository.CheckoutRepository
	productRepo     repository.ProductRepository
	cartRepo        repository.CartRepository
	razorpayService *razorpay.Service
}

func NewOrderUseCase(orderRepo repository.OrderRepository,
	checkoutRepo repository.CheckoutRepository,
	productRepo repository.ProductRepository,
	cartRepo repository.CartRepository,
	razorpayKeyID, razorpaySecret string) OrderUseCase {
	return &orderUseCase{
		orderRepo:       orderRepo,
		checkoutRepo:    checkoutRepo,
		productRepo:     productRepo,
		cartRepo:        cartRepo,
		razorpayService: razorpay.NewService(razorpayKeyID, razorpaySecret),
	}
}

func (u *orderUseCase) GetOrderByID(ctx context.Context, userID, orderID int64) (*domain.Order, error) {
	order, err := u.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		if err == utils.ErrOrderNotFound {
			return nil, utils.ErrOrderNotFound
		}
		return nil, err
	}

	// If userID is provided (not 0), check if the order belongs to the user
	if userID != 0 && order.UserID != userID {
		return nil, utils.ErrUnauthorized
	}

	return order, nil
}

func (u *orderUseCase) GetUserOrders(ctx context.Context, userID int64, page, limit int, sortBy, order, status string) ([]*domain.Order, int64, error) {
	// Validate pagination parameters
	if page < 1 || limit < 1 {
		return nil, 0, utils.ErrInvalidPaginationParams
	}

	// Validate and set default values for sorting
	if sortBy == "" {
		sortBy = "created_at"
	}
	if order == "" {
		order = "desc"
	}
	if order != "asc" && order != "desc" {
		order = "desc"
	}

	// Call repository method to get orders
	orders, totalCount, err := u.orderRepo.GetUserOrders(ctx, userID, page, limit, sortBy, order, status)
	if err != nil {
		return nil, 0, err
	}

	return orders, totalCount, nil
}

func (u *orderUseCase) CancelOrder(ctx context.Context, userID, orderID int64) (*domain.OrderCancellationResult, error) {
	// Get the order
	order, err := u.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		if err == utils.ErrOrderNotFound {
			return nil, utils.ErrOrderNotFound
		}
		return nil, err
	}

	// Check if the order belongs to the user
	if order.UserID != userID {
		return nil, utils.ErrUnauthorized
	}

	// Check if the order is already cancelled
	if order.OrderStatus == "cancelled" {
		return nil, utils.ErrOrderAlreadyCancelled
	}

	// Check if the order is in a cancellable state
	if !isCancellable(order.OrderStatus) {
		return nil, utils.ErrOrderNotCancellable
	}

	// Check if the cancellation period has expired
	if time.Since(order.CreatedAt) > 24*time.Hour {
		return nil, utils.ErrCancellationPeriodExpired
	}

	// Perform the cancellation
	err = u.orderRepo.UpdateOrderStatus(ctx, orderID, "cancelled")
	if err != nil {
		return nil, err
	}

	// Initiate refund if necessary
	refundStatus := "not_applicable"
	payment, err := u.orderRepo.GetPaymentByOrderID(ctx, orderID)
	if err != nil {
		// Log the error but continue, as we don't want to fail the cancellation due to payment retrieval issues
		log.Printf("Error retrieving payment for order %d: %v", orderID, err)
	} else if payment != nil && payment.Status == "paid" {
		refundStatus = "initiated"
		// Here you would typically call a payment service to initiate the refund
		// For now, we'll just update the status
		err = u.orderRepo.UpdateRefundStatus(ctx, orderID, sql.NullString{String: refundStatus, Valid: true})
		if err != nil {
			// Log the error, but don't fail the cancellation
			log.Printf("Error updating refund status for order %d: %v", orderID, err)
		}
	}

	return &domain.OrderCancellationResult{
		OrderID:      orderID,
		RefundStatus: refundStatus,
	}, nil
}

func (u *orderUseCase) GetOrders(ctx context.Context, params domain.OrderQueryParams) ([]*domain.Order, int64, error) {
	// Validate and set default values
	if params.Page < 1 {
		params.Page = 1
	}
	if params.Limit < 1 {
		params.Limit = 10
	} else if params.Limit > 100 {
		params.Limit = 100
	}

	validSortFields := map[string]bool{"created_at": true, "total_amount": true, "order_status": true}
	if params.SortBy != "" && !validSortFields[params.SortBy] {
		return nil, 0, errors.New("invalid sort field")
	}

	if params.SortOrder != "" && params.SortOrder != "asc" && params.SortOrder != "desc" {
		params.SortOrder = "desc"
	}

	// Call repository method
	return u.orderRepo.GetOrders(ctx, params)
}

func (u *orderUseCase) UpdateOrderStatus(ctx context.Context, orderID int64, newStatus string) (*domain.OrderStatusUpdateResult, error) {
	order, err := u.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		return nil, err
	}

	if !isValidOrderStatus(newStatus) {
		return nil, utils.ErrInvalidOrderStatus
	}

	var refundStatus string
	if newStatus == "cancelled" {
		if order.OrderStatus == "cancelled" {
			return nil, utils.ErrOrderAlreadyCancelled
		}
		if !isCancellable(order.OrderStatus) {
			return nil, utils.ErrOrderNotCancellable
		}
		refundStatus = "not_applicable"
		payment, err := u.orderRepo.GetPaymentByOrderID(ctx, orderID)
		if err != nil {
			log.Printf("Error retrieving payment for order %d: %v", orderID, err)
		} else if payment != nil && payment.Status == "paid" {
			refundStatus = "initiated"
			err = u.orderRepo.UpdateRefundStatus(ctx, orderID, sql.NullString{String: refundStatus, Valid: true})
			if err != nil {
				return nil, err
			}
		}
	}

	if newStatus == "delivered" {
		now := time.Now().UTC()
		err = u.orderRepo.SetOrderDeliveredAt(ctx, orderID, &now)
		if err != nil {
			return nil, fmt.Errorf("failed to set delivered at time: %w", err)
		}
	}

	err = u.orderRepo.UpdateOrderStatus(ctx, orderID, newStatus)
	if err != nil {
		return nil, err
	}

	return &domain.OrderStatusUpdateResult{
		OrderID:      orderID,
		NewStatus:    newStatus,
		RefundStatus: refundStatus,
	}, nil
}

func isValidOrderStatus(status string) bool {
	validStatuses := []string{"pending", "processing", "shipped", "delivered", "cancelled"}
	for _, s := range validStatuses {
		if status == s {
			return true
		}
	}
	return false
}

func isCancellable(status string) bool {
	cancellableStatuses := map[string]bool{
		"pending":    true,
		"processing": true,
	}
	return cancellableStatuses[status]
}

func (u *orderUseCase) GetPaymentByOrderID(ctx context.Context, orderID int64) (*domain.Payment, error) {
	return u.orderRepo.GetPaymentByOrderID(ctx, orderID)
}

func (u *orderUseCase) CreatePayment(ctx context.Context, tx *sql.Tx, payment *domain.Payment) error {
	return u.orderRepo.CreatePayment(ctx, tx, payment)
}

func (u *orderUseCase) UpdatePayment(ctx context.Context, payment *domain.Payment) error {
	return u.orderRepo.UpdatePayment(ctx, payment)
}

func (u *orderUseCase) ProcessPayment(ctx context.Context, tx *sql.Tx, orderID int64, paymentMethod string, amount float64) (*domain.Payment, error) {
	payment := &domain.Payment{
		OrderID:       orderID,
		Amount:        amount,
		PaymentMethod: paymentMethod,
		Status:        utils.PaymentStatusPending,
		CreatedAt:     time.Now().UTC(),
		UpdatedAt:     time.Now().UTC(),
	}

	if paymentMethod == utils.PaymentMethodRazorpay {
		razorpayOrder, err := u.razorpayService.CreateOrder(int64(amount*100), "INR")
		if err != nil {
			return nil, fmt.Errorf("failed to create Razorpay order: %w", err)
		}
		payment.RazorpayOrderID = razorpayOrder.ID
		payment.Status = utils.PaymentStatusAwaitingPayment
	} else if paymentMethod != utils.PaymentMethodCOD {
		return nil, fmt.Errorf("unsupported payment method: %s", paymentMethod)
	}

	// Create the payment record in the database
	err := u.orderRepo.CreatePayment(ctx, tx, payment)
	if err != nil {
		return nil, fmt.Errorf("failed to create payment record: %w", err)
	}

	log.Printf("Payment record created: %+v", payment)

	return payment, nil
}

func (u *orderUseCase) UpdateOrderRazorpayID(ctx context.Context, orderID int64, razorpayOrderID string) error {
	return u.orderRepo.UpdateOrderRazorpayID(ctx, orderID, razorpayOrderID)
}

func (u *orderUseCase) PlaceOrderRazorpay(ctx context.Context, userID, checkoutID int64) (*domain.Order, error) {
	tx, err := u.orderRepo.BeginTx(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	checkout, err := u.checkoutRepo.GetCheckoutByID(ctx, checkoutID)
	if err != nil {
		if err == utils.ErrCheckoutNotFound {
			return nil, utils.ErrCheckoutNotFound
		}
		return nil, err
	}

	if checkout.UserID != userID {
		return nil, utils.ErrUnauthorized
	}

	if checkout.Status != utils.CheckoutStatusPending {
		return nil, utils.ErrOrderAlreadyPlaced
	}

	items, err := u.checkoutRepo.GetCheckoutItems(ctx, checkoutID)
	if err != nil {
		return nil, err
	}

	if len(items) == 0 {
		return nil, utils.ErrEmptyCart
	}

	for _, item := range items {
		product, err := u.productRepo.GetByID(ctx, item.ProductID)
		if err != nil {
			return nil, err
		}
		if product.StockQuantity < item.Quantity {
			return nil, utils.ErrInsufficientStock
		}
	}

	if checkout.ShippingAddressID == 0 {
		return nil, utils.ErrInvalidAddress
	}

	order := &domain.Order{
		UserID:            userID,
		TotalAmount:       checkout.TotalAmount,
		DiscountAmount:    checkout.DiscountAmount,
		FinalAmount:       checkout.FinalAmount,
		OrderStatus:       utils.OrderStatusPending,
		DeliveryStatus:    utils.DeliveryStatusPending,
		ShippingAddressID: checkout.ShippingAddressID,
		CouponApplied:     checkout.CouponApplied,
	}

	orderID, err := u.orderRepo.CreateOrder(ctx, tx, order)
	if err != nil {
		return nil, fmt.Errorf("failed to create order: %w", err)
	}
	order.ID = orderID

	payment := &domain.Payment{
		OrderID:       order.ID,
		Amount:        checkout.FinalAmount,
		PaymentMethod: utils.PaymentMethodRazorpay,
		Status:        utils.PaymentStatusPending,
	}

	err = u.orderRepo.CreatePayment(ctx, tx, payment)
	if err != nil {
		return nil, fmt.Errorf("failed to create payment: %w", err)
	}

	for _, item := range items {
		orderItem := &domain.OrderItem{
			OrderID:   order.ID,
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
			Price:     item.Price,
		}
		err = u.orderRepo.AddOrderItem(ctx, tx, orderItem)
		if err != nil {
			return nil, fmt.Errorf("failed to add order item: %w", err)
		}

		err = u.productRepo.UpdateStock(ctx, tx, item.ProductID, -item.Quantity)
		if err != nil {
			return nil, fmt.Errorf("failed to update product stock: %w", err)
		}
	}

	checkout.Status = utils.CheckoutStatusCompleted
	err = u.checkoutRepo.UpdateCheckoutStatus(ctx, tx, checkout)
	if err != nil {
		return nil, fmt.Errorf("failed to update checkout status: %w", err)
	}

	err = u.cartRepo.ClearCart(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to clear user's cart: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return order, nil
}

func (u *orderUseCase) VerifyAndUpdateRazorpayPayment(ctx context.Context, input domain.RazorpayPaymentInput) error {
	log.Printf("Verifying and updating Razorpay payment for order ID: %s", input.OrderID)

	payment, err := u.orderRepo.GetPaymentByRazorpayOrderID(ctx, input.OrderID)
	if err != nil {
		if err == utils.ErrPaymentNotFound {
			log.Printf("Payment not found for Razorpay order ID: %s", input.OrderID)
			return fmt.Errorf("payment not found for Razorpay order ID %s", input.OrderID)
		}
		log.Printf("Error getting payment by Razorpay order ID: %v", err)
		return fmt.Errorf("failed to retrieve payment: %w", err)
	}

	log.Printf("Payment found for Razorpay order ID %s: %+v", input.OrderID, payment)

	attributes := map[string]interface{}{
		"razorpay_order_id":   input.OrderID,
		"razorpay_payment_id": input.PaymentID,
		"razorpay_signature":  input.Signature,
	}

	if err := u.razorpayService.VerifyPaymentSignature(attributes); err != nil {
		log.Printf("Invalid Razorpay signature: %v", err)
		return errors.New("invalid signature")
	}

	payment.Status = utils.PaymentStatusPaid
	payment.RazorpayPaymentID = input.PaymentID
	payment.RazorpaySignature = input.Signature

	err = u.orderRepo.UpdatePayment(ctx, payment)
	if err != nil {
		log.Printf("Error updating payment: %v", err)
		return fmt.Errorf("failed to update payment: %w", err)
	}

	log.Printf("Payment updated successfully: %+v", payment)

	// Update order status
	err = u.orderRepo.UpdateOrderStatus(ctx, payment.OrderID, utils.OrderStatusProcessing)
	if err != nil {
		log.Printf("Failed to update order status: %v", err)
		// Don't return the error here, as the payment was successful
	} else {
		log.Printf("Order status updated to 'paid' for order ID: %d", payment.OrderID)
	}

	log.Printf("Payment successfully verified and updated for order ID: %s", input.OrderID)
	return nil
}

func (u *orderUseCase) InitiateReturn(ctx context.Context, userID, orderID int64, reason string) (*domain.ReturnRequest, error) {
	// Get the order
	order, err := u.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		if err == utils.ErrOrderNotFound {
			return nil, utils.ErrOrderNotFound
		}
		return nil, err
	}

	// Check if the order belongs to the user
	if order.UserID != userID {
		return nil, utils.ErrUnauthorized
	}

	// Check if the order is in 'delivered' status
	if order.OrderStatus != "delivered" {
		return nil, utils.ErrOrderNotEligibleForReturn
	}

	// Check if the return window has expired (e.g., 14 days)
	if order.DeliveredAt == nil {
		return nil, utils.ErrOrderNotEligibleForReturn
	}
	if time.Since(*order.DeliveredAt) > 14*24*time.Hour {
		return nil, utils.ErrReturnWindowExpired
	}

	// Check if a return request already exists
	if order.HasReturnRequest {
		return nil, utils.ErrReturnAlreadyRequested
	}

	// Validate return reason
	if reason == "" {
		return nil, utils.ErrInvalidReturnReason
	}

	// Start a database transaction
	tx, err := u.orderRepo.BeginTx(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	// Create return request
	returnRequest := &domain.ReturnRequest{
		OrderID:       orderID,
		UserID:        userID,
		ReturnReason:  reason,
		IsApproved:    false,
		RequestedDate: time.Now().UTC(),
	}

	// Save return request
	err = u.orderRepo.CreateReturnRequestTx(ctx, tx, returnRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to create return request: %w", err)
	}

	// Update order's has_return_request flag
	err = u.orderRepo.UpdateOrderHasReturnRequestTx(ctx, tx, orderID, true)
	if err != nil {
		return nil, fmt.Errorf("failed to update order return request status: %w", err)
	}

	// Commit the transaction
	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return returnRequest, nil
}

func (u *orderUseCase) PlaceOrderCOD(ctx context.Context, userID, checkoutID int64) (*domain.Order, error) {
	// Start a database transaction
	tx, err := u.orderRepo.BeginTx(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	// Get the checkout session
	checkout, err := u.checkoutRepo.GetCheckoutByID(ctx, checkoutID)
	if err != nil {
		if err == utils.ErrCheckoutNotFound {
			return nil, utils.ErrCheckoutNotFound
		}
		return nil, err
	}

	// Verify the checkout belongs to the user
	if checkout.UserID != userID {
		return nil, utils.ErrUnauthorized
	}

	// Check if the checkout is in a valid state to place an order
	if checkout.Status != "pending" {
		return nil, utils.ErrOrderAlreadyPlaced
	}

	// Check if the order total exceeds the COD limit
	if checkout.FinalAmount > utils.CODLimit {
		return nil, utils.ErrCODLimitExceeded
	}

	// Get checkout items
	items, err := u.checkoutRepo.GetCheckoutItems(ctx, checkoutID)
	if err != nil {
		return nil, err
	}

	if len(items) == 0 {
		return nil, utils.ErrEmptyCart
	}

	// Verify that all items have sufficient stock
	for _, item := range items {
		product, err := u.productRepo.GetByID(ctx, item.ProductID)
		if err != nil {
			return nil, err
		}
		if product.StockQuantity < item.Quantity {
			return nil, utils.ErrInsufficientStock
		}
	}

	// Verify that a valid address is associated with the checkout
	if checkout.ShippingAddressID == 0 {
		return nil, utils.ErrInvalidAddress
	}

	// Create the order
	order := &domain.Order{
		UserID:            userID,
		TotalAmount:       checkout.FinalAmount,
		DiscountAmount:    checkout.DiscountAmount,
		FinalAmount:       checkout.FinalAmount,
		OrderStatus:       "pending",
		DeliveryStatus:    "processing",
		ShippingAddressID: checkout.ShippingAddressID,
		CouponApplied:     checkout.CouponApplied,
		CreatedAt:         time.Now().UTC(),
		UpdatedAt:         time.Now().UTC(),
	}

	// Create the order in the database
	orderID, err := u.orderRepo.CreateOrder(ctx, tx, order)
	if err != nil {
		return nil, fmt.Errorf("failed to create order: %w", err)
	}
	order.ID = orderID

	// Create a payment record
	payment := &domain.Payment{
		OrderID:       order.ID,
		Amount:        order.FinalAmount,
		PaymentMethod: "cod",
		Status:        "pending",
		CreatedAt:     time.Now().UTC(),
		UpdatedAt:     time.Now().UTC(),
	}
	err = u.orderRepo.CreatePayment(ctx, tx, payment)
	if err != nil {
		return nil, fmt.Errorf("failed to create payment record: %w", err)
	}

	// Create order items and update product stock
	for _, item := range items {
		orderItem := &domain.OrderItem{
			OrderID:   order.ID,
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
			Price:     item.Price,
		}
		err = u.orderRepo.AddOrderItem(ctx, tx, orderItem)
		if err != nil {
			return nil, fmt.Errorf("failed to add order item: %w", err)
		}

		// Update product stock
		err = u.productRepo.UpdateStock(ctx, tx, item.ProductID, -item.Quantity)
		if err != nil {
			return nil, fmt.Errorf("failed to update product stock: %w", err)
		}
	}

	// Update checkout status
	checkout.Status = "completed"
	err = u.checkoutRepo.UpdateCheckoutStatus(ctx, tx, checkout)
	if err != nil {
		return nil, fmt.Errorf("failed to update checkout status: %w", err)
	}

	// Clear the user's cart
	err = u.cartRepo.ClearCart(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to clear user's cart: %w", err)
	}

	// Commit the transaction
	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return order, nil
}

func (u *orderUseCase) GenerateInvoice(ctx context.Context, userID, orderID int64) ([]byte, error) {
	order, err := u.orderRepo.GetOrderWithItems(ctx, orderID)
	if err != nil {
		return nil, err
	}

	if order.UserID != userID {
		return nil, utils.ErrUnauthorized
	}

	if order.OrderStatus == "cancelled" {
		return nil, utils.ErrCancelledOrder
	}

	if order.OrderStatus != "delivered" && order.OrderStatus != "completed" {
		return nil, utils.ErrUnpaidOrder
	}

	if len(order.Items) == 0 {
		return nil, utils.ErrEmptyOrder
	}

	// Generate PDF
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

	// Add header
	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(40, 10, "Invoice")

	// Add order details
	pdf.Ln(10)
	pdf.SetFont("Arial", "", 12)
	pdf.Cell(40, 10, fmt.Sprintf("Order ID: %d", order.ID))
	pdf.Ln(10)
	pdf.Cell(40, 10, fmt.Sprintf("Date: %s", order.CreatedAt.Format("2006-01-02 15:04:05")))

	// Add items
	pdf.Ln(10)
	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(80, 10, "Product")
	pdf.Cell(40, 10, "Quantity")
	pdf.Cell(40, 10, "Price")
	pdf.Cell(40, 10, "Subtotal")

	for _, item := range order.Items {
		pdf.Ln(10)
		pdf.SetFont("Arial", "", 12)
		pdf.Cell(80, 10, fmt.Sprintf("Product ID: %d", item.ProductID))
		pdf.Cell(40, 10, fmt.Sprintf("%d", item.Quantity))
		pdf.Cell(40, 10, fmt.Sprintf("$%.2f", item.Price))
		pdf.Cell(40, 10, fmt.Sprintf("$%.2f", float64(item.Quantity)*item.Price))
	}

	// Add total
	pdf.Ln(10)
	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(160, 10, "Total")
	pdf.Cell(40, 10, fmt.Sprintf("$%.2f", order.FinalAmount))

	// Get PDF as bytes
	var buf bytes.Buffer
	err = pdf.Output(&buf)
	if err != nil {
		return nil, fmt.Errorf("failed to generate PDF: %w", err)
	}

	return buf.Bytes(), nil
}
