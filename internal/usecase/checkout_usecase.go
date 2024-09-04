package usecase

import (
	"context"
	"database/sql"
	"time"

	"github.com/mohamedfawas/rmshop-clean-architecture/internal/domain"
	"github.com/mohamedfawas/rmshop-clean-architecture/internal/repository"
	"github.com/mohamedfawas/rmshop-clean-architecture/pkg/utils"
)

type CheckoutUseCase interface {
	CreateCheckout(ctx context.Context, userID int64) (*domain.CheckoutSession, error)
	ApplyCoupon(ctx context.Context, userID int64, checkoutID int64, couponCode string) (*domain.ApplyCouponResponse, error)
	UpdateCheckoutAddress(ctx context.Context, userID, checkoutID int64, addressInput domain.AddressInput) (*domain.CheckoutSession, error)
	GetCheckoutSummary(ctx context.Context, userID, checkoutID int64) (*domain.CheckoutSummary, error)
}

type checkoutUseCase struct {
	checkoutRepo repository.CheckoutRepository
	productRepo  repository.ProductRepository
	couponRepo   repository.CouponRepository
	cartRepo     repository.CartRepository
	userRepo     repository.UserRepository
}

func NewCheckoutUseCase(checkoutRepo repository.CheckoutRepository, productRepo repository.ProductRepository, cartRepo repository.CartRepository, couponRepo repository.CouponRepository, userRepo repository.UserRepository) CheckoutUseCase {
	return &checkoutUseCase{
		checkoutRepo: checkoutRepo,
		productRepo:  productRepo,
		couponRepo:   couponRepo,
		cartRepo:     cartRepo,
		userRepo:     userRepo,
	}
}

func (u *checkoutUseCase) CreateCheckout(ctx context.Context, userID int64) (*domain.CheckoutSession, error) {
	// Get cart items
	cartItems, err := u.checkoutRepo.GetCartItems(ctx, userID)
	if err != nil {
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
		return nil, err
	}

	// Add items to checkout session
	err = u.checkoutRepo.AddCheckoutItems(ctx, session.ID, checkoutItems)
	if err != nil {
		return nil, err
	}

	// Update the session with calculated values
	session.TotalAmount = totalAmount
	session.ItemCount = len(checkoutItems)
	session.FinalAmount = totalAmount

	// Update the checkout session in the database
	err = u.checkoutRepo.UpdateCheckout(ctx, session)
	if err != nil {
		return nil, err
	}

	return session, nil
}

func (u *checkoutUseCase) ApplyCoupon(ctx context.Context, userID int64, checkoutID int64, couponCode string) (*domain.ApplyCouponResponse, error) {
	// Get the checkout session
	checkout, err := u.checkoutRepo.GetCheckoutByID(ctx, checkoutID)
	if err != nil {
		return nil, err
	}

	// Verify the checkout belongs to the user
	if checkout.UserID != userID {
		return nil, utils.ErrUnauthorized
	}

	// Double-check if the checkout is empty by counting items
	items, err := u.checkoutRepo.GetCheckoutItems(ctx, checkoutID)
	if err != nil {
		return nil, err
	}

	if len(items) == 0 {
		return nil, utils.ErrEmptyCheckout
	}

	// Update ItemCount if it's inconsistent
	if checkout.ItemCount != len(items) {
		checkout.ItemCount = len(items)
		// You might want to log this inconsistency
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
	err = u.checkoutRepo.UpdateCheckout(ctx, checkout)
	if err != nil {
		return nil, err
	}

	return &domain.ApplyCouponResponse{
		CheckoutSession: *checkout,
		Message:         message,
	}, nil
}

func (u *checkoutUseCase) UpdateCheckoutAddress(ctx context.Context, userID, checkoutID int64, addressInput domain.AddressInput) (*domain.CheckoutSession, error) {
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

	if addressInput.AddressID != 0 {
		// Update with existing address
		address, err := u.userRepo.GetUserAddressByID(ctx, addressInput.AddressID)
		if err != nil {
			return nil, err
		}
		if address.UserID != userID {
			return nil, utils.ErrUnauthorized
		}
		err = u.checkoutRepo.UpdateCheckoutAddress(ctx, checkoutID, addressInput.AddressID)
		if err != nil {
			return nil, err
		}
	} else if addressInput.NewAddress != nil {
		// Validate new address
		if err := validateAddress(addressInput.NewAddress); err != nil {
			return nil, err
		}
		// Add new address and update checkout
		err = u.checkoutRepo.AddNewAddressToCheckout(ctx, checkoutID, addressInput.NewAddress)
		if err != nil {
			return nil, err
		}
	} else {
		return nil, utils.ErrInvalidAddressInput
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
