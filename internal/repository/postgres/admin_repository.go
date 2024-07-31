package postgres

import (
	"context"
	"database/sql"
	"time"

	"github.com/mohamedfawas/rmshop-clean-architecture/internal/domain"
	"github.com/mohamedfawas/rmshop-clean-architecture/internal/usecase"
)

type adminRepository struct {
	db *sql.DB
}

func NewAdminRepository(db *sql.DB) *adminRepository {
	return &adminRepository{db: db}
}

func (r *adminRepository) GetByUsername(ctx context.Context, username string) (*domain.Admin, error) {
	query := `SELECT id, username, password_hash, created_at, updated_at FROM admins WHERE username = $1`

	var admin domain.Admin
	err := r.db.QueryRowContext(ctx, query, username).Scan(
		&admin.ID, &admin.Username, &admin.PasswordHash, &admin.CreatedAt, &admin.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, usecase.ErrAdminNotFound
		}
		return nil, err
	}

	return &admin, nil
}

func (r *adminRepository) BlacklistToken(ctx context.Context, token string, expiresAt time.Time) error {
	query := `INSERT INTO blacklisted_tokens (token, expires_at) VALUES ($1, $2)`
	_, err := r.db.ExecContext(ctx, query, token, expiresAt)
	return err
}
