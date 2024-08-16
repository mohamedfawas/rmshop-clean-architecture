package validator

import (
	"strings"

	"github.com/mohamedfawas/rmshop-clean-architecture/pkg/utils"
)

func ValidateCategoryName(name string) error {

	if name == "" {
		return utils.ErrInvalidCategoryName
	}

	if len(name) < 2 {
		return utils.ErrCategoryNameTooShort
	}

	if len(name) > 50 { // limit : 50 chosen for better UI experience of user
		return utils.ErrCategoryNameTooLong
	}

	if strings.TrimLeft(name, "0123456789") == "" { //to disallow purely numeric names
		return utils.ErrCategoryNameNumeric
	}
	return nil
}
