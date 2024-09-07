package usecase

import (
	"context"

	"github.com/mohamedfawas/rmshop-clean-architecture/internal/domain"
)

type CouponUseCase interface {
	CreateCoupon(ctx context.Context, input domain.CreateCouponInput) (*domain.Coupon, error)
	ApplyCoupon(ctx context.Context, userID int64, checkoutID int64, couponCode string) (*domain.ApplyCouponResponse, error)
	GetAllCoupons(ctx context.Context, params domain.CouponQueryParams) ([]*domain.Coupon, int64, error)
	UpdateCoupon(ctx context.Context, couponID int64, input domain.CouponUpdateInput) (*domain.Coupon, error)
}
