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
	GetOrders(ctx context.Context, params domain.OrderQueryParams) ([]*domain.Order, int64, error)
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
	UpdateOrderDeliveryStatus(ctx context.Context, orderID int64, deliveryStatus, orderStatus string) error
	CreateRazorpayOrder(ctx context.Context, amount float64, currency string) (*razorpay.Order, error)
	GetRazorpayKeyID() string
	GetPaymentByRazorpayOrderID(ctx context.Context, razorpayOrderID string) (*domain.Payment, error)
	CancelOrder(ctx context.Context, userID, orderID int64) (*domain.OrderCancellationResult, error)
	ApproveCancellation(ctx context.Context, orderID int64) (*domain.OrderStatusUpdateResult, error)
	AdminCancelOrder(ctx context.Context, orderID int64) (*domain.AdminOrderCancellationResult, error)
	GetCancellationRequests(ctx context.Context, params domain.CancellationRequestParams) ([]*domain.CancellationRequest, int64, error)
}

type orderUseCase struct {
	orderRepo       repository.OrderRepository
	checkoutRepo    repository.CheckoutRepository
	productRepo     repository.ProductRepository
	cartRepo        repository.CartRepository
	walletRepo      repository.WalletRepository
	paymentRepo     repository.PaymentRepository
	razorpayService *razorpay.Service
}

func NewOrderUseCase(orderRepo repository.OrderRepository,
	checkoutRepo repository.CheckoutRepository,
	productRepo repository.ProductRepository,
	cartRepo repository.CartRepository,
	walletRepo repository.WalletRepository,
	paymentRepo repository.PaymentRepository,
	razorpayKeyID, razorpaySecret string) OrderUseCase {
	return &orderUseCase{
		orderRepo:       orderRepo,
		checkoutRepo:    checkoutRepo,
		productRepo:     productRepo,
		cartRepo:        cartRepo,
		walletRepo:      walletRepo,
		paymentRepo:     paymentRepo,
		razorpayService: razorpay.NewService(razorpayKeyID, razorpaySecret),
	}
}

