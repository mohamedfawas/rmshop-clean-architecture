package usecase

import (
	"context"
	"database/sql"
	"log"
	"time"

	"github.com/mohamedfawas/rmshop-clean-architecture/internal/domain"
	"github.com/mohamedfawas/rmshop-clean-architecture/internal/repository"
	"github.com/mohamedfawas/rmshop-clean-architecture/pkg/utils"
)

type CheckoutUseCase interface {
	CreateCheckout(ctx context.Context, userID int64) (*domain.CheckoutSession, error)
	ApplyCoupon(ctx context.Context, userID int64, checkoutID int64, couponCode string) (*domain.ApplyCouponResponse, error)
	UpdateCheckoutAddress(ctx context.Context, userID, checkoutID, addressID int64) (*domain.CheckoutSession, error)
	GetCheckoutSummary(ctx context.Context, userID, checkoutID int64) (*domain.CheckoutSummary, error)
	PlaceOrder(ctx context.Context, userID, checkoutID int64, paymentMethod string) (*domain.Order, error)
}

type checkoutUseCase struct {
	checkoutRepo repository.CheckoutRepository
	productRepo  repository.ProductRepository
	couponRepo   repository.CouponRepository
	cartRepo     repository.CartRepository
	userRepo     repository.UserRepository
	orderRepo    repository.OrderRepository
}

func NewCheckoutUseCase(checkoutRepo repository.CheckoutRepository, productRepo repository.ProductRepository, cartRepo repository.CartRepository, couponRepo repository.CouponRepository, userRepo repository.UserRepository, orderRepo repository.OrderRepository) CheckoutUseCase {
	return &checkoutUseCase{
		checkoutRepo: checkoutRepo,
		productRepo:  productRepo,
		couponRepo:   couponRepo,
		cartRepo:     cartRepo,
		userRepo:     userRepo,
		orderRepo:    orderRepo,
	}
}

func (u *checkoutUseCase) CreateCheckout(ctx context.Context, userID int64) (*domain.CheckoutSession, error) {
	// Get cart items
	cartItems, err := u.checkoutRepo.GetCartItems(ctx, userID)
	if err != nil {
		log.Printf("error while retrieving cart items : %v", err)
		return nil, err
	}

	if len(cartItems) == 0 {
		return nil, utils.ErrEmptyCart
	}

	// Validate items and calculate total
	var totalAmount float64
	var checkoutItems []*domain.CheckoutItem

	for _, item := range cartItems {
		product, err := u.productRepo.GetByID(ctx, item.ProductID)
		if err != nil {
			log.Printf("error while retrieving product details : %v", err)
			return nil, err
		}

		if product.StockQuantity < item.Quantity {
			return nil, utils.ErrInsufficientStock
		}

		subtotal := float64(item.Quantity) * product.Price
		checkoutItems = append(checkoutItems, &domain.CheckoutItem{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
			Price:     product.Price,
			Subtotal:  subtotal,
		})

		totalAmount += subtotal
	}

	// Create checkout session
	session, err := u.checkoutRepo.CreateCheckoutSession(ctx, userID)
	if err != nil {
		log.Printf("error while creating checkout session : %v", err)
		return nil, err
	}

	// Add items to checkout session
	err = u.checkoutRepo.AddCheckoutItems(ctx, session.ID, checkoutItems)
	if err != nil {
		log.Printf("error while adding items to checkout session : %v", err)
		return nil, err
	}

	// Update the session with calculated values
	session.TotalAmount = totalAmount
	session.ItemCount = len(checkoutItems)
	session.FinalAmount = totalAmount

	// Update the checkout session in the database
	err = u.checkoutRepo.UpdateCheckoutDetails(ctx, session)
	if err != nil {
		log.Printf("error while updating checkout details : %v", err)
		return nil, err
	}

	return session, nil
}

