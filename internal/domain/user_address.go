package domain

import "time"

type UserAddress struct {
	ID           int64      `json:"id"`
	UserID       int64      `json:"user_id"`
	AddressLine1 string     `json:"address_line1"`
	AddressLine2 string     `json:"address_line2"`
	State        string     `json:"state"`
	City         string     `json:"city"`
	PinCode      string     `json:"pincode"`
	Landmark     string     `json:"landmark"`
	PhoneNumber  string     `json:"phone_number"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	DeletedAt    *time.Time `json:"deleted_at,omitempty"`
}
