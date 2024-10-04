package usecase

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/mohamedfawas/rmshop-clean-architecture/internal/domain"
	"github.com/mohamedfawas/rmshop-clean-architecture/internal/repository"
	invoicegenerator "github.com/mohamedfawas/rmshop-clean-architecture/pkg/invoice_generator"
	"github.com/mohamedfawas/rmshop-clean-architecture/pkg/payment/razorpay"
	"github.com/mohamedfawas/rmshop-clean-architecture/pkg/utils"
)

type OrderUseCase interface {
	GetOrderByID(ctx context.Context, userID, orderID int64) (*domain.Order, error)
	GetUserOrders(ctx context.Context, userID int64, page int) ([]*domain.Order, int64, error)
	GetOrders(ctx context.Context, page int) ([]*domain.Order, int64, error)
	GetPaymentByOrderID(ctx context.Context, orderID int64) (*domain.Payment, error)
	CreatePayment(ctx context.Context, tx *sql.Tx, payment *domain.Payment) error
	UpdatePayment(ctx context.Context, payment *domain.Payment) error
	ProcessPayment(ctx context.Context, tx *sql.Tx, orderID int64, paymentMethod string, amount float64) (*domain.Payment, error)
	VerifyAndUpdateRazorpayPayment(ctx context.Context, input domain.RazorpayPaymentInput) error
	PlaceOrderRazorpay(ctx context.Context, userID int64) (*domain.Order, error)
	UpdateOrderRazorpayID(ctx context.Context, orderID int64, razorpayOrderID string) error
	InitiateReturn(ctx context.Context, userID, orderID int64, reason string) (*domain.ReturnRequest, error)
	PlaceOrderCOD(ctx context.Context, userID int64) (*domain.Order, error)
	GenerateInvoice(ctx context.Context, userID, orderID int64) ([]byte, error)
	UpdateOrderDeliveryStatus(ctx context.Context, orderID int64, deliveryStatus, orderStatus, paymentStatus string) error
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

/*
GetOrderByID:
- Get order details from orders table
- Get order item details from order_items table
- Get product details (from products table) to add product name to order item data
- Get payment details from payments table
*/
func (u *orderUseCase) GetOrderByID(ctx context.Context, userID, orderID int64) (*domain.Order, error) {
	// Get order details
	order, err := u.orderRepo.GetOrderDetails(ctx, orderID)
	if err != nil {
		if err == utils.ErrOrderNotFound {
			return nil, utils.ErrOrderNotFound
		}
		log.Printf("error getting order details from orders table: %v", err)
		return nil, err
	}

	// If userID is provided (not 0), check if the order belongs to the user
	if userID != 0 && order.UserID != userID {
		return nil, utils.ErrUnauthorized
	}

	// Get order items
	items, err := u.orderRepo.GetOrderItems(ctx, orderID)
	if err != nil {
		log.Printf("error getting order items from order_items table : %v", err)
		return nil, err
	}

	// Fetch product names for each item
	for i, item := range items {
		product, err := u.productRepo.GetByID(ctx, item.ProductID)
		if err != nil {
			log.Printf("error getting product details for the order item %d: %v", item.ID, err)
			// If product not found, we'll leave the name empty
			if err != utils.ErrProductNotFound {
				return nil, err
			}
		} else {
			items[i].ProductName = product.Name
		}
	}

	// Add the fetched order item details to order struct
	order.Items = items

	// Get payment details
	payment, err := u.paymentRepo.GetByOrderID(ctx, orderID)
	if err != nil && err != utils.ErrPaymentNotFound {
		log.Printf("error getting payment details: %v", err)
		return nil, err
	}

	// Add the fetched payment details to order struct
	order.Payment = payment

	return order, nil
}

/*
GetUserOrders:
- Get order history of the given user
- Details from the orders table are retrieved
*/
func (u *orderUseCase) GetUserOrders(ctx context.Context, userID int64, page int) ([]*domain.Order, int64, error) {
	return u.orderRepo.GetUserOrders(ctx, userID, page)
}

func (u *orderUseCase) GetOrders(ctx context.Context, page int) ([]*domain.Order, int64, error) {
	ordersPerPage := 10
	offset := (page - 1) * ordersPerPage

	return u.orderRepo.GetOrders(ctx, ordersPerPage, offset)
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

/*
PlaceOrderRazorpay:
- start the transaction
- Get checkout session details
- verify this checkout session belongs to the respective user
- verify checkout status is pending
- Get checkout items
- make sure the checkout is not empty
- Iterate through checkout items, create order items entry respectively
- make sure stock for each product is available
- make sure shipping address is provided
- create order entry
- create payment entry
- create order item entry
- update checkout status
- clear user's cart
*/
func (u *orderUseCase) PlaceOrderRazorpay(ctx context.Context, userID int64) (*domain.Order, error) {
	// Start transaction
	tx, err := u.orderRepo.BeginTx(ctx)
	if err != nil {
		log.Printf("failed to start transaction: %v", err)
		return nil, err
	}
	defer tx.Rollback()

	// Get checkout session details
	checkout, err := u.checkoutRepo.GetCheckoutSession(ctx, userID)
	if err != nil {
		if err == utils.ErrCheckoutNotFound {
			return nil, utils.ErrCheckoutNotFound
		}
		log.Printf("error while retrieving checkout session details: %v", err)
		return nil, err
	}

	// Get cart items
	cartItems, err := u.cartRepo.GetCartByUserID(ctx, userID)
	if err != nil {
		log.Printf("error while retrieving cart items: %v", err)
		return nil, err
	}

	if len(cartItems) == 0 {
		return nil, utils.ErrEmptyCart
	}

	// Validate stock and calculate total
	var totalAmount float64
	for _, item := range cartItems {
		product, err := u.productRepo.GetByID(ctx, item.ProductID)
		if err != nil {
			log.Printf("error while retrieving product details: %v", err)
			return nil, err
		}
		if product.StockQuantity < item.Quantity {
			return nil, utils.ErrInsufficientStock
		}
		totalAmount += float64(item.Quantity) * product.Price
	}

	// make sure proper shipping address is provided
	if checkout.ShippingAddressID == 0 {
		return nil, utils.ErrInvalidAddress
	}

	// Create order entry
	now := time.Now().UTC()
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

	// Create the respective order entry in the database
	orderID, err := u.orderRepo.CreateOrder(ctx, tx, order)
	if err != nil {
		log.Printf("failed to create the order entry in the database : %v", err)
		return nil, err
	}
	order.ID = orderID

	// Create payment entry
	payment := &domain.Payment{
		OrderID:       order.ID,
		Amount:        checkout.FinalAmount,
		PaymentMethod: utils.PaymentMethodRazorpay,
		Status:        utils.PaymentStatusPending,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	// Create the respective payment record in the database
	err = u.orderRepo.CreatePayment(ctx, tx, payment)
	if err != nil {
		log.Printf("failed to create the respective payment record in database : %v", err)
		return nil, err
	}

	// Create order items and update stock
	for _, item := range cartItems {
		orderItem := &domain.OrderItem{
			OrderID:   order.ID,
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
			Price:     item.Price,
		}
		err = u.orderRepo.AddOrderItem(ctx, tx, orderItem)
		if err != nil {
			log.Printf("failed to add order item: %v", err)
			return nil, err
		}

		err = u.productRepo.UpdateStockTx(ctx, tx, item.ProductID, -item.Quantity)
		if err != nil {
			log.Printf("failed to update product stock: %v", err)
			return nil, err
		}
	}

	// Mark the checkout as deleted
	err = u.checkoutRepo.MarkCheckoutAsDeleted(ctx, tx, checkout.ID)
	if err != nil {
		log.Printf("error while marking checkout as deleted: %v", err)
		return nil, err
	}

	// Clear the cart of the respective user
	err = u.cartRepo.ClearCart(ctx, userID)
	if err != nil {
		log.Printf("failed to clear user's cart: %v", err)
		return nil, err
	}

	// Commit the transaction
	if err = tx.Commit(); err != nil {
		log.Printf("failed to commit transaction: %v", err)
		return nil, err
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

/*
PlaceOrderCOD:
- Start the transaction
- Get checkout session details
- Validate status of the checkout session
- Validate the final amount (should not exceed max cod limit)
- Get checkout items
- If there are no checkout items then you can't place the order
- Verify the stock is available for all the checkout items
- Create order entry in orders table
- Create order_items entry (make sure the stock is getting updated for each order item)
- Create payment entry in payments table
- Update checkout status to completed
- Clear the cart after checkout status is completed
- Commit the transaction
*/
func (u *orderUseCase) PlaceOrderCOD(ctx context.Context, userID int64) (*domain.Order, error) {
	// Start a database transaction
	tx, err := u.orderRepo.BeginTx(ctx)
	if err != nil {
		log.Printf("failed to start database transaction in PlaceOrderCOD method : %v", err)
		return nil, err
	}
	defer tx.Rollback()

	// Get the checkout session details using user ID
	checkout, err := u.checkoutRepo.GetCheckoutSession(ctx, userID)
	if err != nil {
		if err == utils.ErrCheckoutNotFound {
			return nil, utils.ErrCheckoutNotFound
		}
		log.Printf("error while retrieving checkout details using user ID : %v", err)
		return nil, err
	}

	// Check if the checkout is in a valid state to place an order
	if checkout.Status != utils.CheckoutStatusPending {
		return nil, utils.ErrOrderAlreadyPlaced
	}

	// Check if the order total exceeds the COD limit
	if checkout.FinalAmount > utils.CODLimit {
		return nil, utils.ErrCODLimitExceeded
	}

	// Get cart items
	cartItems, err := u.cartRepo.GetCartByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// If cart is empty, then you can't place the order
	if len(cartItems) == 0 {
		return nil, utils.ErrEmptyCart
	}

	// Verify that all items have sufficient stock
	for _, item := range cartItems {
		product, err := u.productRepo.GetByID(ctx, item.ProductID)
		if err != nil {
			log.Printf("error while retrieving product details to check stock availability : %v", err)
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
		log.Printf("error while creating the order entry in the database : %v", err)
		return nil, err
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
	// Add the payment record in the database
	err = u.orderRepo.CreatePayment(ctx, tx, payment)
	if err != nil {
		log.Printf("failed to create payment record in payments table : %v", err)
		return nil, err
	}

	// Create order items and update product stock
	for _, item := range cartItems {
		orderItem := &domain.OrderItem{
			OrderID:   order.ID,
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
			Price:     item.Price,
		}
		err = u.orderRepo.AddOrderItem(ctx, tx, orderItem)
		if err != nil {
			log.Printf("error while adding order item entry in order_items table : %v", err)
			return nil, err
		}

		err = u.productRepo.UpdateStockTx(ctx, tx, item.ProductID, -item.Quantity)
		if err != nil {
			log.Printf("failed to update product stock: %v", err)
			return nil, err
		}
	}

	// Mark the checkout as deleted
	err = u.checkoutRepo.MarkCheckoutAsDeleted(ctx, tx, checkout.ID)
	if err != nil {
		log.Printf("error while marking checkout as deleted : %v", err)
		return nil, err
	}

	// Clear the user's cart
	err = u.cartRepo.ClearCart(ctx, userID)
	if err != nil {
		log.Printf("error while clearing user's cart : %v", err)
		return nil, err
	}

	// Commit the transaction
	if err = tx.Commit(); err != nil {
		log.Printf("error while commiting transaction in PlaceOrderCOD method : %v", err)
		return nil, err
	}

	return order, nil
}

/*
getOrderWithItems:
- Used for retrieving specific details for generating order invoice
- Get order details
- Get order items
- Get product details (mainly product name of each order item)
- Get shipping address
*/
func (u *orderUseCase) getOrderWithItems(ctx context.Context, userID, orderID int64) (*domain.Order, error) {
	// Get  order details from orders table
	order, err := u.orderRepo.GetOrderDetails(ctx, orderID)
	if err != nil {
		log.Printf("failed to get order details from orders table : %v", err)
		return nil, err
	}

	// Check if the order belongs to the user
	if order.UserID != userID {
		return nil, utils.ErrUnauthorized
	}

	// Get order items
	items, err := u.orderRepo.GetOrderItems(ctx, orderID)
	if err != nil {
		log.Printf("failed to get order items details from order_items table : %v", err)
		return nil, err
	}

	// Fetch product details for each order item
	for i, item := range items {
		product, err := u.productRepo.GetByID(ctx, item.ProductID)
		if err != nil {
			log.Printf("failed to get product details for item %d: %w", item.ID, err)
			return nil, err
		}
		items[i].ProductName = product.Name
	}
	order.Items = items

	// Get shipping address
	shippingAddress, err := u.orderRepo.GetShippingAddress(ctx, order.ShippingAddressID)
	if err != nil {
		log.Printf("failed to get shipping address details from shipping_addresses table: %v", err)
		return nil, err
	}
	order.ShippingAddress = shippingAddress

	return order, nil
}

/*
GenerateInvoice:
- Get all the details fetched by "getOrderWithItems" method
*/
func (u *orderUseCase) GenerateInvoice(ctx context.Context, userID, orderID int64) ([]byte, error) {
	order, err := u.getOrderWithItems(ctx, userID, orderID)
	if err != nil {
		log.Printf("error while executing getOrderWithItems method : %v", err)
		return nil, err
	}

	// For a cancelled order, no need to generate invoice
	if order.OrderStatus == utils.OrderStatusCancelled {
		return nil, utils.ErrCancelledOrder
	}

	// If order status is pending payment, no need to generate invoice
	if order.OrderStatus == utils.OrderStatusPending {
		return nil, utils.ErrUnpaidOrder
	}

	// If order status is cancelled , no need to generate invoice
	if order.OrderStatus == utils.OrderStatusCancelled && order.OrderStatus == utils.OrderStatusPendingCancellation {
		return nil, utils.ErrOrderCancelled
	}

	// No items in orders
	if len(order.Items) == 0 {
		return nil, utils.ErrEmptyOrder
	}

	// Generate PDF using the utility function
	pdfBytes, err := invoicegenerator.GenerateInvoicePDF(order)
	if err != nil {
		log.Printf("failed to generate invoice PDF: %v", err)
		return nil, err
	}

	return pdfBytes, nil
}

/*
UpdateOrderDeliveryStatus :
- Validates the order status and delivery status given as input
- Verifies if the given order is already delivered before
- If not delivered, record the current time as delivered_at variable
- Get payment details of the given order
- If the payment method used for the order is cod, then the payment_status input will be validated
- Update the payment details for cod orders
- Update order status and delivery status for the given order (transaction method, bcz we are updating many values)
- Make respective changes in db (orders table)
*/
func (u *orderUseCase) UpdateOrderDeliveryStatus(ctx context.Context, orderID int64, deliveryStatus, orderStatus, paymentStatus string) error {
	// Validate delivery status
	if deliveryStatus != "" && !isValidDeliveryStatus(deliveryStatus) {
		return utils.ErrInvalidDeliveryStatus
	}

	// Validate order status
	if orderStatus != "" && !isValidOrderStatus(orderStatus) {
		return utils.ErrInvalidOrderStatus
	}

	// Check if the order exists and is not already delivered
	isDelivered, err := u.orderRepo.IsOrderDelivered(ctx, orderID)
	if err != nil {
		log.Printf("error while checking if the given order is delivered : %v", err)
		return err
	}

	// If order is already delivered
	if isDelivered {
		return utils.ErrOrderAlreadyDelivered
	}

	// Get payment details
	payment, err := u.paymentRepo.GetByOrderID(ctx, orderID)
	if err != nil {
		log.Printf("error while fetching payment details of the given order id : %v", err)
		return err
	}

	// For COD orders, check if payment status is provided and valid
	if payment.PaymentMethod == utils.PaymentMethodCOD {
		if paymentStatus == "" {
			return utils.ErrMissingPaymentStatus
		}
		// If payment status is not paid, then no need to record the delivery
		if paymentStatus != utils.PaymentStatusPaid {
			return utils.ErrInvalidPaymentStatus
		}
	}

	// Record delivery time
	var deliveredAt *time.Time
	if deliveryStatus == utils.DeliveryStatusDelivered {
		now := time.Now().UTC()
		deliveredAt = &now
		orderStatus = utils.OrderStatusCompleted
		// Update payment status for COD orders
		if payment.PaymentMethod == utils.PaymentMethodCOD {
			// update payment details
			payment.Status = utils.PaymentStatusPaid
			payment.UpdatedAt = time.Now().UTC()
			err = u.paymentRepo.UpdatePayment(ctx, payment)
			if err != nil {
				log.Printf("error while updating payment details : %v", err)
				return err
			}
		}
	}

	// Start a transaction
	tx, err := u.orderRepo.BeginTx(ctx)
	if err != nil {
		log.Printf("error while starting transaction in UpdateOrderDeliveryStatus method : %v", err)
		return err
	}
	defer tx.Rollback()

	// Update order status
	err = u.orderRepo.UpdateOrderDeliveryStatus(ctx, tx, orderID, deliveryStatus, orderStatus, deliveredAt)
	if err != nil {
		log.Printf("error while updating order delivery status in orders table : %v", err)
		return err
	}

	// Commit the transaction
	err = tx.Commit()
	if err != nil {
		log.Printf("error while commiting transaction in UpdateOrderDeliveryStatus method : %v", err)
	}

	return err
}

/*
isValidDeliveryStatus:
- Evaluates the delivery status given as input
*/
func isValidDeliveryStatus(status string) bool {
	// Define the valid delivery statuses
	validStatuses := []string{utils.DeliveryStatusPending,
		utils.DeliveryStatusInTransit,
		utils.DeliveryStatusOutForDelivery,
		utils.DeliveryStatusDelivered,
		utils.DeliveryStatusFailedDeliveryAttempt,
		utils.DeliveryStatusReturnedToSender}
	// validate the input status
	for _, s := range validStatuses {
		if status == s {
			return true
		}
	}
	return false
}

/*
isValidOrderStatus:
- Evaluates the order status given as input
*/
func isValidOrderStatus(status string) bool {
	// Define valid order statuses
	validStatuses := []string{utils.OrderStatusPending,
		utils.OrderStatusConfirmed,
		utils.OrderStatusProcessing,
		utils.OrderStatusCompleted,
		utils.OrderStatusShipped,
		utils.OrderStatusPendingCancellation,
		utils.OrderStatusCancelled,
		utils.OrderStatusRefunded}
	// Evaluate the input order status
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

/*
CancelOrder:
- Begins transaction
- Get order details using order id
- validates the order related parameters
- create cancellation request entry
- Add cancellation request entry to cancellation_requests table
- Check if order is eligible for immediate cancellation (unpaid orders, cod)
  - If eligible,
  - Update order status, is_cancelled
  - Update stock for the products in the order
  - Update cancellation requests table
  - If not eligible
  - Update order status

- Commits transaction
*/
func (u *orderUseCase) CancelOrder(ctx context.Context, userID, orderID int64) (*domain.OrderCancellationResult, error) {
	// Begin transaction
	tx, err := u.orderRepo.BeginTx(ctx)
	if err != nil {
		log.Printf("failed to begin transaction: %v", err)
		return nil, err
	}
	defer tx.Rollback()

	// get the order details for the given order id from the orders table
	order, err := u.orderRepo.GetByIDTx(ctx, tx, orderID)
	if err != nil {
		log.Printf("error fetching order details: %v", err)
		return nil, err
	}

	// ensure that the given order belongs to the authenticated user
	if order.UserID != userID {
		return nil, utils.ErrUnauthorized
	}

	// ensure the order is not cancelled before
	if order.IsCancelled {
		return nil, utils.ErrOrderAlreadyCancelled
	}

	// Ensure the order is in a cancellable status
	if !isOrderCancellable(order.OrderStatus) {
		return nil, utils.ErrOrderNotCancellable
	}

	// Create the result to show at in the api response
	result := &domain.OrderCancellationResult{
		OrderID: orderID,
	}

	// Create cancellation request entry
	cancellationRequest := &domain.CancellationRequest{
		OrderID:                   orderID,
		UserID:                    userID,
		CreatedAt:                 time.Now().UTC(),
		CancellationRequestStatus: utils.CancellationStatusPendingReview,
		IsStockUpdated:            false,
	}

	// Create cancellation request entry in the cancellation_requests table
	err = u.orderRepo.CreateCancellationRequestTx(ctx, tx, cancellationRequest)
	if err != nil {
		log.Printf("failed to create cancellation request: %v", err)
		return nil, err
	}

	// If the order is eligible for immediate cancellation, meaning cod orders for which payment is pending
	if order.OrderStatus == utils.OrderStatusPending {
		// Immediate cancellation for unpaid orders
		// Update order_status in orders table and update is_cancelled column value
		err = u.orderRepo.UpdateOrderStatusAndSetCancelledTx(ctx, tx, orderID, utils.OrderStatusCancelled, utils.DeliveryStatusReturnedToSender, true)
		if err != nil {
			log.Printf("failed to update order status and is_cancelled in orders table : %v", err)
			return nil, err
		}

		// Immediate cancellation means we need to update stock quantity respective to cancelled order
		err = u.updateStockForCancelledOrder(ctx, tx, orderID)
		if err != nil {
			log.Printf("failed to update stock: %v", err)
			return nil, err
		}

		// change bool value after updating stock in products table
		cancellationRequest.IsStockUpdated = true
		cancellationRequest.CancellationRequestStatus = utils.CancellationStatusCancelled

		// now update cancellation_requests table
		err = u.orderRepo.UpdateCancellationRequestTx(ctx, tx, cancellationRequest)
		if err != nil {
			log.Printf("failed to update cancellation request: %v", err)
			return nil, err
		}

		// Now update the result for api response
		result.OrderStatus = utils.OrderStatusCancelled
		result.RequiresAdminReview = false
	} else {
		// Set to pending cancellation for paid orders
		err = u.orderRepo.UpdateOrderStatusTx(ctx, tx, orderID, utils.OrderStatusPendingCancellation)
		if err != nil {
			return nil, fmt.Errorf("failed to update order status: %w", err)
		}

		// Now update the result for api response
		result.OrderStatus = utils.OrderStatusPendingCancellation
		result.RequiresAdminReview = true
	}

	// Commit the transaction
	if err = tx.Commit(); err != nil {
		log.Printf("failed to commit transaction: %v", err)
		return nil, err
	}

	return result, nil
}

/*
updateStockForCancelledOrder:
- Get order items
- Update stock quantity in the products table for respective order items
*/
func (u *orderUseCase) updateStockForCancelledOrder(ctx context.Context, tx *sql.Tx, orderID int64) error {
	// Get order items from order_items table
	orderItems, err := u.orderRepo.GetOrderItemsTx(ctx, tx, orderID)
	if err != nil {
		log.Printf("failed to get order items: %v", err)
		return err
	}

	for _, item := range orderItems {
		err = u.productRepo.UpdateStockTx(ctx, tx, item.ProductID, item.Quantity)
		if err != nil {
			log.Printf("failed to update stock for product %d: %v", item.ProductID, err)
			return err
		}
	}

	return nil
}

/*
ApproveCancellation:
- Starts the transaction
- Get order using order id
- validate order related values
- Get cancellation request using order id
- Update order_status in orders table
- Check if refund is needed
  - If refund applicable, then get or create wallet and make refund to the wallet
  - Also, make entry in wallet_transactions table and wallets table respectively

- Update stock_quantity of the products which are part of the order items in this cancelled order
- Update cancellation related details in cancellation_requests table
*/
func (u *orderUseCase) ApproveCancellation(ctx context.Context, orderID int64) (*domain.OrderStatusUpdateResult, error) {
	// Start a database transaction
	tx, err := u.orderRepo.BeginTx(ctx)
	if err != nil {
		log.Printf("failed to start transaction: %v", err)
		return nil, err
	}
	defer tx.Rollback()

	// Get the order details from orders table
	order, err := u.orderRepo.GetByIDTx(ctx, tx, orderID)
	if err != nil {
		if err == utils.ErrOrderNotFound {
			return nil, utils.ErrOrderNotFound
		}
		log.Printf("failed to get order details from orders table : %v", err)
		return nil, err
	}

	// Check if the order is in pending cancellation state
	if order.OrderStatus != utils.OrderStatusPendingCancellation {
		return nil, utils.ErrOrderNotPendingCancellation
	}

	// Get the cancellation request
	cancellationRequest, err := u.orderRepo.GetCancellationRequestByOrderIDTx(ctx, tx, orderID)
	if err != nil {
		log.Printf("failed to get cancellation request using order id: %v", err)
		return nil, err
	}

	// Update order status to cancelled
	err = u.orderRepo.UpdateOrderStatusAndSetCancelledTx(ctx, tx, orderID, utils.OrderStatusCancelled, utils.DeliveryStatusReturnedToSender, true)
	if err != nil {
		log.Printf("failed to update order_status and is_cancelled columns in orders table : %v", err)
		return nil, err
	}

	result := &domain.OrderStatusUpdateResult{
		OrderID:            orderID,
		UpdatedOrderStatus: utils.OrderStatusCancelled,
		RefundStatus:       utils.RefundStatusNotApplicable,
	}

	// Check if refund is needed
	payment, err := u.paymentRepo.GetByOrderIDTx(ctx, tx, orderID)
	if err != nil {
		if err != utils.ErrPaymentNotFound {
			log.Printf("failed to get payment details using order id : %v", err)
			return nil, err
		}
		// If payment not found, we assume no refund is needed
		payment = nil
	}

	// If payment is found , then it is evaluated
	if payment != nil && payment.Status == utils.PaymentStatusPaid && payment.PaymentMethod == utils.PaymentMethodRazorpay {
		// Process refund
		err = u.processRefund(ctx, tx, order, payment)
		if err != nil {
			log.Printf("failed to process refund: %v", err)
			return nil, err
		}
		result.RefundStatus = utils.RefundStatusInitiated
	}

	// Update stock quantity of products which are part of the order items in cancelled order
	err = u.updateStockForCancelledOrder(ctx, tx, orderID)
	if err != nil {
		log.Printf("failed to update stock quantity : %v", err)
		return nil, err
	}

	// Update cancellation request variables
	cancellationRequest.CancellationRequestStatus = utils.CancellationStatusCancelled
	cancellationRequest.IsStockUpdated = true

	// Update cancellation_requests table with new status and is_stock_updated value
	err = u.orderRepo.UpdateCancellationRequestTx(ctx, tx, cancellationRequest)
	if err != nil {
		log.Printf("failed to update cancellation request: %v", err)
		return nil, err
	}

	// Commit the transaction
	if err = tx.Commit(); err != nil {
		log.Printf("failed to commit transaction: %v", err)
		return nil, err
	}

	return result, nil
}

/*
processRefund:
- Part of transaction in the methods this is used
- First try to get the wallet of the given user
  - If the wallet is not available, a wallet is created for the user

- Create a wallet transaction entry in wallet_transactions table
- Update wallet balance in wallets table for the given user id
- Now update the payment status in the payments table
*/
func (u *orderUseCase) processRefund(ctx context.Context, tx *sql.Tx, order *domain.Order, payment *domain.Payment) error {

	// Get the user's wallet
	wallet, err := u.walletRepo.GetWalletTx(ctx, tx, order.UserID)
	if err != nil {
		if err == sql.ErrNoRows {
			// If wallet doesn't exist, create one
			now := time.Now().UTC()
			wallet = &domain.Wallet{
				UserID:    order.UserID,
				Balance:   0,
				CreatedAt: now,
				UpdatedAt: now,
			}
			err = u.walletRepo.CreateWalletTx(ctx, tx, wallet)
			if err != nil {
				log.Printf("failed to create wallet: %v", err)
				return err
			}
		} else {
			log.Printf("failed to get wallet: %v", err)
			return err
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
		ReferenceType:   referenceType,
		BalanceAfter:    newBalance,
		CreatedAt:       time.Now().UTC(),
	}

	// Create wallet transaction entry in wallet_transactions table
	err = u.walletRepo.CreateWalletTransactionTx(ctx, tx, walletTx)
	if err != nil {
		log.Printf("failed to create wallet transaction: %v", err)
		return err
	}

	// Update user's wallet balance respectively for the given user id
	err = u.walletRepo.UpdateWalletBalanceTx(ctx, tx, order.UserID, newBalance)
	if err != nil {
		log.Printf("failed to update wallet balance: %v", err)
		return err
	}

	// Update payment status
	err = u.paymentRepo.UpdateStatusTx(ctx, tx, payment.ID, utils.PaymentStatusRefunded)
	if err != nil {
		log.Printf("failed to update payment status: %v", err)
		return err
	}

	return nil
}

/*
AdminCancelOrder:
- Start the transaction
- Get order details from orders table using order id
- Validate the order status
- Create cancellation request entry in the cancellation_requests table
- Create the response entry for api response
- Get payment details using order id
  - if eligible for refund, process the refund

- Update stock quantity of the product which are part of cancelled order items
- Update cancellation request table (is_stock_updated)
- Update order status, delivery status, is_cancelled
*/
func (u *orderUseCase) AdminCancelOrder(ctx context.Context, orderID int64) (*domain.AdminOrderCancellationResult, error) {
	// Start a database transaction
	tx, err := u.orderRepo.BeginTx(ctx)
	if err != nil {
		log.Printf("failed to start transaction: %v", err)
		return nil, err
	}
	defer tx.Rollback()

	// Get the order details from orders table
	order, err := u.orderRepo.GetByIDTx(ctx, tx, orderID)
	if err != nil {
		if err == utils.ErrOrderNotFound {
			return nil, utils.ErrOrderNotFound
		}
		log.Printf("failed to get order details from orders table : %v", err)
		return nil, err
	}

	// Check if the order is cancellable
	if order.OrderStatus == utils.OrderStatusCancelled {
		return nil, utils.ErrOrderAlreadyCancelled
	}
	if !isOrderCancellable(order.OrderStatus) {
		return nil, utils.ErrOrderNotCancellable
	}

	// Create cancellation request entry
	cancellationRequest := &domain.CancellationRequest{
		OrderID:                   orderID,
		UserID:                    order.UserID,
		CreatedAt:                 time.Now().UTC(),
		CancellationRequestStatus: utils.CancellationStatusCancelled,
		IsStockUpdated:            false, // update this after updating the stock
	}

	// Create cancellation request entry in the database
	err = u.orderRepo.CreateCancellationRequestTx(ctx, tx, cancellationRequest)
	if err != nil {
		log.Printf("failed to create cancellation request: %v", err)
		return nil, err
	}

	// Create the response to show as api response
	result := &domain.AdminOrderCancellationResult{
		OrderID:             orderID,
		OrderStatus:         utils.OrderStatusCancelled,
		RequiresAdminReview: false,
		RefundInitiated:     false,
	}

	// Check if refund is needed
	// First get payment details using order id
	payment, err := u.paymentRepo.GetByOrderIDTx(ctx, tx, orderID)
	if err != nil {
		if err != utils.ErrPaymentNotFound {
			log.Printf("failed to get payment details using order id : %v", err)
			return nil, err
		}
		// If payment not found, we assume no refund is needed
		payment = nil
	}

	if payment != nil && payment.Status == utils.PaymentStatusPaid && payment.PaymentMethod == utils.PaymentMethodRazorpay {
		// Process refund
		err = u.processRefund(ctx, tx, order, payment)
		if err != nil {
			log.Printf("failed to process refund: %v", err)
			return nil, err
		}
		// Update in result we show in api response
		result.RefundInitiated = true
	}

	// Update stock quantities of the products which are part of the order items
	err = u.updateStockForCancelledOrder(ctx, tx, orderID)
	if err != nil {
		log.Printf("failed to update stock quantity for products which are part of cancelled order items: %v", err)
		return nil, err
	}

	// Update cancellation request to mark stock as updated
	cancellationRequest.IsStockUpdated = true
	err = u.orderRepo.UpdateCancellationRequestTx(ctx, tx, cancellationRequest)
	if err != nil {
		log.Printf("failed to update cancellation request: %v", err)
		return nil, err
	}

	// Update order status to cancelled and set is_cancelled to true, update delivery status also
	err = u.orderRepo.UpdateOrderStatusAndSetCancelledTx(ctx, tx, orderID, utils.OrderStatusCancelled, utils.DeliveryStatusReturnedToSender, true)
	if err != nil {
		log.Printf("failed to update order status and set cancelled: %v", err)
		return nil, err
	}

	// Commit the transaction
	if err = tx.Commit(); err != nil {
		log.Printf("failed to commit transaction: %v", err)
		return nil, err
	}

	return result, nil
}

/*
isOrderCancellable:
- used to validate whether the given order status is eligible for cancellation
*/
func isOrderCancellable(status string) bool {
	cancellableStatuses := []string{utils.OrderStatusPending,
		utils.OrderStatusConfirmed,
		utils.OrderStatusProcessing}
	for _, s := range cancellableStatuses {
		if status == s {
			return true
		}
	}
	return false
}

func (u *orderUseCase) GetCancellationRequests(ctx context.Context, params domain.CancellationRequestParams) ([]*domain.CancellationRequest, int64, error) {
	return u.orderRepo.GetCancellationRequests(ctx, params)
}
