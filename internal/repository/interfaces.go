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
	CreateOTP(ctx context.Context, otp *domain.OTP) error
	GetOTPByEmail(ctx context.Context, email string) (*domain.OTP, error)
	DeleteOTP(ctx context.Context, email string) error
	UpdateEmailVerificationStatus(ctx context.Context, userID int64, status bool) error
}

type AdminRepository interface {
	GetByUsername(ctx context.Context, username string) (*domain.Admin, error)
	BlacklistToken(ctx context.Context, token string, expiresAt time.Time) error
}

type CategoryRepository interface {
	Create(ctx context.Context, category *domain.Category) error
	GetByID(ctx context.Context, id int) (*domain.Category, error) // retrieving any category by ID
	GetAll(ctx context.Context) ([]*domain.Category, error)
	GetActiveByID(ctx context.Context, id int) (*domain.Category, error) // New method for retrieving active categories
	Update(ctx context.Context, category *domain.Category) error
	SoftDelete(ctx context.Context, id int) error
	// Add other methods as needed (GetByID, Update, Delete, etc.)
}

type SubCategoryRepository interface {
	Create(ctx context.Context, subCategory *domain.SubCategory) error
	GetByCategoryID(ctx context.Context, categoryID int) ([]*domain.SubCategory, error)
	GetByID(ctx context.Context, id int) (*domain.SubCategory, error)
	Update(ctx context.Context, subCategory *domain.SubCategory) error
	SoftDelete(ctx context.Context, id int) error
}

type ProductRepository interface {
	Create(ctx context.Context, product *domain.Product) error
	GetAll(ctx context.Context) ([]*domain.Product, error)
	GetByID(ctx context.Context, id int64) (*domain.Product, error)
	Update(ctx context.Context, product *domain.Product) error
	SoftDelete(ctx context.Context, id int64) error
	GetActiveProducts(ctx context.Context, page, pageSize int) ([]*domain.Product, int, error)
}
