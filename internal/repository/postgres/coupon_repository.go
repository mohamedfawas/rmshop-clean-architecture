package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"

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

/*
GetByCode :
- Get coupon details from coupons table for the given coupon code
- code, discount_percentage, min_order_amount, is_active and expires_at values are taken
*/
func (r *couponRepository) GetByCode(ctx context.Context, code string) (*domain.Coupon, error) {
	query := `
        SELECT id, code, discount_percentage, min_order_amount, is_active, created_at, updated_at, expires_at
        FROM coupons
        WHERE code = $1 AND is_active = true
    `
	var coupon domain.Coupon
	err := r.db.QueryRowContext(ctx, query, code).Scan(
		&coupon.ID,
		&coupon.Code,
		&coupon.DiscountPercentage,
		&coupon.MinOrderAmount,
		&coupon.IsActive,
		&coupon.CreatedAt,
		&coupon.UpdatedAt,
		&coupon.ExpiresAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, utils.ErrCouponNotFound
		}
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

	// Prepare the initial query arguments with the current time, which will be used to filter out expired coupons
	args := []interface{}{params.CurrentTime} // []interface{} defines slice to hold any type of values (interface)
	// Initialize an empty slice for additional query conditions
	conditions := []string{}

	// Add a condition to filter active coupons if the status is set to "active"
	if params.Status == "active" {
		conditions = append(conditions, "is_active = true")
	}

	// Add a condition to filter coupons with a discount greater than or equal to the provided minimum discount
	if params.MinDiscount != nil {
		// Use positional arguments ($) for parameterized queries and increment as new conditions are added.
		conditions = append(conditions, fmt.Sprintf("discount_percentage >= $%d", len(args)+1))
		args = append(args, *params.MinDiscount)
	}

	// Add a condition to filter coupons with a discount less than or equal to the provided maximum discount
	if params.MaxDiscount != nil {
		conditions = append(conditions, fmt.Sprintf("discount_percentage <= $%d", len(args)+1))
		args = append(args, *params.MaxDiscount)
	}

	// Add a condition to filter coupons that match the search query (partial matching using ILIKE for case-insensitive search).
	if params.Search != "" {
		conditions = append(conditions, fmt.Sprintf("code ILIKE $%d", len(args)+1))
		args = append(args, "%"+params.Search+"%")
	}

	// If there are any conditions, append them to both the main query and the count query
	if len(conditions) > 0 {
		query += " AND " + strings.Join(conditions, " AND ")
		countQuery += " AND " + strings.Join(conditions, " AND ")
	}

	// Validate and set the sorting column. Default to "created_at" if the provided sort field is not valid
	sortColumn := "created_at"
	sortOrder := "DESC" // Default sorting order is descending

	// A map of valid sortable columns
	validColumns := map[string]bool{"created_at": true, "discount_percentage": true, "min_order_amount": true, "code": true}
	if validColumns[params.Sort] {
		sortColumn = params.Sort
	}

	// Validate and set the sort order to either ascending (ASC) or descending (DESC)
	if strings.ToUpper(params.Order) == "ASC" {
		sortOrder = "ASC"
	}

	// Append the ORDER BY clause to the query, using the validated sort column and order
	query += fmt.Sprintf(" ORDER BY %s %s", sortColumn, sortOrder)

	// Add pagination to the query using LIMIT for the number of results and OFFSET for the starting point based on the page number
	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", len(args)+1, len(args)+2)
	args = append(args, params.Limit, (params.Page-1)*params.Limit)

	// Execute the count query to get the total number of coupons matching the filters
	var totalCount int64
	err := r.db.QueryRowContext(ctx, countQuery, args[:len(args)-2]...).Scan(&totalCount)
	if err != nil {
		log.Printf("error while retrieving the total number of coupons matching the filters : %v", err)
		return nil, 0, err
	}

	// Execute the main query to get the list of coupons
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		log.Printf("error while retrieving the list of coupons : %v", err)
		return nil, 0, err
	}
	defer rows.Close()

	// Initialize a slice to hold the retrieved coupons
	var coupons []*domain.Coupon
	for rows.Next() {
		// For each row, scan the result into a Coupon object
		var c domain.Coupon
		err := rows.Scan(&c.ID, &c.Code, &c.DiscountPercentage, &c.MinOrderAmount, &c.IsActive, &c.CreatedAt, &c.UpdatedAt, &c.ExpiresAt)
		if err != nil {
			log.Printf("error while adding the coupon row to the 'coupon' slice : %v", err)
			return nil, 0, err
		}
		// Append the coupon to the list
		coupons = append(coupons, &c)
	}

	// Check if there were any errors encountered during the iteration over the rows
	if err = rows.Err(); err != nil {
		log.Printf("error while iterating over the rows : %v", err)
		return nil, 0, err
	}

	// Return the list of coupons, the total count, and no error
	return coupons, totalCount, nil
}

func (r *couponRepository) GetByID(ctx context.Context, id int64) (*domain.Coupon, error) {
	query := `
        SELECT id, code, discount_percentage, min_order_amount, is_active, 
               created_at, updated_at, expires_at
        FROM coupons 
        WHERE id = $1 AND is_active = true
    `
	var coupon domain.Coupon
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&coupon.ID, &coupon.Code, &coupon.DiscountPercentage, &coupon.MinOrderAmount,
		&coupon.IsActive, &coupon.CreatedAt, &coupon.UpdatedAt, &coupon.ExpiresAt,
	)
	if err == sql.ErrNoRows {
		return nil, utils.ErrCouponNotFound
	}
	if err != nil {
		log.Printf("error while retrieving coupon details using ID : %v", err)
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
	if err != nil {
		log.Printf("error while updating the coupon details : %v", err)
	}
	return err
}
