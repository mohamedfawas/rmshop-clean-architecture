package validator

import (
	"regexp"

	"github.com/mohamedfawas/rmshop-clean-architecture/pkg/utils"
)

func ValidateAddressLine(addressLine string) error {
	if len(addressLine) < 10 {
		return utils.ErrUserAddressTooShort
	}

	if len(addressLine) > 255 {
		return utils.ErrUserAddressTooLong
	}

	return nil
}

func ValidateState(state string) error {
	if len(state) > 100 {
		return utils.ErrInvalidUserStateEntry
	}

	return nil
}

func ValidateCity(city string) error {
	if len(city) > 150 {
		return utils.ErrInvalidUserCityEntry
	}
	return nil
}

func ValidatePinCode(pinCode string) error {
	pinCodePattern := regexp.MustCompile(`^\d{6}$`)
	if !pinCodePattern.MatchString(pinCode) {
		return utils.ErrInvalidPinCode
	}

	return nil
}
