package domain

import "time"

type ShippingAddress struct {
	ID           int64     `json:"id"`
	UserID       int64     `json:"user_id,omitempty"`
	AddressID    int64     `json:"address_id,omitempty"`
	AddressLine1 string    `json:"address_line1"`
	AddressLine2 string    `json:"address_line2,omitempty"`
	City         string    `json:"city"`
	State        string    `json:"state"`
	Landmark     string    `json:"landmark"`
	PinCode      string    `json:"pincode"`
	PhoneNumber  string    `json:"phone_number"`
	CreatedAt    time.Time `json:"created_at,omitempty"`
}

type ShippingAddressResponseInCheckoutSummary struct {
	ID           int64  `json:"id"`
	AddressID    int64  `json:"address_id,omitempty"`
	AddressLine1 string `json:"address_line1"`
	AddressLine2 string `json:"address_line2,omitempty"`
	City         string `json:"city"`
	State        string `json:"state"`
	Landmark     string `json:"landmark"`
	PinCode      string `json:"pincode"`
	PhoneNumber  string `json:"phone_number"`
}