func (u *checkoutUseCase) ApplyCoupon(ctx context.Context, userID int64, checkoutID int64, couponCode string) (*domain.ApplyCouponResponse, error) {
	// Get the checkout session
	checkout, err := u.checkoutRepo.GetCheckoutByID(ctx, checkoutID)
	if err != nil {
		log.Printf("error : get checkout by id : %v", err)
		return nil, err
	}

	// Verify the checkout belongs to the user
	if checkout.UserID != userID {
		return nil, utils.ErrUnauthorized
	}

	// Double-check if the checkout is empty by counting items
	items, err := u.checkoutRepo.GetCheckoutItems(ctx, checkoutID)
	if err != nil {
		log.Printf("error : Get checkout items : %v", err)
		return nil, err
	}

	if len(items) == 0 {
		return nil, utils.ErrEmptyCheckout
	}

	// Update ItemCount if it's inconsistent
	if checkout.ItemCount != len(items) {
		checkout.ItemCount = len(items)
	}

	// Check if a coupon is already applied
	if checkout.CouponApplied {
		return nil, utils.ErrCouponAlreadyApplied
	}

	// Get the coupon
	coupon, err := u.couponRepo.GetByCode(ctx, couponCode)
	if err != nil {
		if err == utils.ErrCouponNotFound {
			return nil, utils.ErrInvalidCouponCode
		}
		log.Printf("error : Get coupon using code : %v", err)
		return nil, err
	}

	// Check if the coupon is active
	if !coupon.IsActive {
		return nil, utils.ErrCouponInactive
	}

	// Check if the coupon has expired
	if coupon.ExpiresAt != nil && coupon.ExpiresAt.Before(time.Now()) {
		return nil, utils.ErrCouponExpired
	}

	// Check if the minimum order amount is met
	if checkout.TotalAmount < coupon.MinOrderAmount {
		return nil, utils.ErrOrderTotalBelowMinimum
	}

	// Calculate the discount
	discountAmount := checkout.TotalAmount * (coupon.DiscountPercentage / 100)

	// Apply maximum discount cap if necessary
	message := ""
	if discountAmount > utils.MaxDiscountAmount {
		discountAmount = utils.MaxDiscountAmount
		message = "Maximum discount cap applied"
	}

	// Update the checkout
	checkout.DiscountAmount = discountAmount
	checkout.FinalAmount = checkout.TotalAmount - discountAmount
	checkout.CouponCode = couponCode
	checkout.CouponApplied = true

	// Save the updated checkout
	err = u.checkoutRepo.UpdateCheckoutDetails(ctx, checkout)
	if err != nil {
		log.Printf("error : Update checkout details : %v", err)
		return nil, err
	}

	return &domain.ApplyCouponResponse{
		CheckoutSession: *checkout,
		Message:         message,
	}, nil
}

func (u *checkoutUseCase) UpdateCheckoutAddress(ctx context.Context, userID, checkoutID, addressID int64) (*domain.CheckoutSession, error) {
	// Get the checkout session
	checkout, err := u.checkoutRepo.GetCheckoutByID(ctx, checkoutID)
	if err != nil {
		return nil, err
	}

	// Check if the checkout belongs to the user
	if checkout.UserID != userID {
		return nil, utils.ErrUnauthorized
	}

	// Check if the checkout is in a valid state to update the address
	if checkout.Status != "pending" {
		return nil, utils.ErrInvalidCheckoutState
	}

	// Get the address
	address, err := u.userRepo.GetUserAddressByID(ctx, addressID)
	if err != nil {
		return nil, err
	}

	// Check if the address belongs to the user
	if address.UserID != userID {
		return nil, utils.ErrAddressNotBelongToUser
	}

	// Create or get existing shipping address
	shippingAddressID, err := u.checkoutRepo.CreateOrGetShippingAddress(ctx, userID, addressID)
	if err != nil {
		return nil, err
	}

	// Update checkout with new shipping address ID
	err = u.checkoutRepo.UpdateCheckoutShippingAddress(ctx, checkoutID, shippingAddressID)
	if err != nil {
		return nil, err
	}

	// Fetch the updated checkout session
	updatedCheckout, err := u.checkoutRepo.GetCheckoutByID(ctx, checkoutID)
	if err != nil {
		return nil, err
	}

	return updatedCheckout, nil
}

