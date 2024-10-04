package usecase

import (
	"context"
	"log"
	"math"
	"time"

	"github.com/mohamedfawas/rmshop-clean-architecture/internal/domain"
	"github.com/mohamedfawas/rmshop-clean-architecture/internal/repository"
	"github.com/mohamedfawas/rmshop-clean-architecture/pkg/payment/razorpay"
	"github.com/mohamedfawas/rmshop-clean-architecture/pkg/utils"
)

type CheckoutUseCase interface {
	CreateOrUpdateCheckout(ctx context.Context, userID int64) (*domain.CheckoutSession, error)
	ApplyCoupon(ctx context.Context, userID int64, couponCode string) (*domain.ApplyCouponResponse, error)
	UpdateCheckoutAddress(ctx context.Context, userID, addressID int64) (*domain.CheckoutSession, error)
	GetCheckoutSummary(ctx context.Context, userID int64) (*domain.CheckoutSummary, error)
	RemoveAppliedCoupon(ctx context.Context, userID int64) (*domain.CheckoutSession, error)
}

type checkoutUseCase struct {
	checkoutRepo    repository.CheckoutRepository
	productRepo     repository.ProductRepository
	couponRepo      repository.CouponRepository
	cartRepo        repository.CartRepository
	userRepo        repository.UserRepository
	orderRepo       repository.OrderRepository
	razorpayService *razorpay.Service
}

func NewCheckoutUseCase(checkoutRepo repository.CheckoutRepository,
	productRepo repository.ProductRepository,
	cartRepo repository.CartRepository,
	couponRepo repository.CouponRepository,
	userRepo repository.UserRepository,
	orderRepo repository.OrderRepository,
	razorpayService *razorpay.Service) CheckoutUseCase {
	return &checkoutUseCase{
		checkoutRepo:    checkoutRepo,
		productRepo:     productRepo,
		couponRepo:      couponRepo,
		cartRepo:        cartRepo,
		userRepo:        userRepo,
		orderRepo:       orderRepo,
		razorpayService: razorpayService,
	}
}

func (u *checkoutUseCase) CreateOrUpdateCheckout(ctx context.Context, userID int64) (*domain.CheckoutSession, error) {
	// Get cart items
	cartItems, err := u.cartRepo.GetCartByUserID(ctx, userID)
	if err != nil {
		log.Printf("error while retrieving cart items: %v", err)
		return nil, err
	}

	if len(cartItems) == 0 {
		return nil, utils.ErrEmptyCart
	}

	// Calculate total amount and validate stock
	var totalAmount float64
	// Iterate through each cart item
	for _, item := range cartItems {
		product, err := u.productRepo.GetByID(ctx, item.ProductID)
		if err != nil {
			log.Printf("error while retrieving product details: %v", err)
			return nil, err
		}

		if product.StockQuantity < item.Quantity {
			return nil, utils.ErrInsufficientStock
		}

		totalAmount += item.Subtotal
	}

	// Get or create checkout session
	session, err := u.checkoutRepo.GetOrCreateCheckoutSession(ctx, userID)
	if err != nil {
		log.Printf("error while getting or creating checkout session: %v", err)
		return nil, err
	}

	// If the session is updated after applying the coupon, then the applied coupon will be removed
	if session.CouponApplied {
		session.CouponApplied = false
		session.CouponCode = ""
	}

	totalAmount = math.Round(totalAmount*100) / 100
	// Update the session with calculated values
	session.TotalAmount = totalAmount
	session.ItemCount = len(cartItems)
	session.FinalAmount = totalAmount // final amount is same as total amount, as coupon is not applied
	session.UpdatedAt = time.Now().UTC()
	session.DiscountAmount = 0

	// Update the checkout session in the database
	err = u.checkoutRepo.UpdateCheckoutDetails(ctx, session)
	if err != nil {
		log.Printf("error while updating checkout details: %v", err)
		return nil, err
	}

	return session, nil
}

