package domain

import (
	"time"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID           int64     `json:"id"`
	Name         string    `json:"name"`
	Email        string    `json:"email"`
	Password     string    `json:"-"` // Password is not exposed in JSON,never included in JSON output. // This field is used for password input, not storage
	PasswordHash string    `json:"-"` // This is what gets stored in the database //never included in JSON output.
	DOB          time.Time `json:"date_of_birth"`
	PhoneNumber  string    `json:"phone_number"`
	IsBlocked    bool      `json:"is_blocked"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	LastLogin    time.Time `json:"last_login,omitempty"` //omitempty option in a JSON tag tells the JSON encoder to omit the field if it has an empty value
}

func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password))
	return err == nil //if passwords match true is returned
}

type BlacklistedToken struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
}
