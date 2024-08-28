package domain

import (
	"time"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID              int64      `json:"id"`
	Name            string     `json:"name"`
	Email           string     `json:"email"`
	Password        string     `json:"-"` // Password is not exposed in JSON,never included in JSON output. // This field is used for password input, not storage
	PasswordHash    string     `json:"-"` // This is what gets stored in the database //never included in JSON output.
	DOB             time.Time  `json:"date_of_birth"`
	PhoneNumber     string     `json:"phone_number"`
	IsBlocked       bool       `json:"is_blocked"`
	IsEmailVerified bool       `json:"is_email_verified"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
	DeletedAt       *time.Time `json:"deleted_at,omitempty"`
	LastLogin       time.Time  `json:"last_login,omitempty"` //omitempty option in a JSON tag tells the JSON encoder to omit the field if it has an empty value
}

type UserUpdatedData struct {
	Name        string `json:"name"`
	PhoneNumber string `json:"phone_number"`
}

type BlacklistedToken struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
}

func (u *User) CheckPassword(password string) bool {
	// Compare the provided password with the stored password hash
	err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password))
	return err == nil // Returns true if passwords match, false otherwise
}
