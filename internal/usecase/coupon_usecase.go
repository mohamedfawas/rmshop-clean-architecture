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
}

type couponUseCase struct {
	couponRepo repository.CouponRepository
}

func NewCouponUseCase(couponRepo repository.CouponRepository) CouponUseCase {
	return &couponUseCase{couponRepo: couponRepo}
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
