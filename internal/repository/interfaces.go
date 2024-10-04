package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/mohamedfawas/rmshop-clean-architecture/internal/domain"
)

type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	UpdateLastLogin(ctx context.Context, userID int64) error
	BlacklistToken(ctx context.Context, token string, expiresAt time.Time) error
	IsTokenBlacklisted(ctx context.Context, token string) (bool, error)
	CreateOTP(ctx context.Context, otp *domain.OTP) error
	CreateUserSignUpVerifcationEntry(ctx context.Context, entry *domain.VerificationEntry) error
	UpdateSignUpVerificationEntry(ctx context.Context, entry *domain.VerificationEntry) error
	DeleteSignUpVerificationEntry(ctx context.Context, email string) error
	DeleteExpiredVerificationEntries(ctx context.Context) error
	GetOTPByEmail(ctx context.Context, email string) (*domain.OTP, error)
	UpdateEmailVerificationStatus(ctx context.Context, userID int64, status bool) error
	DeleteOTP(ctx context.Context, email string) error
	GetOTPResendInfo(ctx context.Context, email string) (int, time.Time, error)
	UpdateOTPResendInfo(ctx context.Context, email string) error
	UpdateSignUpVerificationEntryAfterResendOTP(ctx context.Context, entry *domain.VerificationEntry) error
	GetByID(ctx context.Context, id int64) (*domain.User, error)
	Update(ctx context.Context, user *domain.User) error
	UpdatePassword(ctx context.Context, userID int64, newPasswordHash string) error
	CreatePasswordResetEntry(ctx context.Context, entry *domain.PasswordResetEntry) error
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
	GetByUsername(ctx context.Context, username string) (*domain.Admin, error)
	BlacklistToken(ctx context.Context, token string, expiresAt time.Time) error
	IsTokenBlacklisted(ctx context.Context, token string) (bool, error)
}

type CategoryRepository interface {
	Create(ctx context.Context, category *domain.Category) error
	GetByID(ctx context.Context, id int) (*domain.Category, error)
	GetAll(ctx context.Context) ([]*domain.Category, error)
	Update(ctx context.Context, category *domain.Category) error
	SoftDelete(ctx context.Context, id int) error
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
	SlugExists(ctx context.Context, slug string) (bool, error)
	NameExists(ctx context.Context, name string) (bool, error)
	Update(ctx context.Context, product *domain.Product) error
	SoftDelete(ctx context.Context, id int64) error
	GetByID(ctx context.Context, id int64) (*domain.Product, error)
	NameExistsBeforeUpdate(ctx context.Context, name string, excludeID int64) (bool, error)
	AddImage(ctx context.Context, productID int64, imageURL string, isPrimary bool) error
	GetImageCount(ctx context.Context, productID int64) (int, error)
	DeleteImage(ctx context.Context, productID int64, imageURL string) error
	GetProductImages(ctx context.Context, productID int64) ([]*domain.ProductImage, error)
	SetImageAsPrimary(ctx context.Context, productID int64, imageID int64) error
	UpdateProductPrimaryImage(ctx context.Context, productID int64, imageID *int64) error
	GetImageByURL(ctx context.Context, productID int64, imageURL string) (*domain.ProductImage, error)
	GetPrimaryImage(ctx context.Context, productID int64) (*domain.ProductImage, error)
	UpdateImagePrimary(ctx context.Context, imageID int64, isPrimary bool) error
	GetImageByID(ctx context.Context, imageID int64) (*domain.ProductImage, error)
	DeleteImageByID(ctx context.Context, imageID int64) error
	GetAll(ctx context.Context) ([]*domain.Product, error)
	UpdateStockQuantity(ctx context.Context, productID int64, quantity int) error
	GetProducts(ctx context.Context, params domain.ProductQueryParams) ([]*domain.Product, int64, error)
	GetPublicProductByID(ctx context.Context, id int64) (*domain.PublicProduct, error)
	UpdateStockTx(ctx context.Context, tx *sql.Tx, productID int64, quantity int) error
}