func (u *orderUseCase) GetOrderByID(ctx context.Context, userID, orderID int64) (*domain.Order, error) {
	order, err := u.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		if err == utils.ErrOrderNotFound {
			return nil, utils.ErrOrderNotFound
		}
		log.Printf("error getting order details using order id : %v", err)
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
		log.Printf("failed to start database transaction in PlaceOrderCOD method : %v", err)
		return nil, err
	}
	defer tx.Rollback()

	// Get the checkout session details using checkout id
	checkout, err := u.checkoutRepo.GetCheckoutByID(ctx, checkoutID)
	if err != nil {
		if err == utils.ErrCheckoutNotFound {
			return nil, utils.ErrCheckoutNotFound
		}
		log.Printf("error while retrieving checkout details using checkout id : %v", err)
		return nil, err
	}

	// Verify the checkout belongs to the user
	if checkout.UserID != userID {
		return nil, utils.ErrUnauthorized
	}

	// Check if the checkout is in a valid state to place an order
	if checkout.Status != utils.CheckoutStatusPending {
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

	// If checkout is empty , then you can't place the order
	if len(items) == 0 {
		return nil, utils.ErrEmptyCart
	}

	// Verify that all items have sufficient stock
	for _, item := range items {
		// Retrieve current stock availability of each product
		product, err := u.productRepo.GetByID(ctx, item.ProductID)
		if err != nil {
			log.Printf("error while retrieiving product details to check stock availability : %v", err)
			return nil, err
		}
		if product.StockQuantity < item.Quantity {
			log.Printf("insufficient stock for the product id : %v", product.ID)
			return nil, utils.ErrInsufficientStock
		}
	}

	// Verify that a valid address is associated with the checkout
	if checkout.ShippingAddressID == 0 {
		return nil, utils.ErrInvalidAddress
	}

	// Create the order
	now := time.Now().UTC() // record the current time
	order := &domain.Order{
		UserID:            userID,
		TotalAmount:       checkout.TotalAmount,
		DiscountAmount:    checkout.DiscountAmount,
		FinalAmount:       checkout.FinalAmount,
		OrderStatus:       utils.OrderStatusPending,
		DeliveryStatus:    utils.DeliveryStatusPending,
		ShippingAddressID: checkout.ShippingAddressID,
		CouponApplied:     checkout.CouponApplied,
		CreatedAt:         now,
		UpdatedAt:         now,
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
		PaymentMethod: utils.PaymentMethodCOD,
		Status:        utils.PaymentStatusPending,
		CreatedAt:     now,
		UpdatedAt:     now,
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
	checkout.Status = utils.CheckoutStatusCompleted
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

	if order.OrderStatus == utils.OrderStatusCancelled {
		return nil, utils.ErrCancelledOrder
	}

	if order.OrderStatus != utils.OrderStatusCompleted && order.OrderStatus != utils.OrderStatusConfirmed {
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

	// Add shipping address
	pdf.Ln(10)
	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(40, 10, "Shipping Address:")
	pdf.Ln(6)
	pdf.SetFont("Arial", "", 10)
	pdf.Cell(40, 6, order.ShippingAddress.AddressLine1)
	if order.ShippingAddress.AddressLine2 != "" {
		pdf.Ln(6)
		pdf.Cell(40, 6, order.ShippingAddress.AddressLine2)
	}
	pdf.Ln(6)
	pdf.Cell(40, 6, fmt.Sprintf("%s, %s, %s", order.ShippingAddress.City, order.ShippingAddress.State, order.ShippingAddress.PinCode))
	pdf.Ln(6)
	pdf.Cell(40, 6, fmt.Sprintf("Phone: %s", order.ShippingAddress.PhoneNumber))

	// Add items
	pdf.Ln(10)
	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(80, 10, "Product")
	pdf.Cell(30, 10, "Quantity")
	pdf.Cell(40, 10, "Price")
	pdf.Cell(40, 10, "Subtotal")

	for _, item := range order.Items {
		pdf.Ln(8)
		pdf.SetFont("Arial", "", 10)
		pdf.Cell(80, 8, fmt.Sprintf("%s (ID: %d)", item.ProductName, item.ProductID))
		pdf.Cell(30, 8, fmt.Sprintf("%d", item.Quantity))
		pdf.Cell(40, 8, fmt.Sprintf("$%.2f", item.Price))
		pdf.Cell(40, 8, fmt.Sprintf("$%.2f", float64(item.Quantity)*item.Price))
	}

	// Add total, discount, and final amount
	pdf.Ln(10)
	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(150, 8, "Total Amount")
	pdf.Cell(40, 8, fmt.Sprintf("$%.2f", order.TotalAmount))

	if order.DiscountAmount > 0 {
		pdf.Ln(8)
		pdf.Cell(150, 8, "Discount")
		pdf.Cell(40, 8, fmt.Sprintf("-$%.2f", order.DiscountAmount))
	}

	pdf.Ln(8)
	pdf.SetFont("Arial", "B", 14)
	pdf.Cell(150, 10, "Final Amount")
	pdf.Cell(40, 10, fmt.Sprintf("$%.2f", order.FinalAmount))

	// Get PDF as bytes
	var buf bytes.Buffer
	err = pdf.Output(&buf)
	if err != nil {
		return nil, fmt.Errorf("failed to generate PDF: %w", err)
	}

	return buf.Bytes(), nil
}

func (u *orderUseCase) UpdateOrderDeliveryStatus(ctx context.Context, orderID int64, deliveryStatus, orderStatus string) error {
	// Validate delivery status
	if !isValidDeliveryStatus(deliveryStatus) {
		return utils.ErrInvalidDeliveryStatus
	}

	// Validate order status
	if orderStatus != "" && !isValidOrderStatus(orderStatus) {
		return utils.ErrInvalidOrderStatus
	}

	// Check if the order exists and is not already delivered
	isDelivered, err := u.orderRepo.IsOrderDelivered(ctx, orderID)
	if err != nil {
		return err
	}
	if isDelivered {
		return utils.ErrOrderAlreadyDelivered
	}

	var deliveredAt *time.Time
	if deliveryStatus == "delivered" {
		now := time.Now().UTC()
		deliveredAt = &now
		if orderStatus == "" {
			orderStatus = "completed"
		}
	}

	return u.orderRepo.UpdateOrderDeliveryStatus(ctx, orderID, deliveryStatus, orderStatus, deliveredAt)
}

func isValidDeliveryStatus(status string) bool {
	validStatuses := []string{"pending", "in_transit", "out_for_delivery", "delivered", "failed_attempt", "returned_to_sender"}
	for _, s := range validStatuses {
		if status == s {
			return true
		}
	}
	return false
}

func isValidOrderStatus(status string) bool {
	validStatuses := []string{"pending", "processing", "shipped", "completed", "cancelled"}
	for _, s := range validStatuses {
		if status == s {
			return true
		}
	}
	return false
}

func (u *orderUseCase) CreateRazorpayOrder(ctx context.Context, amount float64, currency string) (*razorpay.Order, error) {
	return u.razorpayService.CreateOrder(int64(amount), currency)
}

func (u *orderUseCase) GetRazorpayKeyID() string {
	return u.razorpayService.GetKeyID()
}

func (u *orderUseCase) GetPaymentByRazorpayOrderID(ctx context.Context, razorpayOrderID string) (*domain.Payment, error) {
	log.Printf("razorpay order id used to get payment details : %v", razorpayOrderID)
	payment, err := u.paymentRepo.GetByRazorpayOrderID(ctx, razorpayOrderID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, utils.ErrPaymentNotFound
		}
		log.Printf("error while getting payment using razorpay order id : %v", err)
		return nil, err
	}
	return payment, nil
}

func (u *orderUseCase) VerifyAndUpdateRazorpayPayment(ctx context.Context, input domain.RazorpayPaymentInput) error {

	// Verify the payment signature
	attributes := map[string]interface{}{
		"razorpay_order_id":   input.OrderID,
		"razorpay_payment_id": input.PaymentID,
		"razorpay_signature":  input.Signature,
	}
	if err := u.razorpayService.VerifyPaymentSignature(attributes); err != nil {
		log.Printf("error while verifying payment signature : %v", err)
		return err
	}

	// Get the payment by Razorpay order ID
	payment, err := u.paymentRepo.GetByRazorpayOrderID(ctx, input.OrderID)
	if err != nil {
		log.Printf("error while retrieving payment using order id  : %v", err)
		return err
	}

	// Update the payment status
	payment.Status = utils.PaymentStatusPaid
	payment.RazorpayPaymentID = input.PaymentID
	payment.RazorpaySignature = input.Signature

	if err := u.paymentRepo.UpdatePayment(ctx, payment); err != nil {
		log.Printf("error while updating payment : %v", err)
		return err
	}

	// Update the order status
	if err := u.orderRepo.UpdateOrderStatus(ctx, payment.OrderID, utils.OrderStatusConfirmed); err != nil {
		log.Printf("error while update order status : %v", err)
		return err
	}

	return nil
}

func (u *orderUseCase) CancelOrder(ctx context.Context, userID, orderID int64) (*domain.OrderCancellationResult, error) {
	order, err := u.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, utils.ErrOrderNotFound
		}
		return nil, err
	}

	if order.UserID != userID {
		return nil, utils.ErrUnauthorized
	}

	if order.OrderStatus == "cancelled" {
		return nil, utils.ErrOrderAlreadyCancelled
	}

	if order.OrderStatus != "processing" && order.OrderStatus != "pending_payment" && order.OrderStatus != "confirmed" {
		return nil, utils.ErrOrderNotCancellable
	}

	result := &domain.OrderCancellationResult{
		OrderID:             orderID,
		RequiresAdminReview: order.OrderStatus == "processing" || order.OrderStatus == "confirmed",
	}

	if order.OrderStatus == "pending_payment" {
		err = u.orderRepo.UpdateOrderStatus(ctx, orderID, "cancelled")
		if err != nil {
			return nil, err
		}
		result.OrderStatus = "cancelled"
	} else {
		err = u.orderRepo.UpdateOrderStatus(ctx, orderID, "pending_cancellation")
		if err != nil {
			return nil, err
		}
		err = u.orderRepo.CreateCancellationRequest(ctx, orderID, userID)
		if err != nil {
			return nil, err
		}
		result.OrderStatus = "pending_cancellation"
	}

	return result, nil
}

