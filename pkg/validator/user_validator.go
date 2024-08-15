package validator

import (
	"fmt"
	"regexp"
	"time"
)

// IsValidName checks if the name is not empty
func IsValidName(name string) bool {
	return len(name) > 0
}

// IsValidEmail checks if the email follows a valid format
func IsValidEmail(email string) bool {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}

// IsValidPassword checks if the password is at least 6 characters long
func IsValidPassword(password string) bool {
	return len(password) >= 6
}

// IsValidPhoneNumber checks if the phone number is exactly 10 digits
func IsValidPhoneNumber(phoneNumber string) bool {
	phoneRegex := regexp.MustCompile(`^\d{10}$`)
	return phoneRegex.MatchString(phoneNumber)
}

// IsValidDateOfBirth checks if the date of birth is in the correct format (YYYY-MM-DD)
func IsValidDateOfBirth(dob string) bool {
	_, err := time.Parse("2006-01-02", dob)
	return err == nil
}

// ValidateUserInput checks all user input fields
func ValidateUserInput(name, email, password, phoneNumber, dob string) error {
	if !IsValidName(name) {
		return fmt.Errorf("invalid name")
	}
	if !IsValidEmail(email) {
		return fmt.Errorf("invalid email format")
	}
	if !IsValidPassword(password) {
		return fmt.Errorf("password must be at least 6 characters long")
	}
	if !IsValidPhoneNumber(phoneNumber) {
		return fmt.Errorf("invalid phone number format")
	}
	if !IsValidDateOfBirth(dob) {
		return fmt.Errorf("invalid date of birth format (use YYYY-MM-DD)")
	}
	return nil
}
