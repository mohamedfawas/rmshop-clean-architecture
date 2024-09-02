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
	ApplyCoupon(ctx context.Context, userID int64, input domain.ApplyCouponInput) (*domain.ApplyCouponResponse, error)
}

type couponUseCase struct {
	couponRepo repository.CouponRepository
	cartRepo   repository.CartRepository
}

func NewCouponUseCase(couponRepo repository.CouponRepository, cartRepo repository.CartRepository) CouponUseCase {
	return &couponUseCase{
		couponRepo: couponRepo,
		cartRepo:   cartRepo,
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

func (u *couponUseCase) ApplyCoupon(ctx context.Context, userID int64, input domain.ApplyCouponInput) (*domain.ApplyCouponResponse, error) {
	// Check if a coupon is already applied
	appliedCoupon, err := u.cartRepo.GetAppliedCoupon(ctx, userID)
	if err != nil {
		return nil, err
	}
	if appliedCoupon != nil {
		return nil, utils.ErrCouponAlreadyApplied
	}

	// Get the coupon
	coupon, err := u.couponRepo.GetByCode(ctx, input.CouponCode)
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
	if coupon.ExpiresAt.Before(time.Now()) {
		return nil, utils.ErrCouponExpired
	}

	// Get the cart total
	cartTotal, err := u.cartRepo.GetCartTotal(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Check if the cart is empty
	if cartTotal == 0 {
		return nil, utils.ErrEmptyCart
	}

	// Check if the minimum order amount is met
	if cartTotal < coupon.MinOrderAmount {
		return nil, utils.ErrOrderTotalBelowMinimum
	}

	// Calculate the discount
	discountAmount := cartTotal * (coupon.DiscountPercentage / 100)
	totalAfterDiscount := cartTotal - discountAmount

	// Apply maximum discount cap if necessary
	note := ""
	if discountAmount > utils.MaxDiscountAmount {
		discountAmount = utils.MaxDiscountAmount
		totalAfterDiscount = cartTotal - discountAmount
		note = "Maximum discount cap applied"
	}

	// Apply the coupon
	err = u.cartRepo.ApplyCoupon(ctx, userID, coupon)
	if err != nil {
		return nil, err
	}

	return &domain.ApplyCouponResponse{
		DiscountAmount:     discountAmount,
		TotalAfterDiscount: totalAfterDiscount,
		Note:               note,
	}, nil
}
