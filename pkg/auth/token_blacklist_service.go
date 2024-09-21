package auth

import (
	"context"
	"database/sql"
	"log"
	"time"
)

// TokenBlacklist represents the structure for interacting with the blacklisted tokens in the database.
// It provides methods to add tokens to the blacklist, check if a token is blacklisted,
// and clean up expired tokens.
type TokenBlacklist struct {
	db *sql.DB
}

// NewTokenBlacklist creates a new instance of TokenBlacklist.
// It takes a *sql.DB object as input, representing the database connection pool,
// and returns a pointer to TokenBlacklist.
//
// Parameters:
//   - db: a pointer to the database connection object.
//
// Returns:
//   - *TokenBlacklist: a pointer to the newly created TokenBlacklist structure.
func NewTokenBlacklist(db *sql.DB) *TokenBlacklist {
	return &TokenBlacklist{db: db}
}

// Add inserts a token into the blacklisted_tokens table with an expiration time.
// The token will be considered blacklisted until its expiration time.
//
// Parameters:
//   - ctx: a context object to control the query timeout or cancellation.
//   - token: the token string to be blacklisted.
//   - expiresAt: the time when the token will expire and no longer be blacklisted.
//
// Returns:
//   - error: an error object if there was an issue adding the token to the blacklist, or nil if successful.
func (tb *TokenBlacklist) Add(ctx context.Context, token string, expiresAt time.Time) error {
	query := `INSERT INTO blacklisted_tokens (token, expires_at) 
				VALUES ($1,$2)`
	_, err := tb.db.ExecContext(ctx, query, token, expiresAt)
	if err != nil {
		log.Printf("error while blacklisting the token : %v", err)
	}
	return err
}

// IsBlacklisted checks if a token is currently blacklisted and unexpired.
//
// Parameters:
//   - ctx: a context object to control the query timeout or cancellation.
//   - token: the token string to be checked.
//
// Returns:
//   - bool: true if the token is blacklisted and unexpired, false otherwise.
//   - error: an error object if there was an issue querying the database, or nil if successful.
func (tb *TokenBlacklist) IsBlacklisted(ctx context.Context, token string) (bool, error) {
	query := `SELECT EXISTS (SELECT 1 FROM blacklisted_tokens WHERE token=$1 AND expires_at > NOW())`
	var exists bool
	err := tb.db.QueryRowContext(ctx, query, token).Scan(&exists)
	if err != nil {
		log.Printf("error while checking if the token is blacklisted : %v", err)
	}
	return exists, err
}

// CleanupExpired removes expired tokens from the blacklisted_tokens table.
// This function is meant to be run periodically to ensure the blacklist only contains valid, unexpired tokens.
//
// Parameters:
//   - ctx: a context object to control the query timeout or cancellation.
//
// Returns:
//   - error: an error object if there was an issue deleting expired tokens, or nil if successful.
func (tb *TokenBlacklist) CleanupExpired(ctx context.Context) error {
	_, err := tb.db.ExecContext(ctx, "DELETE FROM blacklisted_tokens WHERE expires_at <= NOW()")
	if err != nil {
		log.Printf("error while clean up expired tokens : %v", err)
	}
	return err
}
