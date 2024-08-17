package usecase

import "errors"

var (
	ErrAdminNotFound           = errors.New("admin not found")
	ErrInvalidAdminCredentials = errors.New("invalid admin credentials")
	ErrInvalidAdminToken       = errors.New("invalid admin token")
	ErrDuplicateEmail          = errors.New("email already exists")
	ErrUserNotFound            = errors.New("user not found")
	ErrInvalidCredentials      = errors.New("invalid credentials")
	ErrInvalidToken            = errors.New("invalid token")
	ErrOTPNotFound             = errors.New("OTP not found")
	ErrInvalidOTP              = errors.New("invalid OTP")
	ErrExpiredOTP              = errors.New("OTP has expired")
	ErrEmailAlreadyVerified    = errors.New("email already verified")
	ErrInvalidInput            = errors.New("invalid input")
	ErrDatabaseUnavailable     = errors.New("database unavailable")
	ErrSMTPServerIssue         = errors.New("SMTP server issue")
	ErrNonExEmail              = errors.New("OTP not found for given email") //non existent email
	ErrSignupExpired           = errors.New("signup process has expired")
	ErrTooManyResendAttempts   = errors.New("too many resend attempts")
	ErrUserBlocked             = errors.New("user is blocked")
	ErrTokenAlreadyBlacklisted = errors.New("token already blacklisted")
	ErrInvalidCategory         = errors.New("invalid category ID")
	ErrInvalidSubCategory      = errors.New("invalid sub-category ID")
	ErrProductNotFound         = errors.New("product not found")
	ErrDuplicateImageURL       = errors.New("duplicate image URL")
)
