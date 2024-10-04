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

	GetAllCoupons(ctx context.Context, params domain.CouponQueryParams) ([]*domain.Coupon, int64, error)
	UpdateCoupon(ctx context.Context, couponID int64, input domain.CouponUpdateInput) (*domain.Coupon, error)
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
	// Validate input coupon details
	if err := validator.ValidateCouponInput(input); err != nil {
		log.Printf("validation error : %v", err)
		return nil, err
	}

	// Parse the expiry date
	var expiresAt *time.Time // expiry date can be null
	if input.ExpiresAt != "" {
		parsedTime, err := time.Parse("2006-01-02", input.ExpiresAt)
		if err != nil {
			return nil, utils.ErrInvalidExpiryDate
		}
		expiresAt = &parsedTime
	}

	// Check if coupon with the same code already exists
	existingCoupon, err := u.couponRepo.GetByCode(ctx, input.Code)
	if err != nil && err != utils.ErrCouponNotFound { // no problem if coupon is not found
		log.Printf("error while retrieving the coupon details using id : %v", err)
		return nil, err
	}
	// if coupon already exists
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
		log.Printf("error while creating coupon entry in database : %v", err)
		return nil, err
	}

	return coupon, nil
}

func (u *couponUseCase) GetAllCoupons(ctx context.Context, params domain.CouponQueryParams) ([]*domain.Coupon, int64, error) {
	// Validate and set default values
	if params.Page < 1 {
		params.Page = 1
	}
	if params.Limit < 1 {
		params.Limit = 10
	} else if params.Limit > 20 {
		params.Limit = 20
	}

	// Define valid sorting fields for coupons (e.g., by creation date, discount, or minimum order amount).
	validSortFields := map[string]bool{"created_at": true, "discount_percentage": true, "min_order_amount": true}
	// If the `Sort` field is provided but is not one of the valid fields, default it to "created_at".
	if params.Sort != "" && !validSortFields[params.Sort] {
		params.Sort = "created_at"
	}

	if params.Order != "asc" && params.Order != "desc" {
		params.Order = "desc"
	}

	// Set the current time to filter out expired coupons later in the repository query.
	params.CurrentTime = time.Now().UTC()

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

	coupon.UpdatedAt = time.Now().UTC()

	err = u.couponRepo.Update(ctx, coupon)
	if err != nil {
		log.Printf("error : %v", err)
		return nil, err
	}

	return coupon, nil
}
