package http

import (
	"log"
	"net/http"
	"text/template"
	"time"

	"github.com/gorilla/mux"
	"github.com/mohamedfawas/rmshop-clean-architecture/internal/delivery/http/handlers"
	"github.com/mohamedfawas/rmshop-clean-architecture/internal/delivery/http/middleware"
	"github.com/mohamedfawas/rmshop-clean-architecture/pkg/auth"
	"github.com/sirupsen/logrus"
	"golang.org/x/time/rate"
)

type statusRecorder struct {
	http.ResponseWriter
	statusCode int
}

/*
LogginMiddleware : used to log the details of every requests

Ouput example:
INFO[2024-09-25 12:00:00] Incoming request method=GET url=/api/v1/login remote=192.168.1.100 agent="Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/93.0.4577.63 Safari/537.36"
*/
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now() // record start time, for recording request processing

		// Log the request details using Logrus
		// WithFields : used to add structured fields (key-value pairs) to the log entry.
		logrus.WithFields(logrus.Fields{
			"method": r.Method,
			"url":    r.URL.Path,
			"remote": r.RemoteAddr,
			"agent":  r.UserAgent(),
		}).Info("Incoming request") // Info method :  used to log the message as an informational message

		// Capture the response and response status code
		rec := statusRecorder{ResponseWriter: w, statusCode: http.StatusOK}
		// pass the request and record to next handler in chain
		next.ServeHTTP(&rec, r)

		// Log the response status and duration
		logrus.WithFields(logrus.Fields{
			"status":   rec.statusCode,
			"duration": time.Since(start).String(),
		}).Info("Completed request")
	})
}

// chainMiddleware is a function that takes a list of middleware functions
// and returns a single middleware function that applies them in sequence.
func chainMiddleware(middlewares ...func(http.HandlerFunc) http.HandlerFunc) func(http.HandlerFunc) http.HandlerFunc {

	// This returned function takes the final handler (the main logic for the request)
	return func(final http.HandlerFunc) http.HandlerFunc {

		// This function is the one that actually handles the request (w, r)
		return func(w http.ResponseWriter, r *http.Request) {

			// Start by setting 'currentHandler' as the final handler
			currentHandler := final

			// Loop through the middlewares in reverse order
			// (this is needed because middleware wraps the next one)
			for i := len(middlewares) - 1; i >= 0; i-- {

				// Wrap the current handler with the middleware
				currentHandler = middlewares[i](currentHandler)
			}

			// After all middlewares have wrapped the handler, call the final one
			currentHandler(w, r)
		}
	}
}