type CartRepository interface {
	AddCartItem(ctx context.Context, item *domain.CartItem) error
	GetCartItemByProductID(ctx context.Context, userID, productID int64) (*domain.CartItem, error)
	UpdateCartItem(ctx context.Context, item *domain.CartItem) error
	GetCartByUserID(ctx context.Context, userID int64) ([]*domain.CartItem, error)
	UpdateCartItemQuantity(ctx context.Context, cartItem *domain.CartItem) error
	GetCartItemByID(ctx context.Context, itemID int64) (*domain.CartItem, error)
	DeleteCartItem(ctx context.Context, itemID int64) error
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
	GetCartItems(ctx context.Context, userID int64) ([]*domain.CartItemWithProduct, error)
	GetCheckoutByID(ctx context.Context, checkoutID int64) (*domain.CheckoutSession, error)
	UpdateCheckoutDetails(ctx context.Context, checkout *domain.CheckoutSession) error
	UpdateCheckoutStatus(ctx context.Context, tx *sql.Tx, checkout *domain.CheckoutSession) error
	BeginTx(ctx context.Context) (*sql.Tx, error)
	CreateOrGetShippingAddress(ctx context.Context, userID, addressID int64) (int64, error)
	UpdateCheckoutShippingAddress(ctx context.Context, checkoutID, shippingAddressID int64) error
	GetShippingAddress(ctx context.Context, addressID int64) (*domain.ShippingAddress, error)
	GetOrCreateCheckoutSession(ctx context.Context, userID int64) (*domain.CheckoutSession, error)
	GetCheckoutSession(ctx context.Context, userID int64) (*domain.CheckoutSession, error)
	MarkCheckoutAsDeleted(ctx context.Context, tx *sql.Tx, checkoutID int64) error
}

type OrderRepository interface {
	GetByID(ctx context.Context, id int64) (*domain.Order, error)
	GetUserOrders(ctx context.Context, userID int64, page int) ([]*domain.Order, int64, error)
	UpdateOrderStatus(ctx context.Context, orderID int64, status string) error
	GetOrders(ctx context.Context, limit, offset int) ([]*domain.Order, int64, error)
	GetPaymentByOrderID(ctx context.Context, orderID int64) (*domain.Payment, error)
	UpdatePayment(ctx context.Context, payment *domain.Payment) error
	UpdateOrderStatusTx(ctx context.Context, tx *sql.Tx, orderID int64, status string) error
	CreateReturnRequestTx(ctx context.Context, tx *sql.Tx, returnRequest *domain.ReturnRequest) error
	BeginTx(ctx context.Context) (*sql.Tx, error)
	CreateOrder(ctx context.Context, tx *sql.Tx, order *domain.Order) (int64, error)
	CreatePayment(ctx context.Context, tx *sql.Tx, payment *domain.Payment) error
	AddOrderItem(ctx context.Context, tx *sql.Tx, item *domain.OrderItem) error
	UpdateOrderRazorpayID(ctx context.Context, orderID int64, razorpayOrderID string) error
	UpdateOrderHasReturnRequestTx(ctx context.Context, tx *sql.Tx, orderID int64, hasReturnRequest bool) error
	UpdateOrderHasReturnRequest(ctx context.Context, orderID int64, hasReturnRequest bool) error
	UpdateOrderDeliveryStatus(ctx context.Context, tx *sql.Tx, orderID int64, deliveryStatus, orderStatus string, deliveredAt *time.Time) error
	IsOrderDelivered(ctx context.Context, orderID int64) (bool, error)
	GetOrderByID(ctx context.Context, id int64) (*domain.Order, error)
	GetByIDTx(ctx context.Context, tx *sql.Tx, id int64) (*domain.Order, error)
	GetOrderItemsTx(ctx context.Context, tx *sql.Tx, orderID int64) ([]*domain.OrderItem, error)
	GetCancellationRequests(ctx context.Context, params domain.CancellationRequestParams) ([]*domain.CancellationRequest, int64, error)
	GetOrderDetails(ctx context.Context, id int64) (*domain.Order, error)
	GetOrderItems(ctx context.Context, orderID int64) ([]domain.OrderItem, error)
	GetShippingAddress(ctx context.Context, addressID int64) (*domain.ShippingAddress, error)
	UpdateOrderStatusAndSetCancelledTx(ctx context.Context, tx *sql.Tx, orderID int64, order_status, delivery_status string, isCancelled bool) error
	CreateCancellationRequestTx(ctx context.Context, tx *sql.Tx, request *domain.CancellationRequest) error
	UpdateCancellationRequestTx(ctx context.Context, tx *sql.Tx, request *domain.CancellationRequest) error
	GetCancellationRequestByOrderIDTx(ctx context.Context, tx *sql.Tx, orderID int64) (*domain.CancellationRequest, error)
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
	AddBalance(ctx context.Context, tx *sql.Tx, userID int64, amount float64) error
	CreateTransaction(ctx context.Context, tx *sql.Tx, transaction *domain.WalletTransaction) error
	GetTransactions(ctx context.Context, userID int64, page, limit int, sort, order, transactionType string) ([]*domain.WalletTransaction, int64, error)
	CreateTransactionTx(ctx context.Context, tx *sql.Tx, transaction *domain.WalletTransaction) error
	UpdateBalanceTx(ctx context.Context, tx *sql.Tx, userID int64, amount float64) error
	GetWalletTx(ctx context.Context, tx *sql.Tx, userID int64) (*domain.Wallet, error)
	UpdateWalletBalanceTx(ctx context.Context, tx *sql.Tx, userID int64, newBalance float64) error
	CreateWalletTransactionTx(ctx context.Context, tx *sql.Tx, transaction *domain.WalletTransaction) error
	CreateWalletTx(ctx context.Context, tx *sql.Tx, wallet *domain.Wallet) error
}

