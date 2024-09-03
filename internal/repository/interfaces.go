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
	GetUserAddressByID(ctx context.Context, addressID int64) (*domain.UserAddress, error)
	UpdateUserAddress(ctx context.Context, address *domain.UserAddress) error
	GetUserAddresses(ctx context.Context, userID int64) ([]*domain.UserAddress, error)
	GetUserAddressCount(ctx context.Context, userID int64) (int, error)
	DeleteUserAddress(ctx context.Context, addressID int64) error
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

type CartRepository interface {
	AddCartItem(ctx context.Context, item *domain.CartItem) error
	GetCartItemByProductID(ctx context.Context, userID, productID int64) (*domain.CartItem, error)
	UpdateCartItem(ctx context.Context, item *domain.CartItem) error
	GetCartByUserID(ctx context.Context, userID int64) ([]*domain.CartItemWithProduct, error)
	UpdateCartItemQuantity(ctx context.Context, userID, itemID int64, quantity int) error
	GetCartItemByID(ctx context.Context, itemID int64) (*domain.CartItem, error)
	DeleteCartItem(ctx context.Context, itemID int64) error
	GetCartTotal(ctx context.Context, userID int64) (float64, error)
	ApplyCoupon(ctx context.Context, userID int64, coupon *domain.Coupon) error
	RemoveCoupon(ctx context.Context, userID int64) error
	GetAppliedCoupon(ctx context.Context, userID int64) (*domain.Coupon, error)
}

type CouponRepository interface {
	Create(ctx context.Context, coupon *domain.Coupon) error
	GetByCode(ctx context.Context, code string) (*domain.Coupon, error)
	IsApplied(ctx context.Context, checkoutID int64) (bool, error)
}

type CheckoutRepository interface {
	CreateCheckoutSession(ctx context.Context, userID int64) (*domain.CheckoutSession, error)
	AddCheckoutItems(ctx context.Context, sessionID int64, items []*domain.CheckoutItem) error
	GetCartItems(ctx context.Context, userID int64) ([]*domain.CartItemWithProduct, error)
	GetCheckoutByID(ctx context.Context, checkoutID int64) (*domain.CheckoutSession, error)
	UpdateCheckout(ctx context.Context, checkout *domain.CheckoutSession) error
	GetCheckoutItems(ctx context.Context, checkoutID int64) ([]*domain.CheckoutItem, error)
}
