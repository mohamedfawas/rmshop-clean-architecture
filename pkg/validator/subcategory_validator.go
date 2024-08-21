package validator

import (
	"strings"

	"github.com/mohamedfawas/rmshop-clean-architecture/pkg/utils"
)

func ValidateSubCategoryName(name string) error {

	if name == "" {
		return utils.ErrInvalidSubCategoryName
	}

	if len(name) < 2 {
		return utils.ErrSubCategoryNameTooShort
	}

	if len(name) > 50 { // limit : 50 chosen for better UI experience of user
		return utils.ErrSubCategoryNameTooLong
	}

	if strings.TrimLeft(name, "0123456789") == "" { //to disallow purely numeric names
		return utils.ErrSubCategoryNameNumeric
	}
	return nil
}
