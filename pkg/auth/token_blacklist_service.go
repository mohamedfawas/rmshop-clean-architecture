package auth

import (
	"context"
	"database/sql"
	"log"
	"time"
)

type TokenBlacklist struct {
	db *sql.DB
}

func NewTokenBlacklist(db *sql.DB) *TokenBlacklist {
	return &TokenBlacklist{db: db}
}

func (tb *TokenBlacklist) Add(ctx context.Context, token string, expiresAt time.Time) error {
	query := `INSERT INTO blacklisted_tokens (token, expires_at) 
				VALUES ($1,$2)`
	_, err := tb.db.ExecContext(ctx, query, token, expiresAt)
	if err != nil {
		log.Printf("error while blacklisting the token : %v", err)
	}
	return err
}

func (tb *TokenBlacklist) IsBlacklisted(ctx context.Context, token string) (bool, error) {
	query := `SELECT EXISTS (SELECT 1 FROM blacklisted_tokens WHERE token=$1 AND expires_at > NOW())`
	var exists bool
	err := tb.db.QueryRowContext(ctx, query, token).Scan(&exists)
	if err != nil {
		log.Printf("error while checking if the token is blacklisted : %v", err)
	}
	return exists, err
}

func (tb *TokenBlacklist) CleanupExpired(ctx context.Context) error {
	_, err := tb.db.ExecContext(ctx, "DELETE FROM blacklisted_tokens WHERE expires_at <= NOW()")
	return err
}