func validateAddress(address *domain.UserAddress) error {
	if address.AddressLine1 == "" || address.City == "" || address.State == "" || address.PinCode == "" || address.PhoneNumber == "" {
		return utils.ErrMissingRequiredFields
	}
	// Add more validation as needed
	return nil
}

func (u *checkoutUseCase) GetCheckoutSummary(ctx context.Context, userID, checkoutID int64) (*domain.CheckoutSummary, error) {
	summary, err := u.checkoutRepo.GetCheckoutWithItems(ctx, checkoutID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, utils.ErrCheckoutNotFound
		}
		return nil, err
	}

	if summary.UserID != userID {
		return nil, utils.ErrUnauthorized
	}

	return summary, nil
}

func (u *checkoutUseCase) PlaceOrder(ctx context.Context, userID, checkoutID int64, paymentMethod string) (*domain.Order, error) {
	// Get the checkout session
	checkout, err := u.checkoutRepo.GetCheckoutByID(ctx, checkoutID)
	if err != nil {
		log.Printf("Error getting checkout: %v", err)
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

	// Get checkout items
	items, err := u.checkoutRepo.GetCheckoutItems(ctx, checkoutID)
	if err != nil {
		log.Printf("Error getting checkout items: %v", err)
		return nil, err
	}

	// Log the state for debugging
	log.Printf("Placing order for checkout %d: Items count: %d, Checkout item count: %d",
		checkoutID, len(items), checkout.ItemCount)

	if len(items) == 0 {
		return nil, utils.ErrEmptyCart
	}

	// Verify that all items have sufficient stock
	for _, item := range items {
		product, err := u.productRepo.GetByID(ctx, item.ProductID)
		if err != nil {
			log.Printf("Error getting product %d: %v", item.ProductID, err)
			return nil, err
		}
		if product.StockQuantity < item.Quantity {
			return nil, utils.ErrInsufficientStock
		}
	}

	// Verify that a valid address is associated with the checkout
	if checkout.ShippingAddress.ID == 0 {
		return nil, utils.ErrInvalidAddress
	}

	// Start a database transaction
	tx, err := u.checkoutRepo.BeginTx(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// Create the order
	order := &domain.Order{
		UserID:            userID,
		TotalAmount:       checkout.FinalAmount,
		OrderStatus:       "pending",
		DeliveryStatus:    "processing",
		ShippingAddressID: checkout.ShippingAddress.ID,
	}

	// Create the order in the database
	orderID, err := u.orderRepo.CreateOrder(ctx, tx, order)
	if err != nil {
		log.Printf("Error creating order: %v", err)
		return nil, err
	}

	// Create the payment
	payment := &domain.Payment{
		OrderID:       orderID,
		Amount:        checkout.FinalAmount,
		PaymentMethod: paymentMethod,
		Status:        "pending",
	}

	// Add the payment to the database
	err = u.orderRepo.CreatePayment(ctx, tx, payment)
	if err != nil {
		log.Printf("Error creating payment: %v", err)
		return nil, err
	}

	// Add order items
	for _, item := range items {
		orderItem := &domain.OrderItem{
			OrderID:   order.ID,
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
			Price:     item.Price,
		}
		err = u.orderRepo.AddOrderItem(ctx, tx, orderItem)
		if err != nil {
			log.Printf("Error adding order item: %v", err)
			return nil, err
		}

		// Update product stock
		err = u.productRepo.UpdateStock(ctx, tx, item.ProductID, -item.Quantity)
		if err != nil {
			log.Printf("Error updating product stock: %v", err)
			return nil, err
		}
	}

	// Update checkout status
	checkout.Status = "completed"
	err = u.checkoutRepo.UpdateCheckoutStatus(ctx, tx, checkout)
	if err != nil {
		log.Printf("Error updating checkout status: %v", err)
		return nil, err
	}

	// Clear the user's cart
	err = u.cartRepo.ClearCart(ctx, userID)
	if err != nil {
		log.Printf("Error clearing user's cart: %v", err)
		return nil, err
	}

	// Commit the transaction
	err = tx.Commit()
	if err != nil {
		log.Printf("Error committing transaction: %v", err)
		return nil, err
	}

	return order, nil
}