func NewRouter(userHandler *handlers.UserHandler,
	adminHandler *handlers.AdminHandler,
	categoryHandler *handlers.CategoryHandler,
	subCategoryHandler *handlers.SubCategoryHandler,
	productHandler *handlers.ProductHandler,
	tokenBlacklist *auth.TokenBlacklist,
	cartHandler *handlers.CartHandler,
	couponHandler *handlers.CouponHandler,
	checkoutHandler *handlers.CheckoutHandler,
	orderHandler *handlers.OrderHandler,
	inventoryHandler *handlers.InventoryHandler,
	paymentHandler *handlers.PaymentHandler,
	wishlistHandler *handlers.WishlistHandler,
	walletHandler *handlers.WalletHandler,
	salesHandler *handlers.SalesHandler,
	analyticsHandler *handlers.AnalyticsHandler,
	returnHandler *handlers.ReturnHandler,
	templates *template.Template) http.Handler {
	log.Println("Setting up router...")

	r := mux.NewRouter() // set up mux router

	// Set up logging middleware
	r.Use(loggingMiddleware)

	// JWT middleware : token validation
	jwtAuth := middleware.JWTAuthMiddleware(tokenBlacklist)

	// Admin auth middleware
	adminAuth := middleware.AdminAuthMiddleware

	// User auth middleware
	userAuth := middleware.UserAuthMiddleware

	// Admin login and logout
	r.HandleFunc("/admin/login", adminHandler.Login).Methods("POST")
	r.HandleFunc("/admin/logout", chainMiddleware(jwtAuth, adminAuth)(adminHandler.Logout)).Methods("POST")

	// Admin routes: Category management
	r.HandleFunc("/admin/categories", chainMiddleware(jwtAuth, adminAuth)(categoryHandler.CreateCategory)).Methods("POST")
	r.HandleFunc("/admin/categories", chainMiddleware(jwtAuth, adminAuth)(categoryHandler.GetAllCategories)).Methods("GET")
	r.HandleFunc("/admin/categories/{categoryId}", chainMiddleware(jwtAuth, adminAuth)(categoryHandler.GetActiveCategoryByID)).Methods("GET")
	r.HandleFunc("/admin/categories/{categoryId}", chainMiddleware(jwtAuth, adminAuth)(categoryHandler.UpdateCategory)).Methods("PUT")
	r.HandleFunc("/admin/categories/{categoryId}", chainMiddleware(jwtAuth, adminAuth)(categoryHandler.SoftDeleteCategory)).Methods("DELETE")

	// Admin routes: Subcategory management
	r.HandleFunc("/admin/categories/{categoryId}/subcategories", chainMiddleware(jwtAuth, adminAuth)(subCategoryHandler.CreateSubCategory)).Methods("POST")
	r.HandleFunc("/admin/categories/{categoryId}/subcategories", chainMiddleware(jwtAuth, adminAuth)(subCategoryHandler.GetSubCategoriesByCategory)).Methods("GET")
	r.HandleFunc("/admin/categories/{categoryId}/subcategories/{subcategoryId}", chainMiddleware(jwtAuth, adminAuth)(subCategoryHandler.GetSubCategoryByID)).Methods("GET")
	r.HandleFunc("/admin/categories/{categoryId}/subcategories/{subcategoryId}", chainMiddleware(jwtAuth, adminAuth)(subCategoryHandler.UpdateSubCategory)).Methods("PUT")
	r.HandleFunc("/admin/categories/{categoryId}/subcategories/{subcategoryId}", chainMiddleware(jwtAuth, adminAuth)(subCategoryHandler.SoftDeleteSubCategory)).Methods("DELETE")

	// Admin routes: Product management
	r.HandleFunc("/admin/products", chainMiddleware(jwtAuth, adminAuth)(productHandler.CreateProduct)).Methods("POST")
	r.HandleFunc("/admin/products", chainMiddleware(jwtAuth, adminAuth)(productHandler.GetAllProducts)).Methods("GET")
	r.HandleFunc("/admin/products/{productId}", chainMiddleware(jwtAuth, adminAuth)(productHandler.GetProductByID)).Methods("GET")
	r.HandleFunc("/admin/products/{productId}", chainMiddleware(jwtAuth, adminAuth)(productHandler.UpdateProduct)).Methods("PUT")
	r.HandleFunc("/admin/products/{productId}", chainMiddleware(jwtAuth, adminAuth)(productHandler.SoftDeleteProduct)).Methods("DELETE")

	// Admin routes : Product image management
	r.HandleFunc("/admin/products/{productId}/images", chainMiddleware(jwtAuth, adminAuth)(productHandler.AddProductImages)).Methods("POST")
	r.HandleFunc("/admin/products/{productId}/images/{imageId}", chainMiddleware(jwtAuth, adminAuth)(productHandler.DeleteProductImage)).Methods("DELETE")

	// Admin routes : Coupon management
	r.HandleFunc("/admin/coupons", chainMiddleware(jwtAuth, adminAuth)(couponHandler.CreateCoupon)).Methods("POST")
	r.HandleFunc("/admin/coupons/{coupon_id}", chainMiddleware(jwtAuth, adminAuth)(couponHandler.UpdateCoupon)).Methods("PATCH")

	// admin routes : order management
	r.HandleFunc("/admin/orders", chainMiddleware(jwtAuth, adminAuth)(orderHandler.GetOrders)).Methods("GET")

	// admin : inventory management
	r.HandleFunc("/admin/inventory", chainMiddleware(jwtAuth, adminAuth)(inventoryHandler.GetInventory)).Methods("GET")
	r.HandleFunc("/admin/inventory/{productId}", chainMiddleware(jwtAuth, adminAuth)(inventoryHandler.UpdateProductStock)).Methods("PATCH")

	// User routes : login, sign up
	r.HandleFunc("/user/login", userHandler.Login).Methods("POST")
	r.HandleFunc("/user/signup", userHandler.InitiateSignUp).Methods("POST")
	r.HandleFunc("/user/verify-otp", userHandler.VerifyOTP).Methods("POST")
	// Create a new IP rate limiter
	otpResendLimiter := middleware.NewIPRateLimiter(rate.Every(30*time.Second), 1)
	// Allow 1 request per 30 seconds for OTP resend
	r.HandleFunc("/user/resend-otp", middleware.RateLimitMiddleware(userHandler.ResendOTP, otpResendLimiter)).Methods("POST")

	// User routes : Logout
	r.HandleFunc("/user/logout", chainMiddleware(jwtAuth, userAuth)(userHandler.Logout)).Methods("POST")

	// User routes : User profile management
	r.HandleFunc("/user/profile", chainMiddleware(jwtAuth, userAuth)(userHandler.GetUserProfile)).Methods("GET")
	r.HandleFunc("/user/profile", chainMiddleware(jwtAuth, userAuth)(userHandler.UpdateProfile)).Methods("PUT")

	// User routes : User Address management
	r.HandleFunc("/user/addresses", chainMiddleware(jwtAuth, userAuth)(userHandler.AddUserAddress)).Methods("POST")
	r.HandleFunc("/user/addresses/{addressId}", chainMiddleware(jwtAuth, userAuth)(userHandler.UpdateUserAddress)).Methods("PATCH")
	r.HandleFunc("/user/addresses", chainMiddleware(jwtAuth, userAuth)(userHandler.GetUserAddresses)).Methods("GET")
	r.HandleFunc("/user/addresses/{addressId}", chainMiddleware(jwtAuth, userAuth)(userHandler.DeleteUserAddress)).Methods("DELETE")

	// wish list
	r.HandleFunc("/user/wishlist/items", chainMiddleware(jwtAuth)(wishlistHandler.AddToWishlist)).Methods("POST")
	r.HandleFunc("/user/wishlist/items/{productId}", chainMiddleware(jwtAuth, userAuth)(wishlistHandler.RemoveFromWishlist)).Methods("DELETE")
	r.HandleFunc("/user/wishlist", chainMiddleware(jwtAuth, userAuth)(wishlistHandler.GetUserWishlist)).Methods("GET")

	// user wallet
	r.HandleFunc("/user/wallet/balance", chainMiddleware(jwtAuth, userAuth)(walletHandler.GetWalletBalance)).Methods("GET")
	r.HandleFunc("/user/wallet/transactions", chainMiddleware(jwtAuth, userAuth)(walletHandler.GetWalletTransactions)).Methods("GET")

	// User routes : Cart management
	r.HandleFunc("/user/cart/items", chainMiddleware(jwtAuth, userAuth)(cartHandler.AddToCart)).Methods("POST")
	r.HandleFunc("/user/cart", chainMiddleware(jwtAuth, userAuth)(cartHandler.GetUserCart)).Methods("GET")
	r.HandleFunc("/user/cart/items/{itemId}", chainMiddleware(jwtAuth, userAuth)(cartHandler.UpdateCartItemQuantity)).Methods("PATCH")
	r.HandleFunc("/user/cart/items/{itemId}", chainMiddleware(jwtAuth, userAuth)(cartHandler.DeleteCartItem)).Methods("DELETE")

	// User routes : Checkout
	r.HandleFunc("/user/checkout", chainMiddleware(jwtAuth, userAuth)(checkoutHandler.CreateCheckout)).Methods("POST")
	// apply coupon
	r.HandleFunc("/user/checkout/{checkout_id}/apply-coupon", chainMiddleware(jwtAuth, userAuth)(checkoutHandler.ApplyCoupon)).Methods("POST")
	// remove coupon
	r.HandleFunc("/user/checkout/{checkout_id}/apply-coupon", chainMiddleware(jwtAuth, userAuth)(checkoutHandler.RemoveAppliedCoupon)).Methods("DELETE")
	// add shipping address to checkout
	r.HandleFunc("/user/checkout/{checkout_id}/address", chainMiddleware(jwtAuth, userAuth)(checkoutHandler.UpdateCheckoutAddress)).Methods("PATCH")
	// get checkout summary
	r.HandleFunc("/user/checkout/{checkout_id}/summary", chainMiddleware(jwtAuth, userAuth)(checkoutHandler.GetCheckoutSummary)).Methods("GET")

	// User routes : Order management
	// place order using razorpay
	r.HandleFunc("/user/checkout/{checkout_id}/place-order/razorpay", chainMiddleware(jwtAuth, userAuth)(orderHandler.PlaceOrderRazorpay)).Methods("POST")
	// place order using cod
	r.HandleFunc("/user/checkout/{checkout_id}/place-order/cod", chainMiddleware(jwtAuth, userAuth)(orderHandler.PlaceOrderCOD)).Methods("POST")
	// Get order details by order id
	r.HandleFunc("/user/orders/{order_id}", chainMiddleware(jwtAuth, userAuth)(orderHandler.GetOrderDetails)).Methods("GET")
	// Get order history
	r.HandleFunc("/user/orders", chainMiddleware(jwtAuth, userAuth)(orderHandler.GetUserOrders)).Methods("GET")

	// order return
	r.HandleFunc("/user/orders/{orderId}/return", chainMiddleware(jwtAuth, userAuth)(returnHandler.InitiateReturn)).Methods("POST")
	r.HandleFunc("/user/orders/{orderId}/return", chainMiddleware(jwtAuth, userAuth)(returnHandler.GetReturnRequestByOrderID)).Methods("GET")
	r.HandleFunc("/user/returns", chainMiddleware(jwtAuth, userAuth)(returnHandler.GetUserReturnRequests)).Methods("GET")

	// admin : order delivery update
	r.HandleFunc("/admin/orders/{orderId}/delivery-status", chainMiddleware(jwtAuth, adminAuth)(orderHandler.UpdateOrderDeliveryStatus)).Methods("PATCH")

	// order return : admin approve/reject
	r.HandleFunc("/admin/returns/{returnId}", chainMiddleware(jwtAuth, adminAuth)(returnHandler.UpdateReturnRequest)).Methods("PATCH")
	// order return : admin initiate refund
	r.HandleFunc("/admin/returns/{returnId}/refund", chainMiddleware(jwtAuth, adminAuth)(returnHandler.InitiateRefund)).Methods("POST")
	// order return : admin refund complete // Remove this code, belongs to old approach
	r.HandleFunc("/admin/returns/{returnId}/refund", chainMiddleware(jwtAuth, adminAuth)(returnHandler.CompleteRefund)).Methods("PATCH")

	// Order cancellation
	// user initiate order cancellation
	r.HandleFunc("/user/orders/{orderId}/cancel", chainMiddleware(jwtAuth, userAuth)(orderHandler.CancelOrder)).Methods("POST")
	// Admin gets all the cancellation requests created by users
	r.HandleFunc("/admin/orders/cancellation-requests", chainMiddleware(jwtAuth, adminAuth)(orderHandler.GetCancellationRequests)).Methods("GET")
	// Admin approve order cancellation
	r.HandleFunc("/admin/orders/{orderId}/cancellation", chainMiddleware(jwtAuth, adminAuth)(orderHandler.AdminApproveCancellation)).Methods("PATCH")
	// Admin initiate order cancellation
	r.HandleFunc("/admin/orders/{orderId}/cancel", chainMiddleware(jwtAuth, adminAuth)(orderHandler.AdminCancelOrder)).Methods("POST")

	// order invoice
	r.HandleFunc("/user/orders/{orderId}/invoice", chainMiddleware(jwtAuth, userAuth)(orderHandler.GetOrderInvoice)).Methods("GET")

	// Public routes
	r.HandleFunc("/user/forgot-password", middleware.RateLimitMiddleware(userHandler.ForgotPassword, otpResendLimiter)).Methods("POST")
	r.HandleFunc("/user/reset-password", userHandler.ResetPassword).Methods("POST")

	// Public routes : Homepage
	r.HandleFunc("/products", productHandler.GetProducts).Methods("GET")
	r.HandleFunc("/products/{productId}", productHandler.GetPublicProductByID).Methods("GET")
	r.HandleFunc("/coupons", couponHandler.GetAllCoupons).Methods("GET")

	// razorpay gateway: front end api end points
	r.HandleFunc("/home/payment", paymentHandler.RenderPaymentPage).Methods("GET")
	r.HandleFunc("/home/razorpay-payment", paymentHandler.ProcessRazorpayPayment).Methods("POST")
	r.HandleFunc("/payment-failure", paymentHandler.RenderPaymentFailurePage).Methods("GET")
	r.HandleFunc("/payment-success", paymentHandler.RenderPaymentSuccessPage).Methods("GET")

	// sales report
	r.HandleFunc("/admin/sales-report/daily", chainMiddleware(jwtAuth, adminAuth)(salesHandler.GetDailySalesReport)).Methods("GET")
	r.HandleFunc("/admin/sales-report/weekly", chainMiddleware(jwtAuth, adminAuth)(salesHandler.GetWeeklySalesReport)).Methods("GET")
	r.HandleFunc("/admin/sales-report/monthly", chainMiddleware(jwtAuth, adminAuth)(salesHandler.GetMonthlySalesReport)).Methods("GET")
	r.HandleFunc("/admin/sales-report/custom", chainMiddleware(jwtAuth, adminAuth)(salesHandler.GetCustomSalesReport)).Methods("GET")

	// admin : analytics
	r.HandleFunc("/admin/analytics/top-products", chainMiddleware(jwtAuth, adminAuth)(analyticsHandler.GetTopProducts)).Methods("GET")
	r.HandleFunc("/admin/analytics/top-categories", chainMiddleware(jwtAuth, adminAuth)(analyticsHandler.GetTopCategories)).Methods("GET")
	r.HandleFunc("/admin/analytics/top-subcategories", chainMiddleware(jwtAuth, adminAuth)(analyticsHandler.GetTopSubcategories)).Methods("GET")

	log.Println("Router setup complete")
	return r
}
