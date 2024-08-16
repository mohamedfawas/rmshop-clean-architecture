package repository

import "errors"

var (
	ErrUserNotFound              = errors.New("user not found")
	ErrOTPNotFound               = errors.New("OTP not found")
	ErrDuplicateEmail            = errors.New("email already exists")
	ErrVerificationEntryNotFound = errors.New("verification entry not found")
)
