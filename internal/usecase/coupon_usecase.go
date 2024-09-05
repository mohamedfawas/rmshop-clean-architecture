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
	GetAllCoupons(ctx context.Context, params domain.CouponQueryParams) ([]*domain.Coupon, int64, error)
	UpdateCoupon(ctx context.Context, couponID int64, input domain.CouponUpdateInput) (*domain.Coupon, error)
	SoftDeleteCoupon(ctx context.Context, couponID int64) error
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

func (u *couponUseCase) GetAllCoupons(ctx context.Context, params domain.CouponQueryParams) ([]*domain.Coupon, int64, error) {
	// Validate and set default values
	if params.Page < 1 {
		params.Page = 1
	}
	if params.Limit < 1 {
		params.Limit = 10
	} else if params.Limit > 100 {
		params.Limit = 100
	}

	// Validate sorting parameters
	validSortFields := map[string]bool{"created_at": true, "discount_percentage": true, "min_order_amount": true}
	if params.Sort != "" && !validSortFields[params.Sort] {
		params.Sort = "created_at"
	}
	if params.Order != "asc" && params.Order != "desc" {
		params.Order = "desc"
	}

	// Set the current time for filtering out expired coupons
	params.CurrentTime = time.Now()

	// Call repository method
	return u.couponRepo.GetAllCoupons(ctx, params)
}

func (u *couponUseCase) UpdateCoupon(ctx context.Context, couponID int64, input domain.CouponUpdateInput) (*domain.Coupon, error) {
	coupon, err := u.couponRepo.GetByID(ctx, couponID)
	if err != nil {
		return nil, utils.ErrCouponNotFound
	}

	if input.Code != nil {
		if err := validator.ValidateCouponCode(*input.Code); err != nil {
			return nil, utils.ErrInvalidCouponCode
		}
		coupon.Code = *input.Code
	}

	if input.DiscountPercentage != nil {
		if err := validator.ValidateDiscountPercentage(*input.DiscountPercentage); err != nil {
			return nil, utils.ErrInvalidDiscountPercentage
		}
		coupon.DiscountPercentage = *input.DiscountPercentage
	}

	if input.MinOrderAmount != nil {
		if err := validator.ValidateMinOrderAmount(*input.MinOrderAmount); err != nil {
			return nil, utils.ErrInvalidMinOrderAmount
		}
		coupon.MinOrderAmount = *input.MinOrderAmount
	}

	if input.ExpiresAt != nil {
		expiryTime, err := time.Parse("2006-01-02", *input.ExpiresAt)
		if err != nil {
			return nil, utils.ErrInvalidExpiryDate
		}
		// Set the time to end of day (23:59:59)
		expiryTime = time.Date(expiryTime.Year(), expiryTime.Month(), expiryTime.Day(), 23, 59, 59, 0, time.UTC)
		if expiryTime.Before(time.Now()) {
			return nil, utils.ErrInvalidExpiryDate
		}
		coupon.ExpiresAt = &expiryTime
	}

	coupon.UpdatedAt = time.Now()

	err = u.couponRepo.Update(ctx, coupon)
	if err != nil {
		return nil, err
	}

	return coupon, nil
}

func (u *couponUseCase) SoftDeleteCoupon(ctx context.Context, couponID int64) error {
	// Check if the coupon exists
	coupon, err := u.couponRepo.GetByID(ctx, couponID)
	if err != nil {
		if err == utils.ErrCouponNotFound {
			return utils.ErrCouponNotFound
		}
		return err
	}

	// Check if the coupon is already soft deleted
	if coupon.IsDeleted {
		return utils.ErrCouponAlreadyDeleted
	}

	// Check if the coupon is in use
	isInUse, err := u.couponRepo.IsCouponInUse(ctx, couponID)
	if err != nil {
		return err
	}
	if isInUse {
		return utils.ErrCouponInUse
	}

	// Perform soft delete
	err = u.couponRepo.SoftDelete(ctx, couponID)
	if err != nil {
		return err
	}

	return nil
}