type SalesRepository interface {
	GetDailySalesData(ctx context.Context, date time.Time) ([]domain.DailySales, error)
	GetWeeklySalesData(ctx context.Context, startDate time.Time) ([]domain.DailySales, error)
	GetMonthlySalesData(ctx context.Context, year int, month time.Month) ([]domain.DailySales, error)
	GetCustomSalesData(ctx context.Context, startDate, endDate time.Time) ([]domain.DailySales, error)
}

type AnalyticsRepository interface {
	GetTopProducts(ctx context.Context, startDate, endDate time.Time, limit int, sortBy string) ([]domain.TopProduct, error)
	GetTopCategories(ctx context.Context, params domain.TopCategoriesParams) ([]domain.TopCategory, error)
	GetTopSubcategories(ctx context.Context, params domain.SubcategoryAnalyticsParams) ([]domain.TopSubcategory, error)
}

type ReturnRepository interface {
	CreateReturnRequest(ctx context.Context, returnRequest *domain.ReturnRequest) error
	GetReturnRequestByOrderID(ctx context.Context, orderID int64) (*domain.ReturnRequest, error)
	GetUserReturnRequests(ctx context.Context, userID int64) ([]*domain.ReturnRequest, error)
	UpdateApprovedOrRejected(ctx context.Context, returnRequest *domain.ReturnRequest) error
	GetByID(ctx context.Context, id int64) (*domain.ReturnRequest, error)
	UpdateRefundDetails(ctx context.Context, returnRequest *domain.ReturnRequest) error
	BeginTx(ctx context.Context) (*sql.Tx, error)
	UpdateReturnRequest(ctx context.Context, returnRequest *domain.ReturnRequest) error
	GetPendingReturnRequests(ctx context.Context, params domain.ReturnRequestParams) ([]*domain.ReturnRequest, int64, error)
}

type PaymentRepository interface {
	GetByOrderID(ctx context.Context, orderID int64) (*domain.Payment, error)
	InitiateRefund(ctx context.Context, paymentID int64) error
	GetByRazorpayOrderID(ctx context.Context, razorpayOrderID string) (*domain.Payment, error)
	UpdatePayment(ctx context.Context, payment *domain.Payment) error
	GetByOrderIDTx(ctx context.Context, tx *sql.Tx, orderID int64) (*domain.Payment, error)
	UpdateStatusTx(ctx context.Context, tx *sql.Tx, paymentID int64, status string) error
}
