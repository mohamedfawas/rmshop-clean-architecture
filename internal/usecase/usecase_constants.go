package usecase

import "time"

const (
	maxResendAttempts   = 3
	resendCooldown      = 1 * time.Minute
	signupExpiration    = 1 * time.Hour
	MaxImagesPerProduct = 10
)
