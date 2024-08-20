package validator

import (
	"strings"

	"github.com/mohamedfawas/rmshop-clean-architecture/pkg/utils"
)

func ValidateAuthHeaderAndReturnToken(authHeader string) (string, error) {
	if authHeader == "" {
		return "", utils.ErrMissingAuthToken
	}

	// Check for correct Authorization scheme
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return "", utils.ErrAuthHeaderFormat
	}

	token := parts[1]
	if token == "" {
		return "", utils.ErrEmptyToken
	}

	return token, nil
}
