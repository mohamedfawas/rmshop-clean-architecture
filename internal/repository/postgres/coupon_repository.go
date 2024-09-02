package postgres

import (
	"context"
	"database/sql"
	"log"

	"github.com/mohamedfawas/rmshop-clean-architecture/internal/domain"
	"github.com/mohamedfawas/rmshop-clean-architecture/pkg/utils"
)

type couponRepository struct {
	db *sql.DB
}

func NewCouponRepository(db *sql.DB) *couponRepository {
	return &couponRepository{db: db}
}

func (r *couponRepository) Create(ctx context.Context, coupon *domain.Coupon) error {
	query := `
		INSERT INTO coupons (code, discount_percentage, min_order_amount, is_active, created_at, updated_at, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id`

	err := r.db.QueryRowContext(ctx, query,
		coupon.Code,
		coupon.DiscountPercentage,
		coupon.MinOrderAmount,
		coupon.IsActive,
		coupon.CreatedAt,
		coupon.UpdatedAt,
		coupon.ExpiresAt,
	).Scan(&coupon.ID)

	if err != nil {
		// Check for unique constraint violation
		if utils.IsDuplicateKeyError(err) {
			return utils.ErrDuplicateCouponCode
		}
		log.Printf("error while creating the coupon entry : %v", err)
		return err
	}

	return nil
}

func (r *couponRepository) GetByCode(ctx context.Context, code string) (*domain.Coupon, error) {
	query := `
        SELECT id, code, discount_percentage, min_order_amount, is_active, created_at, updated_at, expires_at
        FROM coupons
        WHERE code = $1 AND is_active = true
    `
	var coupon domain.Coupon
	err := r.db.QueryRowContext(ctx, query, code).Scan(
		&coupon.ID, &coupon.Code, &coupon.DiscountPercentage, &coupon.MinOrderAmount,
		&coupon.IsActive, &coupon.CreatedAt, &coupon.UpdatedAt, &coupon.ExpiresAt,
	)
	if err == sql.ErrNoRows {
		return nil, utils.ErrCouponNotFound
	}
	if err != nil {
		log.Printf("error while retrieving coupon details : %v", err)
		return nil, err
	}
	return &coupon, nil
}
