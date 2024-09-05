package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

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

func (r *couponRepository) GetAllCoupons(ctx context.Context, params domain.CouponQueryParams) ([]*domain.Coupon, int64, error) {
	query := `
		SELECT id, code, discount_percentage, min_order_amount, is_active, created_at, updated_at, expires_at
		FROM coupons
		WHERE (expires_at IS NULL OR expires_at > $1)
	`

	countQuery := `
		SELECT COUNT(*)
		FROM coupons
		WHERE (expires_at IS NULL OR expires_at > $1)
	`

	args := []interface{}{params.CurrentTime}
	conditions := []string{}

	if params.Status == "active" {
		conditions = append(conditions, "is_active = true")
	}

	if params.MinDiscount != nil {
		conditions = append(conditions, fmt.Sprintf("discount_percentage >= $%d", len(args)+1))
		args = append(args, *params.MinDiscount)
	}

	if params.MaxDiscount != nil {
		conditions = append(conditions, fmt.Sprintf("discount_percentage <= $%d", len(args)+1))
		args = append(args, *params.MaxDiscount)
	}

	if params.Search != "" {
		conditions = append(conditions, fmt.Sprintf("code ILIKE $%d", len(args)+1))
		args = append(args, "%"+params.Search+"%")
	}

	if len(conditions) > 0 {
		query += " AND " + strings.Join(conditions, " AND ")
		countQuery += " AND " + strings.Join(conditions, " AND ")
	}

	// Validate and set sorting
	sortColumn := "created_at"
	sortOrder := "DESC"

	validColumns := map[string]bool{"created_at": true, "discount_percentage": true, "min_order_amount": true, "code": true}
	if validColumns[params.Sort] {
		sortColumn = params.Sort
	}

	if strings.ToUpper(params.Order) == "ASC" {
		sortOrder = "ASC"
	}

	query += fmt.Sprintf(" ORDER BY %s %s", sortColumn, sortOrder)

	// Add pagination
	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", len(args)+1, len(args)+2)
	args = append(args, params.Limit, (params.Page-1)*params.Limit)

	// Execute count query
	var totalCount int64
	err := r.db.QueryRowContext(ctx, countQuery, args[:len(args)-2]...).Scan(&totalCount)
	if err != nil {
		return nil, 0, err
	}

	// Execute main query
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var coupons []*domain.Coupon
	for rows.Next() {
		var c domain.Coupon
		err := rows.Scan(&c.ID, &c.Code, &c.DiscountPercentage, &c.MinOrderAmount, &c.IsActive, &c.CreatedAt, &c.UpdatedAt, &c.ExpiresAt)
		if err != nil {
			return nil, 0, err
		}
		coupons = append(coupons, &c)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, err
	}

	return coupons, totalCount, nil
}

func (r *couponRepository) GetByID(ctx context.Context, id int64) (*domain.Coupon, error) {
	query := `
        SELECT id, code, discount_percentage, min_order_amount, is_active, 
               created_at, updated_at, expires_at, deleted_at, is_deleted
        FROM coupons 
        WHERE id = $1 AND is_deleted = false
    `
	var coupon domain.Coupon
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&coupon.ID, &coupon.Code, &coupon.DiscountPercentage, &coupon.MinOrderAmount,
		&coupon.IsActive, &coupon.CreatedAt, &coupon.UpdatedAt, &coupon.ExpiresAt,
		&coupon.DeletedAt, &coupon.IsDeleted,
	)
	if err == sql.ErrNoRows {
		return nil, utils.ErrCouponNotFound
	}
	if err != nil {
		return nil, err
	}
	return &coupon, nil
}

func (r *couponRepository) Update(ctx context.Context, coupon *domain.Coupon) error {
	query := `UPDATE coupons 
              SET code = $1, discount_percentage = $2, min_order_amount = $3, 
                  is_active = $4, updated_at = $5, expires_at = $6
              WHERE id = $7`

	_, err := r.db.ExecContext(ctx, query,
		coupon.Code, coupon.DiscountPercentage, coupon.MinOrderAmount,
		coupon.IsActive, coupon.UpdatedAt, coupon.ExpiresAt, coupon.ID,
	)
	return err
}

func (r *couponRepository) SoftDelete(ctx context.Context, couponID int64) error {
	query := `
        UPDATE coupons 
        SET is_deleted = true, deleted_at = $1, updated_at = $1, is_active = false 
        WHERE id = $2 AND is_deleted = false
    `
	result, err := r.db.ExecContext(ctx, query, time.Now(), couponID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return utils.ErrCouponNotFound
	}

	return nil
}

func (r *couponRepository) IsCouponInUse(ctx context.Context, couponID int64) (bool, error) {
	query := `
        SELECT EXISTS(
            SELECT 1 
            FROM checkout_sessions 
            WHERE coupon_code = (SELECT code FROM coupons WHERE id = $1)
            AND created_at > NOW() - INTERVAL '30 days'
            AND status != 'cancelled'
        )
    `
	var inUse bool
	err := r.db.QueryRowContext(ctx, query, couponID).Scan(&inUse)
	if err != nil {
		return false, err
	}
	return inUse, nil
}
