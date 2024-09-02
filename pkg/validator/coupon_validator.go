package validator

import (
	"regexp"
	"time"

	"github.com/mohamedfawas/rmshop-clean-architecture/internal/domain"
	"github.com/mohamedfawas/rmshop-clean-architecture/pkg/utils"
)

func ValidateCouponInput(input domain.CreateCouponInput) error {
	// Validate coupon code format
	if !isValidCouponCode(input.Code) {
		return utils.ErrInvalidCouponCode
	}

	// Validate discount percentage
	if !isValidDiscountPercentage(input.DiscountPercentage) {
		return utils.ErrInvalidDiscountPercentage
	}

	// Validate minimum order amount
	if !isValidMinOrderAmount(input.MinOrderAmount) {
		return utils.ErrInvalidMinOrderAmount
	}

	// Validate expiry date
	if !isValidExpiryDate(input.ExpiresAt) {
		return utils.ErrInvalidExpiryDate
	}

	return nil
}

// isValidCouponCode checks if the coupon code is alphanumeric and between 4 to 20 characters
func isValidCouponCode(code string) bool {
	match, _ := regexp.MatchString(`^[A-Z0-9]{4,20}$`, code)
	return match
}

// isValidDiscountPercentage checks if the discount percentage is between 0 and 100
func isValidDiscountPercentage(percentage float64) bool {
	return percentage > 0 && percentage <= 100
}

// isValidMinOrderAmount checks if the minimum order amount is non-negative
func isValidMinOrderAmount(amount float64) bool {
	return amount >= 0
}

// isValidExpiryDate checks if the expiry date is in the future and in the correct format
func isValidExpiryDate(expiresAtStr string) bool {
	if expiresAtStr == "" {
		return true // No expiry date is considered valid
	}

	expiresAt, err := time.Parse("2006-01-02", expiresAtStr)
	if err != nil {
		return false // Invalid date format
	}

	return expiresAt.After(time.Now().UTC())
}
