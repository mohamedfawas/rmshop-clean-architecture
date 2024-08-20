package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/mohamedfawas/rmshop-clean-architecture/internal/domain"
)

type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error                //fz
	GetByEmail(ctx context.Context, email string) (*domain.User, error) //fz
	UpdateLastLogin(ctx context.Context, userID int64) error
	BlacklistToken(ctx context.Context, token string, expiresAt time.Time) error
	IsTokenBlacklisted(ctx context.Context, token string) (bool, error)
	CreateOTP(ctx context.Context, otp *domain.OTP) error
	CreateVerificationEntry(ctx context.Context, entry *domain.VerificationEntry) error               //fz
	GetVerificationEntryByEmail(ctx context.Context, email string) (*domain.VerificationEntry, error) //fz
	UpdateVerificationEntry(ctx context.Context, entry *domain.VerificationEntry) error               //fz
	DeleteVerificationEntry(ctx context.Context, email string) error
	DeleteExpiredVerificationEntries(ctx context.Context) error
	GetOTPByEmail(ctx context.Context, email string) (*domain.OTP, error)
	UpdateEmailVerificationStatus(ctx context.Context, userID int64, status bool) error
	DeleteOTP(ctx context.Context, email string) error
	GetOTPResendInfo(ctx context.Context, email string) (int, time.Time, error)                       //fz
	UpdateOTPResendInfo(ctx context.Context, email string) error                                      //fz
	UpdateVerificationEntryAfterResendOTP(ctx context.Context, entry *domain.VerificationEntry) error //fz
}

type AdminRepository interface {
	GetByUsername(ctx context.Context, username string) (*domain.Admin, error)   //fz
	BlacklistToken(ctx context.Context, token string, expiresAt time.Time) error //fz
	IsTokenBlacklisted(ctx context.Context, token string) (bool, error)          //fz
}

type CategoryRepository interface {
	Create(ctx context.Context, category *domain.Category) error   //faz 15 aug
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
	GetProductImages(ctx context.Context, productID int64) ([]domain.ProductImage, error)
	UpdatePrimaryImage(ctx context.Context, productID int64, imageID int64) error
	BeginTx(ctx context.Context) (*sql.Tx, error)
	CreateWithTx(ctx context.Context, tx *sql.Tx, product *domain.Product) error
	AddProductImagesWithTx(ctx context.Context, tx *sql.Tx, productID int64, images []domain.ProductImage) ([]int64, error)
	UpdatePrimaryImageWithTx(ctx context.Context, tx *sql.Tx, productID int64, imageID int64) error
	AddProductImages(ctx context.Context, productID int64, images []domain.ProductImage) ([]int64, error)
	GetProductImageByID(ctx context.Context, imageID int64) (*domain.ProductImage, error)
	RemoveProductImage(ctx context.Context, imageID int64) error
}
