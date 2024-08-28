package domain

import "time"

type PasswordResetEntry struct {
	ID         int64     `json:"id"`
	Email      string    `json:"email"`
	OTPCode    string    `json:"otp_code"`
	ExpiresAt  time.Time `json:"expires_at"`
	IsVerified bool      `json:"is_verified"`
	CreatedAt  time.Time `json:"created_at"`
}
