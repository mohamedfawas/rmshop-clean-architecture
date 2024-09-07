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

// CreateCoupon handles the creation of a new coupon by validating the input,
// checking for existing coupons with the same code,
// and inserting the new coupon into the database.
// It returns the created coupon if successful or an error if validation fails,
// the coupon already exists, or if there is an issue with database operations.
//
// Parameters:
// - ctx: context for managing request-scoped values and cancellation signals.
// - input: domain.CreateCouponInput object containing the details of the coupon to be created.
//
// Returns:
// - *domain.Coupon: a pointer to the created Coupon object if the creation is successful.
// - error: an error if the creation fails due to validation errors, existing coupon with the same code, or database errors.
//
// Possible errors:
// - utils.ErrInvalidExpiryDate: returned if the provided expiry date is invalid and cannot be parsed.
// - utils.ErrDuplicateCouponCode: returned if a coupon with the same code already exists in the database.
// - error: returned if any other error occurs during the coupon retrieval or creation process.
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

func (u *couponUseCase) ApplyCoupon(ctx context.Context, userID int64, checkoutID int64, couponCode string) (*domain.ApplyCouponResponse, error) {
	// Get the checkout session
	checkout, err := u.checkoutRepo.GetCheckoutByID(ctx, checkoutID)
	if err != nil {
		log.Printf("error while retrieving checkout session using ID : %v", err)
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
		log.Printf("error while retrieving coupon using coupon code : %v", err)
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
		log.Printf("error while updating checkout details : %v", err)
		return nil, err
	}

	return &domain.ApplyCouponResponse{
		CheckoutSession: *checkout,
		Message:         message,
	}, nil
}

// GetAllCoupons validates the query parameters, sets default values for pagination and sorting,
// and delegates the query execution to the coupon repository to retrieve a list of coupons.
// It ensures that pagination limits, sorting fields, and sort order are within valid bounds
// and sets the current time to filter out expired coupons.
//
// Parameters:
//   - ctx: A context object that manages request deadlines, cancelation signals,
//          and other request-scoped values.
//   - params: A CouponQueryParams struct containing filters such as status,
//             discount range, pagination, and sorting.
//
// Returns:
//   - []*domain.Coupon: A slice of pointers to Coupon objects that match the query parameters.
//   - int64: The total count of matching coupons for pagination purposes.
//   - error: Returns an error if any issue occurs during query validation or execution.

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

// UpdateCoupon updates an existing coupon with the provided input data.
// It validates the input fields before updating the coupon and ensures that the
// coupon code, discount percentage, minimum order amount, and expiry date are valid.
//
// Parameters:
//   - ctx: The context for managing request flow, cancellation, and timeouts.
//   - couponID: The ID of the coupon to be updated.
//   - input: A domain.CouponUpdateInput struct containing the fields to update.
//
// Returns:
//   - *domain.Coupon: Returns the updated coupon object upon successful update.
//   - error: Returns an error if the coupon is not found, if validation fails, or if
//     there is an issue updating the coupon in the repository.
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