/*
ApplyCoupon:
- Get checkout session details
- Verify it belongs to the user
- Verify that the checkout is not empty by counting checkout_items
- Check if coupon is already applied
- Get the given coupon details and validation and verification of the coupon is done
- If valid coupon, discount is applied
- Checks if discount applied is above maximum discount value
- After applying the coupon, details of checkout session is updated
*/
func (u *checkoutUseCase) ApplyCoupon(ctx context.Context, userID int64, couponCode string) (*domain.ApplyCouponResponse, error) {
	// Get or create checkout session
	checkout, err := u.checkoutRepo.GetOrCreateCheckoutSession(ctx, userID)
	if err != nil {
		log.Printf("error while getting or creating checkout session: %v", err)
		return nil, err
	}

	log.Printf("coupon already applied : %v", err)
	// Check if a coupon is already applied
	if checkout.CouponApplied {
		return nil, utils.ErrCouponAlreadyApplied
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

	// Calculate total amount from cart items
	var totalAmount float64
	for _, item := range cartItems {
		totalAmount += item.Subtotal
	}

	// Update checkout with calculated values
	checkout.TotalAmount = math.Round(totalAmount*100) / 100
	checkout.ItemCount = len(cartItems)

	// Get the coupon
	coupon, err := u.couponRepo.GetByCode(ctx, couponCode)
	if err != nil {
		if err == utils.ErrCouponNotFound {
			return nil, utils.ErrInvalidCouponCode
		}
		log.Printf("error while retrieving coupon details using given coupon code : %v", err)
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

	// Update the checkout details of checkout session
	discountAmount = math.Round(discountAmount*100) / 100 // Round to two decimal places
	checkout.DiscountAmount = discountAmount

	finalAmount := checkout.TotalAmount - discountAmount
	finalAmount = math.Round(finalAmount*100) / 100 // Round to two decimal places
	checkout.FinalAmount = finalAmount

	checkout.CouponCode = couponCode
	checkout.CouponApplied = true

	// Save the updated checkout details in checkout sessions table
	err = u.checkoutRepo.UpdateCheckoutDetails(ctx, checkout)
	if err != nil {
		log.Printf("error while updating checkout details in checkout_sessions table : %v", err)
		return nil, err
	}

	return &domain.ApplyCouponResponse{
		CheckoutSession: *checkout,
		Message:         message,
	}, nil
}

/*
UpdateCheckoutAddress:
- Get checkout session details
- Verify the checkout_status
- Get/create shipping_address details using existing user addresses.
- Update checkout_sessions table with new shipping_address_id
- Get updated checkout session details from checkout_sessions table
*/
func (u *checkoutUseCase) UpdateCheckoutAddress(ctx context.Context, userID, addressID int64) (*domain.CheckoutSession, error) {
	// Get or create checkout session
	checkout, err := u.checkoutRepo.GetOrCreateCheckoutSession(ctx, userID)
	if err != nil {
		log.Printf("error while getting or creating checkout session: %v", err)
		return nil, err
	}

	// Check if the checkout is in a valid state to update the address
	if checkout.Status != utils.CheckoutStatusPending {
		return nil, utils.ErrInvalidCheckoutState
	}

	// Get the address from the user_address table using the user address id
	address, err := u.userRepo.GetUserAddressByID(ctx, addressID)
	if err != nil {
		log.Printf("error while retrieving user address details using user address id : %v", err)
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

	// Update checkout_sessions table with new shipping address ID
	err = u.checkoutRepo.UpdateCheckoutShippingAddress(ctx, checkout.ID, shippingAddressID)
	if err != nil {
		log.Printf("error while updating checkout session details with new shipping_address_id: %v", err)
		return nil, err
	}
	// Fetch the updated checkout session details from checkout_sessions table
	updatedCheckout, err := u.checkoutRepo.GetCheckoutByID(ctx, checkout.ID)
	if err != nil {
		log.Printf("error while fetching checkout session details from checkout_sessions table: %v", err)
		return nil, err
	}

	return updatedCheckout, nil
}

/*
RemoveAppliedCoupon :
- Get checkout session details from checkout_sessions table
- Remove the applied coupon details from the retrieved sessions details
- Update the session details in the database, by removing all the applied coupon related details
*/
func (u *checkoutUseCase) RemoveAppliedCoupon(ctx context.Context, userID int64) (*domain.CheckoutSession, error) {
	// Get or create checkout session
	checkout, err := u.checkoutRepo.GetOrCreateCheckoutSession(ctx, userID)
	if err != nil {
		log.Printf("error while getting or creating checkout session: %v", err)
		return nil, err
	}

	// Check if the checkout is in a valid state to remove coupon
	if checkout.Status == utils.CheckoutStatusCompleted {
		return nil, utils.ErrCheckoutCompleted
	}

	// Check if a coupon is applied
	if !checkout.CouponApplied {
		return nil, utils.ErrNoCouponApplied
	}

	// Remove the coupon
	checkout.CouponApplied = false
	checkout.CouponCode = ""
	checkout.DiscountAmount = 0
	checkout.FinalAmount = math.Round(checkout.TotalAmount*100) / 100

	// Update the checkout in the repository
	err = u.checkoutRepo.UpdateCheckoutDetails(ctx, checkout)
	if err != nil {
		log.Printf("error while updating checkout details: %v", err)
		return nil, err
	}

	return checkout, nil
}

func (u *checkoutUseCase) GetCheckoutSummary(ctx context.Context, userID int64) (*domain.CheckoutSummary, error) {
	// Get cart items
	cartItems, err := u.cartRepo.GetCartByUserID(ctx, userID)
	if err != nil {
		log.Printf("error while retrieving cart items: %v", err)
		return nil, err
	}

	if len(cartItems) == 0 {
		return &domain.CheckoutSummary{
			UserID:         userID,
			Status:         utils.CheckoutStatusPending,
			TotalAmount:    0,
			DiscountAmount: 0,
			FinalAmount:    0,
			ItemCount:      0,
		}, nil
	}

	// Calculate totals
	var totalAmount float64
	var itemCount int
	var items []*domain.CheckoutItemDetail

	for _, cartItem := range cartItems {
		product, err := u.productRepo.GetByID(ctx, cartItem.ProductID)
		if err != nil {
			log.Printf("error retrieving product details: %v", err)
			return nil, err
		}

		totalAmount += cartItem.Subtotal

		items = append(items, &domain.CheckoutItemDetail{
			ID:        cartItem.ID,
			ProductID: cartItem.ProductID,
			Name:      product.Name,
			Quantity:  cartItem.Quantity,
			Price:     cartItem.Price,
			Subtotal:  cartItem.Subtotal,
		})
	}
	itemCount = len(cartItems)

	// Get checkout session
	checkout, err := u.checkoutRepo.GetCheckoutSession(ctx, userID)
	if err != nil {
		log.Printf("error getting or creating checkout session: %v", err)
		return nil, err
	}

	if itemCount != checkout.ItemCount {
		return nil, utils.ErrCartUpdatedAfterCreatingCheckoutSession
	}

	// Get shipping address if set
	var addressResponse *domain.ShippingAddressResponseInCheckoutSummary
	if checkout.ShippingAddressID != 0 {
		address, err := u.checkoutRepo.GetShippingAddress(ctx, checkout.ShippingAddressID)
		if err == nil {
			addressResponse = &domain.ShippingAddressResponseInCheckoutSummary{
				ID:           address.ID,
				AddressID:    address.AddressID,
				AddressLine1: address.AddressLine1,
				AddressLine2: address.AddressLine2,
				City:         address.City,
				State:        address.State,
				Landmark:     address.Landmark,
				PinCode:      address.PinCode,
				PhoneNumber:  address.PhoneNumber,
			}
		}
	}

	summary := &domain.CheckoutSummary{
		ID:             checkout.ID,
		UserID:         userID,
		TotalAmount:    checkout.TotalAmount,
		DiscountAmount: checkout.DiscountAmount,
		FinalAmount:    checkout.FinalAmount,
		ItemCount:      itemCount,
		Status:         checkout.Status,
		CouponCode:     checkout.CouponCode,
		CouponApplied:  checkout.CouponApplied,
		Address:        addressResponse,
		Items:          items,
	}

	return summary, nil
}
