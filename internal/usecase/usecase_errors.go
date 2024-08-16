package usecase

import "errors"

var (
	ErrDuplicateEmail        = errors.New("email already exists")
	ErrUserNotFound          = errors.New("user not found")
	ErrInvalidCredentials    = errors.New("invalid credentials")
	ErrInvalidToken          = errors.New("invalid token")
	ErrOTPNotFound           = errors.New("OTP not found")
	ErrInvalidOTP            = errors.New("invalid OTP")
	ErrExpiredOTP            = errors.New("OTP has expired")
	ErrEmailAlreadyVerified  = errors.New("email already verified")
	ErrInvalidInput          = errors.New("invalid input")
	ErrDatabaseUnavailable   = errors.New("database unavailable")
	ErrSMTPServerIssue       = errors.New("SMTP server issue")
	ErrNonExEmail            = errors.New("OTP not found for given email") //non existent email
	ErrSignupExpired         = errors.New("signup process has expired")
	ErrTooManyResendAttempts = errors.New("too many resend attempts")
)
