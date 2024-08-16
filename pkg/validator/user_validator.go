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

// // IsValidName checks if the name is not empty
// func IsValidName(name string) bool {
// 	return len(name) > 0
// }

// // IsValidEmail checks if the email follows a valid format
// func IsValidEmail(email string) bool {
// 	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
// 	return emailRegex.MatchString(email)
// }

// // IsValidPassword checks if the password is at least 8 characters utils.
// func IsValidPassword(password string) bool {
// 	return len(password) >= 8
// }

// // IsValidPhoneNumber checks if the phone number is exactly 10 digits
// func IsValidPhoneNumber(phoneNumber string) bool {
// 	phoneRegex := regexp.MustCompile(`^\d{10}$`)
// 	return phoneRegex.MatchString(phoneNumber)
// }

// // IsValidDateOfBirth checks if the date of birth is in the correct format (YYYY-MM-DD)
// func IsValidDateOfBirth(dob string) bool {
// 	_, err := time.Parse("2006-01-02", dob)
// 	return err == nil
// }

// // ValidateUserInput checks all user input fields
// func ValidateUserInput(name, email, password, phoneNumber, dob string) error {

// 	//trim the trailing and leading whitespace
// 	name = strings.TrimSpace(name)
// 	email = strings.TrimSpace(email)
// 	phoneNumber = strings.TrimSpace(phoneNumber)
// 	dob = strings.TrimSpace(dob)

// 	if !IsValidName(name) {
// 		return fmt.Errorf("invalid name")
// 	}
// 	if !IsValidEmail(email) {
// 		return fmt.Errorf("invalid email format")
// 	}
// 	if !IsValidPassword(password) {
// 		return fmt.Errorf("password must be at least 6 characters long")
// 	}
// 	if !IsValidPhoneNumber(phoneNumber) {
// 		return fmt.Errorf("invalid phone number format")
// 	}
// 	if !IsValidDateOfBirth(dob) {
// 		return fmt.Errorf("invalid date of birth format (use YYYY-MM-DD)")
// 	}
// 	return nil
// }

// func ParseDateOfBirth(dob string) time.Time {
// 	dob, err := time.Parse("2006-01-02", dob)
// 	if err != nil {
// 		log.Printf("Error parsing date of birth: %v", err)
// 		return dob
// 	}
// }
