package postgres

import (
	"context"
	"database/sql"
	"log"
	"time"

	"github.com/mohamedfawas/rmshop-clean-architecture/internal/domain"
	"github.com/mohamedfawas/rmshop-clean-architecture/pkg/utils"
)

type adminRepository struct {
	db *sql.DB
}

func NewAdminRepository(db *sql.DB) *adminRepository {
	return &adminRepository{db: db}
}

func (r *adminRepository) GetByUsername(ctx context.Context, username string) (*domain.Admin, error) {
	// SQL query to select admin details by username
	query := `SELECT id, username, password_hash, created_at, updated_at FROM admins WHERE username = $1`

	var admin domain.Admin
	// Execute the query and scan the result into the admin struct
	err := r.db.QueryRowContext(ctx, query, username).Scan(
		&admin.ID, &admin.Username, &admin.PasswordHash, &admin.CreatedAt, &admin.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, utils.ErrAdminNotFound
		}
		log.Printf("error while retrieving admin details : %v", err)
		return nil, err
	}

	// Return the admin struct and nil error
	return &admin, nil
}

func (r *adminRepository) BlacklistToken(ctx context.Context, token string, expiresAt time.Time) error {
	query := `INSERT INTO blacklisted_tokens (token, expires_at) VALUES ($1, $2)`
	_, err := r.db.ExecContext(ctx, query, token, expiresAt)
	if err != nil {
		log.Printf("error while adding the token to blacklisted_tokens : %v", err)
	}
	return err
}

func (r *adminRepository) IsTokenBlacklisted(ctx context.Context, token string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM blacklisted_tokens WHERE token = $1)`
	var exists bool
	err := r.db.QueryRowContext(ctx, query, token).Scan(&exists)
	if err != nil {
		log.Printf("error while checking if admin token is balcklisted : %v", err)
	}
	return exists, err
}