func (u *orderUseCase) ApproveCancellation(ctx context.Context, orderID int64) (*domain.OrderStatusUpdateResult, error) {

	// Start a database transaction
	tx, err := u.orderRepo.BeginTx(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	// Get the order
	order, err := u.orderRepo.GetByIDTx(ctx, tx, orderID)
	if err != nil {
		if err == utils.ErrOrderNotFound {
			return nil, utils.ErrOrderNotFound
		}
		return nil, fmt.Errorf("failed to get order: %w", err)
	}

	// Check if the order is in pending cancellation state
	if order.OrderStatus != "pending_cancellation" {
		return nil, utils.ErrOrderNotPendingCancellation
	}

	// Update order status to cancelled
	err = u.orderRepo.UpdateOrderStatusTx(ctx, tx, orderID, "cancelled")
	if err != nil {
		return nil, fmt.Errorf("failed to update order status: %w", err)
	}

	result := &domain.OrderStatusUpdateResult{
		OrderID:      orderID,
		NewStatus:    "cancelled",
		RefundStatus: "not_applicable",
	}

	// Check if refund is needed
	payment, err := u.paymentRepo.GetByOrderIDTx(ctx, tx, orderID)
	if err != nil {
		if err != utils.ErrPaymentNotFound {
			return nil, fmt.Errorf("failed to get payment: %w", err)
		}
		// If payment not found, we assume no refund is needed
		payment = nil
	}

	if payment != nil && payment.Status == "paid" && payment.PaymentMethod == "razorpay" {
		// Process refund
		err = u.processRefund(ctx, tx, order, payment)
		if err != nil {
			return nil, fmt.Errorf("failed to process refund: %w", err)
		}
		result.RefundStatus = "initiated"
	}

	// Update inventory (add back the quantities)
	err = u.updateInventory(ctx, tx, orderID)
	if err != nil {
		return nil, fmt.Errorf("failed to update inventory: %w", err)
	}

	// Commit the transaction
	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return result, nil
}

func (u *orderUseCase) processRefund(ctx context.Context, tx *sql.Tx, order *domain.Order, payment *domain.Payment) error {
	if u.walletRepo == nil {
		return fmt.Errorf("wallet repository is nil")
	}

	// Get the user's wallet
	wallet, err := u.walletRepo.GetWalletTx(ctx, tx, order.UserID)
	if err != nil {
		if err == sql.ErrNoRows {
			// If wallet doesn't exist, create one
			wallet = &domain.Wallet{
				UserID:  order.UserID,
				Balance: 0,
			}
			err = u.walletRepo.CreateWalletTx(ctx, tx, wallet)
			if err != nil {
				return fmt.Errorf("failed to create wallet: %w", err)
			}
		} else {
			return fmt.Errorf("failed to get wallet: %w", err)
		}
	}

	// Calculate new balance
	newBalance := wallet.Balance + payment.Amount

	// Create a wallet transaction for the refund
	referenceType := "order_cancellation"
	walletTx := &domain.WalletTransaction{
		UserID:          order.UserID,
		Amount:          payment.Amount,
		TransactionType: "refund",
		ReferenceID:     &order.ID,
		ReferenceType:   &referenceType,
		BalanceAfter:    newBalance,
		CreatedAt:       time.Now().UTC(),
	}

	err = u.walletRepo.CreateWalletTransactionTx(ctx, tx, walletTx)
	if err != nil {
		return fmt.Errorf("failed to create wallet transaction: %w", err)
	}

	// Update user's wallet balance
	err = u.walletRepo.UpdateWalletBalanceTx(ctx, tx, order.UserID, newBalance)
	if err != nil {
		return fmt.Errorf("failed to update wallet balance: %w", err)
	}

	if u.paymentRepo == nil {
		return fmt.Errorf("payment repository is nil")
	}

	// Update payment status
	err = u.paymentRepo.UpdateStatusTx(ctx, tx, payment.ID, "refunded")
	if err != nil {
		return fmt.Errorf("failed to update payment status: %w", err)
	}

	return nil
}

func (u *orderUseCase) updateInventory(ctx context.Context, tx *sql.Tx, orderID int64) error {
	// Get order items
	orderItems, err := u.orderRepo.GetOrderItemsTx(ctx, tx, orderID)
	if err != nil {
		return fmt.Errorf("failed to get order items: %w", err)
	}

	// Update inventory for each item
	for _, item := range orderItems {
		err = u.productRepo.UpdateStockTx(ctx, tx, item.ProductID, item.Quantity)
		if err != nil {
			return fmt.Errorf("failed to update stock for product %d: %w", item.ProductID, err)
		}
	}

	return nil
}

func (u *orderUseCase) AdminCancelOrder(ctx context.Context, orderID int64) (*domain.AdminOrderCancellationResult, error) {
	// Start a database transaction
	tx, err := u.orderRepo.BeginTx(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	// Get the order
	order, err := u.orderRepo.GetByIDTx(ctx, tx, orderID)
	if err != nil {
		if err == utils.ErrOrderNotFound {
			return nil, utils.ErrOrderNotFound
		}
		return nil, fmt.Errorf("failed to get order: %w", err)
	}

	// Check if the order is cancellable
	if order.OrderStatus == "cancelled" {
		return nil, utils.ErrOrderAlreadyCancelled
	}
	if !isOrderCancellable(order.OrderStatus) {
		return nil, utils.ErrOrderNotCancellable
	}

	// Update order status to cancelled
	err = u.orderRepo.UpdateOrderStatusTx(ctx, tx, orderID, "cancelled")
	if err != nil {
		return nil, fmt.Errorf("failed to update order status: %w", err)
	}

	result := &domain.AdminOrderCancellationResult{
		OrderID:             orderID,
		OrderStatus:         "cancelled",
		RequiresAdminReview: false,
		RefundInitiated:     false,
	}

	// Check if refund is needed
	payment, err := u.paymentRepo.GetByOrderIDTx(ctx, tx, orderID)
	if err != nil {
		if err != utils.ErrPaymentNotFound {
			return nil, fmt.Errorf("failed to get payment: %w", err)
		}
		// If payment not found, we assume no refund is needed
		payment = nil
	}

	if payment != nil && payment.Status == "paid" && payment.PaymentMethod == "razorpay" {
		// Process refund
		err = u.processRefund(ctx, tx, order, payment)
		if err != nil {
			return nil, fmt.Errorf("failed to process refund: %w", err)
		}
		result.RefundInitiated = true
	}

	// Update inventory (add back the quantities)
	err = u.updateInventory(ctx, tx, orderID)
	if err != nil {
		return nil, fmt.Errorf("failed to update inventory: %w", err)
	}

	// Commit the transaction
	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return result, nil
}

func isOrderCancellable(status string) bool {
	cancellableStatuses := []string{"pending_payment", "confirmed", "processing"}
	for _, s := range cancellableStatuses {
		if status == s {
			return true
		}
	}
	return false
}

func (u *orderUseCase) GetCancellationRequests(ctx context.Context, params domain.CancellationRequestParams) ([]*domain.CancellationRequest, int64, error) {
	// Validate and set default values
	if params.Page < 1 {
		params.Page = 1
	}
	if params.Limit < 1 {
		params.Limit = 10
	} else if params.Limit > 100 {
		params.Limit = 100
	}

	validSortFields := map[string]bool{"created_at": true, "order_id": true}
	if params.SortBy != "" && !validSortFields[params.SortBy] {
		params.SortBy = "created_at"
	}

	if params.SortOrder != "asc" && params.SortOrder != "desc" {
		params.SortOrder = "desc"
	}

	return u.orderRepo.GetCancellationRequests(ctx, params)
}
