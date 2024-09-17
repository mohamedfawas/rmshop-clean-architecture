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
	UpdateStock(ctx context.Context, tx *sql.Tx, productID int64, quantityChange int) error
	UpdateStockQuantity(ctx context.Context, productID int64, quantity int) error
	GetProducts(ctx context.Context, params domain.ProductQueryParams) ([]*domain.Product, int64, error)
	GetPublicProductByID(ctx context.Context, id int64) (*domain.PublicProduct, error)
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
	ClearCart(ctx context.Context, userID int64) error
}

type CouponRepository interface {
	Create(ctx context.Context, coupon *domain.Coupon) error
	GetByCode(ctx context.Context, code string) (*domain.Coupon, error)
	IsApplied(ctx context.Context, checkoutID int64) (bool, error)
	GetAllCoupons(ctx context.Context, params domain.CouponQueryParams) ([]*domain.Coupon, int64, error)
	GetByID(ctx context.Context, id int64) (*domain.Coupon, error)
	Update(ctx context.Context, coupon *domain.Coupon) error
}

type CheckoutRepository interface {
	CreateCheckoutSession(ctx context.Context, userID int64) (*domain.CheckoutSession, error)
	AddCheckoutItems(ctx context.Context, sessionID int64, items []*domain.CheckoutItem) error
	GetCartItems(ctx context.Context, userID int64) ([]*domain.CartItemWithProduct, error)
	GetCheckoutByID(ctx context.Context, checkoutID int64) (*domain.CheckoutSession, error)
	GetCheckoutItems(ctx context.Context, checkoutID int64) ([]*domain.CheckoutItem, error)
	UpdateCheckoutAddress(ctx context.Context, checkoutID int64, addressID int64) error
	AddNewAddressToCheckout(ctx context.Context, checkoutID int64, address *domain.UserAddress) error
	GetCheckoutWithItems(ctx context.Context, checkoutID int64) (*domain.CheckoutSummary, error)
	UpdateCheckoutDetails(ctx context.Context, checkout *domain.CheckoutSession) error
	UpdateCheckoutStatus(ctx context.Context, tx *sql.Tx, checkout *domain.CheckoutSession) error
	BeginTx(ctx context.Context) (*sql.Tx, error)
	CreateOrGetShippingAddress(ctx context.Context, userID, addressID int64) (int64, error)
	UpdateCheckoutShippingAddress(ctx context.Context, checkoutID, shippingAddressID int64) error
	GetCheckoutWithAddressByID(ctx context.Context, checkoutID int64) (*domain.CheckoutSession, error)
}

type OrderRepository interface {
	GetByID(ctx context.Context, id int64) (*domain.Order, error)
	GetUserOrders(ctx context.Context, userID int64, page, limit int, sortBy, order, status string) ([]*domain.Order, int64, error)
	UpdateOrderStatus(ctx context.Context, orderID int64, status string) error
	UpdateRefundStatus(ctx context.Context, orderID int64, refundStatus sql.NullString) error
	GetOrders(ctx context.Context, params domain.OrderQueryParams) ([]*domain.Order, int64, error)
	UpdateOrderPaymentStatus(ctx context.Context, orderID int64, status string, paymentID string) error
	GetPaymentByOrderID(ctx context.Context, orderID int64) (*domain.Payment, error)
	UpdatePayment(ctx context.Context, payment *domain.Payment) error
	GetPaymentByRazorpayOrderID(ctx context.Context, razorpayOrderID string) (*domain.Payment, error)
	SetOrderDeliveredAt(ctx context.Context, orderID int64, deliveredAt *time.Time) error
	CreateRefundTx(ctx context.Context, tx *sql.Tx, refund *domain.Refund) error
	UpdateRefundStatusTx(ctx context.Context, tx *sql.Tx, orderID int64, refundStatus sql.NullString) error
	UpdateOrderStatusTx(ctx context.Context, tx *sql.Tx, orderID int64, status string) error
	CreateReturnRequestTx(ctx context.Context, tx *sql.Tx, returnRequest *domain.ReturnRequest) error
	BeginTx(ctx context.Context) (*sql.Tx, error)
	CreateOrder(ctx context.Context, tx *sql.Tx, order *domain.Order) (int64, error)
	CreatePayment(ctx context.Context, tx *sql.Tx, payment *domain.Payment) error
	AddOrderItem(ctx context.Context, tx *sql.Tx, item *domain.OrderItem) error
	UpdateOrderRazorpayID(ctx context.Context, orderID int64, razorpayOrderID string) error
	UpdateOrderHasReturnRequestTx(ctx context.Context, tx *sql.Tx, orderID int64, hasReturnRequest bool) error
	GetOrderWithItems(ctx context.Context, orderID int64) (*domain.Order, error)
}

type InventoryRepository interface {
	GetInventory(ctx context.Context, params domain.InventoryQueryParams) ([]*domain.InventoryItem, int64, error)
}

type WishlistRepository interface {
	AddItem(ctx context.Context, item *domain.WishlistItem) error
	ItemExists(ctx context.Context, userID, productID int64) (bool, error)
	GetWishlistItemCount(ctx context.Context, userID int64) (int, error)
	RemoveItem(ctx context.Context, userID, productID int64) error
	GetUserWishlistItems(ctx context.Context, userID int64, page, limit int, sortBy, order string) ([]*domain.WishlistItem, int64, error)
}

type WalletRepository interface {
	GetByUserID(ctx context.Context, userID int64) (*domain.Wallet, error)
}

type SalesRepository interface {
	GetSalesData(ctx context.Context, startDate, endDate time.Time, couponApplied bool) ([]domain.DailySalesData, error)
	GetTopSellingProducts(ctx context.Context, startDate, endDate time.Time, limit int) ([]domain.TopSellingProduct, error)
}

type AnalyticsRepository interface {
	GetTopProducts(ctx context.Context, startDate, endDate time.Time, limit int, sortBy string) ([]domain.TopProduct, error)
}
