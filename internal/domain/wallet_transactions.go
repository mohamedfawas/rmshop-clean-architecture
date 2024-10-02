package domain

import (
	"time"
)

type WalletTransaction struct {
	ID              int64     `json:"id"`
	UserID          int64     `json:"user_id"`
	Amount          float64   `json:"amount"`
	TransactionType string    `json:"transaction_type"`
	ReferenceID     *int64    `json:"reference_id,omitempty"`
	ReferenceType   string    `json:"reference_type,omitempty"`
	BalanceAfter    float64   `json:"balance_after"`
	CreatedAt       time.Time `json:"created_at"`
}
