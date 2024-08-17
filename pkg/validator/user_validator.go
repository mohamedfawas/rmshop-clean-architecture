package validator

import (
	"regexp"
	"time"
	"unicode"

	"github.com/mohamedfawas/rmshop-clean-architecture/pkg/utils"
)

func ValidateUserName(name string) error {
	if name == "" {
		return utils.ErrInvalidUserName
	}
	if len(name) < 2 {
		return utils.ErrUserNameTooShort
	}
	if len(name) > 200 {
		return utils.ErrUserNameTooLong
	}
	if matched, _ := regexp.MatchString(`[0-9]`, name); matched {
		return utils.ErrUserNameWithNumericVals
	}

	return nil
}

func ValidateUserEmail(email string) error {

	if email == "" {
		return utils.ErrMissingEmail
	}

	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(email) {
		return utils.ErrInvalidEmail
	}
	return nil
}

func ValidatePassword(password string) error {
	if password == "" {
		return utils.ErrPasswordInvalid
	}
	if len(password) < 8 {
		return utils.ErrPasswordTooShort
	}
	if len(password) > 64 {
		return utils.ErrPasswordTooLong
	}

	var (
		hasUpper   bool
		hasLower   bool
		hasNumber  bool
		hasSpecial bool
	)

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsDigit(char):
			hasNumber = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	if !hasUpper || !hasLower || !hasNumber || !hasSpecial {
		return utils.ErrPasswordSecurity
	}

	return nil
}

func ValidateUserDOB(dob string) error {
	if _, err := time.Parse("2006-01-02", dob); err != nil {
		return utils.ErrDOBFormat
	}
	return nil
}

func ValidatePhoneNumber(phoneNumber string) error {
	phoneRegex := regexp.MustCompile(`^\d{10}$`)
	if !phoneRegex.MatchString(phoneNumber) {
		return utils.ErrInvalidPhoneNumber
	}
	return nil
}

func ValidateOTP(otp string) error {
	if len(otp) == 0 {
		return utils.ErrMissingOTP
	}
	if len(otp) != 6 {
		return utils.ErrOtpLength
	}

	matched, _ := regexp.MatchString(`^\d{6}$`, otp)
	if !matched {
		return utils.ErrOtpNums
	}

	return nil
}

func ValidateUserLoginCredentials(email, password string) error {
	if email == "" || password == "" {
		return utils.ErrLoginCredentialsMissing
	}
	return nil
}
