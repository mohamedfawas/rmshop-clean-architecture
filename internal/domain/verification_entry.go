package domain

import (
	"time"
)

type VerificationEntry struct {
	ID           int64     `json:"id"`
	Email        string    `json:"email"`
	OTPCode      string    `json:"otp_code"`
	UserData     *User     `json:"user_data"`
	PasswordHash string    `json:"password_hash"`
	ExpiresAt    time.Time `json:"expires_at"`
	IsVerified   bool      `json:"is_verified"`
	CreatedAt    time.Time `json:"created_at"`
}
