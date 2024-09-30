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
	CreateCheckout(ctx context.Context, userID int64) (*domain.CheckoutSession, error)
	ApplyCoupon(ctx context.Context, userID int64, checkoutID int64, couponCode string) (*domain.ApplyCouponResponse, error)
	UpdateCheckoutAddress(ctx context.Context, userID, checkoutID, addressID int64) (*domain.CheckoutSession, error)
	GetCheckoutSummary(ctx context.Context, userID, checkoutID int64) (*domain.CheckoutSummary, error)
	RemoveAppliedCoupon(ctx context.Context, userID, checkoutID int64) (*domain.CheckoutSession, error)
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

/*
CreateCheckout:
- Gets cart items currently in user's cart
- Iterate over each cart item entry in cart_items table and retrieves the entries
- Creates checkout_sessions entry
- Add the cart items to checkout_items and associate it to the created checkout_session
- Updates the total_amount, final_amount and discount_amount values
*/
func (u *checkoutUseCase) CreateCheckout(ctx context.Context, userID int64) (*domain.CheckoutSession, error) {
	// Get cart items
	cartItems, err := u.checkoutRepo.GetCartItems(ctx, userID)
	if err != nil {
		log.Printf("error while retrieving cart items and product details : %v", err)
		return nil, err
	}

	if len(cartItems) == 0 {
		return nil, utils.ErrEmptyCart
	}

	// Validate items and calculate total
	var totalAmount float64
	var checkoutItems []*domain.CheckoutItem

	// Iterate over cartitems
	for _, item := range cartItems {
		// Get product details to get the current stock quantity
		product, err := u.productRepo.GetByID(ctx, item.ProductID)
		if err != nil {
			log.Printf("error while retrieving product details from products table : %v", err)
			return nil, err
		}

		// Compare with stock quantity of the product
		if product.StockQuantity < item.Quantity {
			return nil, utils.ErrInsufficientStock
		}

		// Calculate the subtotal for each product
		subtotal := float64(item.Quantity) * product.Price

		// Add each checkout item
		checkoutItems = append(checkoutItems, &domain.CheckoutItem{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
			Price:     product.Price,
			Subtotal:  subtotal,
		})

		totalAmount += subtotal
	}

	// First we need to create checkout session
	session, err := u.checkoutRepo.CreateCheckoutSession(ctx, userID)
	if err != nil {
		log.Printf("error while creating checkout session : %v", err)
		return nil, err
	}

	// Add items to checkout items table and associate it to created checkout sessions
	err = u.checkoutRepo.AddCheckoutItems(ctx, session.ID, checkoutItems)
	if err != nil {
		log.Printf("error while adding items to checkout_items table : %v", err)
		return nil, err
	}

	// Update the session with calculated values
	session.TotalAmount = totalAmount
	session.ItemCount = len(checkoutItems) // item count : different product items, there is a limit for max different product items you can include in a checkout
	session.FinalAmount = totalAmount      // final amount is same as total amount, bcz coupon not applied
	session.UpdatedAt = time.Now().UTC()

	// Update the checkout session in the database
	err = u.checkoutRepo.UpdateCheckoutDetails(ctx, session)
	if err != nil {
		log.Printf("error while updating checkout details : %v", err)
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
func (u *checkoutUseCase) ApplyCoupon(ctx context.Context, userID int64, checkoutID int64, couponCode string) (*domain.ApplyCouponResponse, error) {
	// Get the checkout session
	checkout, err := u.checkoutRepo.GetCheckoutByID(ctx, checkoutID)
	if err != nil {
		log.Printf("error while retrieving checkout session details using id : %v", err)
		return nil, err
	}

	// Verify the checkout belongs to the user
	if checkout.UserID != userID {
		return nil, utils.ErrUnauthorized
	}

	// Double-check if the checkout is empty by counting items
	items, err := u.checkoutRepo.GetCheckoutItems(ctx, checkoutID)
	if err != nil {
		log.Printf("error while retrieving checkout item values : %v", err)
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
func (u *checkoutUseCase) UpdateCheckoutAddress(ctx context.Context, userID, checkoutID, addressID int64) (*domain.CheckoutSession, error) {
	// Get the checkout session
	checkout, err := u.checkoutRepo.GetCheckoutByID(ctx, checkoutID)
	if err != nil {
		log.Printf("error whiler retrieving checkout data using checkout id : %v", err)
		return nil, err
	}

	// Check if the checkout belongs to the user
	if checkout.UserID != userID {
		return nil, utils.ErrUnauthorized
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
	err = u.checkoutRepo.UpdateCheckoutShippingAddress(ctx, checkoutID, shippingAddressID)
	if err != nil {
		log.Printf("error while updating checkout session details with new shipping_address_id : %v", err)
		return nil, err
	}
	// Fetch the updated checkout session details from checkout_sessions table
	updatedCheckout, err := u.checkoutRepo.GetCheckoutByID(ctx, checkoutID)
	if err != nil {
		log.Printf("error while fetching checkout session details form checkout_sessions table : %v", err)
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
func (u *checkoutUseCase) RemoveAppliedCoupon(ctx context.Context, userID, checkoutID int64) (*domain.CheckoutSession, error) {
	// Get the checkout session
	checkout, err := u.checkoutRepo.GetCheckoutByID(ctx, checkoutID)
	if err != nil {
		if err == utils.ErrCheckoutNotFound {
			return nil, utils.ErrCheckoutNotFound
		}
		log.Printf("error while getting checkout using ID : %v", err)
		return nil, err
	}

	// Verify the checkout belongs to the user
	if checkout.UserID != userID {
		return nil, utils.ErrUnauthorized
	}

	// Check if the checkout is in a valid state to remove coupon
	if checkout.Status == utils.CheckoutStatusCompleted { // can;t remove coupon from completed checkouts
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
		return nil, err
	}

	return checkout, nil
}

/*
GetCheckoutSummary:
- Get details from checkout_sessions, checkout_items, products, shipping address tables
*/
func (u *checkoutUseCase) GetCheckoutSummary(ctx context.Context, userID, checkoutID int64) (*domain.CheckoutSummary, error) {
	// Get checkout details
	checkout, err := u.checkoutRepo.GetCheckoutByID(ctx, checkoutID)
	if err != nil {
		log.Printf("error while retrieving checkout details from checkout_sessions table : %v", err)
		return nil, err
	}

	// Verify the checkout belongs to the user
	if checkout.UserID != userID {
		return nil, utils.ErrUnauthorized
	}

	// Get checkout items with product details
	items, err := u.checkoutRepo.GetCheckoutItemsWithProductDetails(ctx, checkoutID)
	if err != nil {
		log.Printf("error while retrieving details from checkout_items and products table : %v", err)
		return nil, err
	}

	// Calculate actual item count and update if necessary
	actualItemCount := len(items) // count of different product items part of this checkout

	// Update the item count if it's different from item count associated with the checkout
	if actualItemCount != checkout.ItemCount {
		if err := u.checkoutRepo.UpdateCheckoutItemCount(ctx, checkoutID, actualItemCount); err != nil {
			log.Printf("error while updating the item count in checkout_sessions table : %v", err)
			return nil, err
		}
		// Update the variable associated with the checkout data
		checkout.ItemCount = actualItemCount
	}

	// Get shipping address if set
	var address *domain.ShippingAddress
	if checkout.ShippingAddressID != 0 {
		address, err = u.checkoutRepo.GetShippingAddress(ctx, checkout.ShippingAddressID)
		if err != nil {
			log.Printf("error while retrieving shipping address details : %v", err)
			return nil, err
		}
	} else {
		return nil, utils.ErrShippingAddressNotAssigned
	}

	// Update values for address response in checkout summmary
	addressResponse := &domain.ShippingAddressResponseInCheckoutSummary{
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

	// Assemble the checkout summary
	summary := &domain.CheckoutSummary{
		ID:             checkout.ID,
		UserID:         checkout.UserID,
		TotalAmount:    checkout.TotalAmount,
		DiscountAmount: checkout.DiscountAmount,
		FinalAmount:    checkout.FinalAmount,
		ItemCount:      checkout.ItemCount,
		Status:         checkout.Status,
		CouponCode:     checkout.CouponCode,
		CouponApplied:  checkout.CouponApplied,
		Address:        addressResponse,
		Items:          items,
	}

	// Handle empty checkout
	if len(items) == 0 {
		summary.Status = "empty"
		summary.TotalAmount = 0
		summary.DiscountAmount = 0
		summary.FinalAmount = 0
		summary.ItemCount = 0
	}

	return summary, nil
}
