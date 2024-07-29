package repository

import (
	"context"
	"time"

	"github.com/mohamedfawas/rmshop-clean-architecture/internal/domain"
)

type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error                //Create user method
	GetByEmail(ctx context.Context, email string) (*domain.User, error) //Get user details using email
	UpdateLastLogin(ctx context.Context, userID int64) error            //ecord the most recent time a user successfully authenticated (logged in) to the system
	BlacklistToken(ctx context.Context, token string, expiresAt time.Time) error
	IsTokenBlacklisted(ctx context.Context, token string) (bool, error)
}
