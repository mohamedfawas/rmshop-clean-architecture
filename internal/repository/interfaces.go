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

type AdminRepository interface {
	GetByUsername(ctx context.Context, username string) (*domain.Admin, error)
	BlacklistToken(ctx context.Context, token string, expiresAt time.Time) error
}

type CategoryRepository interface {
	Create(ctx context.Context, category *domain.Category) error
	GetByID(ctx context.Context, id int) (*domain.Category, error)
	// Add other methods as needed (GetByID, Update, Delete, etc.)
}

type SubCategoryRepository interface {
	Create(ctx context.Context, subCategory *domain.SubCategory) error
}

type ProductRepository interface {
	Create(ctx context.Context, product *domain.Product) error
}
