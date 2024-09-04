package usecase

import (
	"context"
	"log"
	"time"

	"github.com/mohamedfawas/rmshop-clean-architecture/internal/domain"
	"github.com/mohamedfawas/rmshop-clean-architecture/internal/repository"
	"github.com/mohamedfawas/rmshop-clean-architecture/pkg/utils"
	"github.com/mohamedfawas/rmshop-clean-architecture/pkg/validator"
)

type CouponUseCase interface {
	CreateCoupon(ctx context.Context, input domain.CreateCouponInput) (*domain.Coupon, error)
	ApplyCoupon(ctx context.Context, userID int64, checkoutID int64, couponCode string) (*domain.ApplyCouponResponse, error)
}

type couponUseCase struct {
	couponRepo   repository.CouponRepository
	checkoutRepo repository.CheckoutRepository
}

func NewCouponUseCase(couponRepo repository.CouponRepository, checkoutRepo repository.CheckoutRepository) CouponUseCase {
	return &couponUseCase{
		couponRepo:   couponRepo,
		checkoutRepo: checkoutRepo,
	}
}

func (u *couponUseCase) CreateCoupon(ctx context.Context, input domain.CreateCouponInput) (*domain.Coupon, error) {
	// Validate input
	if err := validator.ValidateCouponInput(input); err != nil {
		log.Printf("validation error : %v", err)
		return nil, err
	}

	// Parse the expiry date
	var expiresAt *time.Time
	if input.ExpiresAt != "" {
		parsedTime, err := time.Parse("2006-01-02", input.ExpiresAt)
		if err != nil {
			return nil, utils.ErrInvalidExpiryDate
		}
		expiresAt = &parsedTime
	}

	// Check if coupon with the same code already exists
	existingCoupon, err := u.couponRepo.GetByCode(ctx, input.Code)
	if err != nil && err != utils.ErrCouponNotFound {
		log.Printf("error while retrieving the coupon details using id : %v", err)
		return nil, err
	}
	if existingCoupon != nil {
		return nil, utils.ErrDuplicateCouponCode
	}

	// Create new coupon
	now := time.Now().UTC()
	coupon := &domain.Coupon{
		Code:               input.Code,
		DiscountPercentage: input.DiscountPercentage,
		MinOrderAmount:     input.MinOrderAmount,
		IsActive:           true,
		CreatedAt:          now,
		UpdatedAt:          now,
		ExpiresAt:          expiresAt,
	}

	// Save coupon to database
	err = u.couponRepo.Create(ctx, coupon)
	if err != nil {
		log.Printf("error while creating coupon entry in db : %v", err)
		return nil, err
	}

	return coupon, nil
}

func (u *couponUseCase) ApplyCoupon(ctx context.Context, userID int64, checkoutID int64, couponCode string) (*domain.ApplyCouponResponse, error) {
	// Get the checkout session
	checkout, err := u.checkoutRepo.GetCheckoutByID(ctx, checkoutID)
	if err != nil {
		return nil, err
	}

	// Verify the checkout belongs to the user
	if checkout.UserID != userID {
		return nil, utils.ErrUnauthorized
	}

	// Check if the checkout is empty
	if checkout.ItemCount == 0 {
		return nil, utils.ErrEmptyCheckout
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

	log.Printf("min order amount : %v", coupon.MinOrderAmount)
	log.Printf("total amount : %v", checkout.TotalAmount)
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
		return nil, err
	}

	return &domain.ApplyCouponResponse{
		CheckoutSession: *checkout,
		Message:         message,
	}, nil
}
