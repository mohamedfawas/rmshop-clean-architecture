package repository

import (
	"context"
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
	CreateUserSignUpVerifcationEntry(ctx context.Context, entry *domain.VerificationEntry) error //fz
	UpdateSignUpVerificationEntry(ctx context.Context, entry *domain.VerificationEntry) error    //fz
	DeleteSignUpVerificationEntry(ctx context.Context, email string) error
	DeleteExpiredVerificationEntries(ctx context.Context) error
	GetOTPByEmail(ctx context.Context, email string) (*domain.OTP, error)
	UpdateEmailVerificationStatus(ctx context.Context, userID int64, status bool) error
	DeleteOTP(ctx context.Context, email string) error
	GetOTPResendInfo(ctx context.Context, email string) (int, time.Time, error)                             //fz
	UpdateOTPResendInfo(ctx context.Context, email string) error                                            //fz
	UpdateSignUpVerificationEntryAfterResendOTP(ctx context.Context, entry *domain.VerificationEntry) error //fz
	GetByID(ctx context.Context, id int64) (*domain.User, error)                                            //fz
	Update(ctx context.Context, user *domain.User) error                                                    //fz
	UpdatePassword(ctx context.Context, userID int64, newPasswordHash string) error                         //fz
	CreatePasswordResetEntry(ctx context.Context, entry *domain.PasswordResetEntry) error                   //fz
	FindSignUpVerificationEntryByEmail(ctx context.Context, email string) (*domain.VerificationEntry, error)
	FindPasswordResetEntryByEmail(ctx context.Context, email string) (*domain.PasswordResetEntry, error)
	DeletePasswordResetVerificationEntry(ctx context.Context, email string) error
	UserAddressExists(ctx context.Context, address *domain.UserAddress) (bool, error)
	AddUserAddress(ctx context.Context, address *domain.UserAddress) error
}

type AdminRepository interface {
	GetByUsername(ctx context.Context, username string) (*domain.Admin, error)   //fz
	BlacklistToken(ctx context.Context, token string, expiresAt time.Time) error //fz
	IsTokenBlacklisted(ctx context.Context, token string) (bool, error)          //fz
}

type CategoryRepository interface {
	Create(ctx context.Context, category *domain.Category) error   //fz
	GetByID(ctx context.Context, id int) (*domain.Category, error) // fz
	GetAll(ctx context.Context) ([]*domain.Category, error)
	GetActiveByID(ctx context.Context, id int) (*domain.Category, error) // New method for retrieving active categories
	Update(ctx context.Context, category *domain.Category) error
	SoftDelete(ctx context.Context, id int) error
	// Add other methods as needed (GetByID, Update, Delete, etc.)
}

type SubCategoryRepository interface {
	Create(ctx context.Context, subCategory *domain.SubCategory) error //fz
	GetByCategoryID(ctx context.Context, categoryID int) ([]*domain.SubCategory, error)
	GetByID(ctx context.Context, id int) (*domain.SubCategory, error) //fz
	Update(ctx context.Context, subCategory *domain.SubCategory) error
	SoftDelete(ctx context.Context, id int) error
}

type ProductRepository interface {
	Create(ctx context.Context, product *domain.Product) error                              //fz
	SlugExists(ctx context.Context, slug string) (bool, error)                              //fz
	NameExists(ctx context.Context, name string) (bool, error)                              //fz
	Update(ctx context.Context, product *domain.Product) error                              //fz
	SoftDelete(ctx context.Context, id int64) error                                         //fz
	GetByID(ctx context.Context, id int64) (*domain.Product, error)                         //fz
	NameExistsBeforeUpdate(ctx context.Context, name string, excludeID int64) (bool, error) //fz
	AddImage(ctx context.Context, productID int64, imageURL string, isPrimary bool) error   //fz
	GetImageCount(ctx context.Context, productID int64) (int, error)                        //fz
	DeleteImage(ctx context.Context, productID int64, imageURL string) error                //fz
	GetProductImages(ctx context.Context, productID int64) ([]*domain.ProductImage, error)  //fz
	SetImageAsPrimary(ctx context.Context, productID int64, imageID int64) error            //fz
	UpdateProductPrimaryImage(ctx context.Context, productID int64, imageID *int64) error
	GetImageByURL(ctx context.Context, productID int64, imageURL string) (*domain.ProductImage, error)
	GetPrimaryImage(ctx context.Context, productID int64) (*domain.ProductImage, error)
	UpdateImagePrimary(ctx context.Context, imageID int64, isPrimary bool) error
	GetImageByID(ctx context.Context, imageID int64) (*domain.ProductImage, error) //fz
	DeleteImageByID(ctx context.Context, imageID int64) error                      //fz
	GetAll(ctx context.Context) ([]*domain.Product, error)                         //fz
}
