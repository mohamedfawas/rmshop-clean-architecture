package domain

import "time"

type PaymentDetails struct {
	OrderID         int64   `json:"order_id"`          // Our internal order ID
	RazorpayOrderID string  `json:"razorpay_order_id"` // Razorpay's order ID
	Amount          float64 `json:"amount"`            // Amount in the smallest currency unit (e.g., paise for INR)
	Currency        string  `json:"currency"`          // Currency code (e.g., "INR")
	RazorpayKeyID   string  `json:"razorpay_key_id"`   // Razorpay API Key ID (public key)
	PaymentURL      string  `json:"payment_url"`       // URL for redirecting to Razorpay payment page
	Status          string  `json:"status"`            // Payment status (e.g., "created", "paid", "failed")
	CreatedAt       int64   `json:"created_at"`        // Timestamp of when the payment was created
	ExpiresAt       int64   `json:"expires_at"`        // Timestamp of when the payment expires
}

type Payment struct {
	ID                int64      `json:"id"`
	OrderID           int64      `json:"order_id"`
	Amount            float64    `json:"amount"`
	PaymentMethod     string     `json:"payment_method"`
	Status            string     `json:"status"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
	ExpiresAt         *time.Time `json:"expires_at,omitempty"`
	RazorpayOrderID   string     `json:"razorpay_order_id,omitempty"`
	RazorpayPaymentID string     `json:"razorpay_payment_id,omitempty"`
	RazorpaySignature string     `json:"razorpay_signature,omitempty"`
}
