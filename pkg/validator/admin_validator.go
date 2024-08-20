package validator

import (
	"github.com/mohamedfawas/rmshop-clean-architecture/pkg/utils"
)

func ValidateAdminCredentials(username, password string) error {
	if username == "" || password == "" {
		return utils.ErrMissingAdminCredentials
	}

	// Check username length (assuming max length is 50)
	if len(username) > 50 {
		return utils.ErrAdminUsernameTooLong
	}

	// Check password length (assuming max length is 64)
	if len(password) > 64 {
		return utils.ErrAdminPasswordTooLong
	}

	return nil
}
