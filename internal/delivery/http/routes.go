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

// loggingMiddleware is a middleware function that logs incoming HTTP requests
// and their corresponding responses using Logrus. It logs the method, URL path,
// remote address, user agent, status code, and the time taken to process the request.
//
// Parameters:
//
//	next: The next http.Handler in the middleware chain.
//
// Returns:
//
//	http.Handler: A wrapped http.Handler that logs request and response details.
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Log the request details using Logrus
		logrus.WithFields(logrus.Fields{
			"method": r.Method,
			"url":    r.URL.Path,
			"remote": r.RemoteAddr,
			"agent":  r.UserAgent(),
		}).Info("Incoming request")

		// Capture the response status code
		rec := statusRecorder{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(&rec, r)

		// Log the response status and duration
		logrus.WithFields(logrus.Fields{
			"status":   rec.statusCode,
			"duration": time.Since(start).String(),
		}).Info("Completed request")
	})
}

// chainMiddleware chains a series of middleware functions into a single middleware.
// The returned middleware, when applied to an http.HandlerFunc, will execute each
// middleware in the order provided, wrapping the final handler with all the middleware.
//
// Parameters:
//   - middlewares: A variadic number of middleware functions. Each middleware should
//     take an http.HandlerFunc and return an http.HandlerFunc.
//
// Returns:
//   - A function that takes an http.HandlerFunc (final handler) as input and returns
//     a new http.HandlerFunc that chains the provided middlewares around the final handler.
func chainMiddleware(middlewares ...func(http.HandlerFunc) http.HandlerFunc) func(http.HandlerFunc) http.HandlerFunc {
	return func(final http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			last := final
			for i := len(middlewares) - 1; i >= 0; i-- {
				last = middlewares[i](last)
			}
			last(w, r)
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
	r.HandleFunc("/admin/orders/{orderId}/status", chainMiddleware(jwtAuth, adminAuth)(orderHandler.UpdateOrderStatus)).Methods("PATCH")

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

	// User routes : Cart management
	r.HandleFunc("/user/cart/items", chainMiddleware(jwtAuth, userAuth)(cartHandler.AddToCart)).Methods("POST")
	r.HandleFunc("/user/cart", chainMiddleware(jwtAuth, userAuth)(cartHandler.GetUserCart)).Methods("GET")
	r.HandleFunc("/user/cart/items/{itemId}", chainMiddleware(jwtAuth, userAuth)(cartHandler.UpdateCartItemQuantity)).Methods("PATCH")
	r.HandleFunc("/user/cart/items/{itemId}", chainMiddleware(jwtAuth, userAuth)(cartHandler.DeleteCartItem)).Methods("DELETE")

	// apply coupon : remove this code later
	r.HandleFunc("/user/cart/apply-coupon", chainMiddleware(jwtAuth, userAuth)(couponHandler.ApplyCoupon)).Methods("POST")

	// User routes : Checkout
	r.HandleFunc("/user/checkout", chainMiddleware(jwtAuth, userAuth)(checkoutHandler.CreateCheckout)).Methods("POST")
	// apply coupon
	r.HandleFunc("/user/checkout/{checkout_id}/apply-coupon", chainMiddleware(jwtAuth, userAuth)(checkoutHandler.ApplyCoupon)).Methods("POST")
	// remove coupon
	r.HandleFunc("/user/checkout/{checkout_id}/apply-coupon", chainMiddleware(jwtAuth, userAuth)(checkoutHandler.RemoveAppliedCoupon)).Methods("DELETE")
	r.HandleFunc("/user/checkout/{checkout_id}/address", chainMiddleware(jwtAuth, userAuth)(checkoutHandler.UpdateCheckoutAddress)).Methods("PATCH")
	r.HandleFunc("/user/checkout/{checkout_id}/summary", chainMiddleware(jwtAuth, userAuth)(checkoutHandler.GetCheckoutSummary)).Methods("GET")
	// r.HandleFunc("/user/checkout/{checkout_id}/place-order", chainMiddleware(jwtAuth, userAuth)(checkoutHandler.PlaceOrder)).Methods("POST")

	// User routes : Order management
	// place order using razorpay
	r.HandleFunc("/user/checkout/{checkout_id}/place-order/razorpay", chainMiddleware(jwtAuth, userAuth)(orderHandler.PlaceOrderRazorpay)).Methods("POST")
	r.HandleFunc("/user/checkout/{checkout_id}/place-order/cod", chainMiddleware(jwtAuth, userAuth)(orderHandler.PlaceOrderCOD)).Methods("POST")

	r.HandleFunc("/user/orders/{order_id}", chainMiddleware(jwtAuth, userAuth)(orderHandler.GetOrderConfirmation)).Methods("GET")
	r.HandleFunc("/user/orders", chainMiddleware(jwtAuth, userAuth)(orderHandler.GetUserOrders)).Methods("GET")
	r.HandleFunc("/user/orders/{orderId}/cancel", chainMiddleware(jwtAuth, userAuth)(orderHandler.CancelOrder)).Methods("POST")

	// order return
	r.HandleFunc("/user/orders/{orderId}/return", chainMiddleware(jwtAuth, userAuth)(orderHandler.InitiateReturn)).Methods("POST")

	// Public routes
	r.HandleFunc("/user/forgot-password", middleware.RateLimitMiddleware(userHandler.ForgotPassword, otpResendLimiter)).Methods("POST")
	r.HandleFunc("/user/reset-password", userHandler.ResetPassword).Methods("POST")

	// Public routes : Homepage
	r.HandleFunc("/products", productHandler.GetProducts).Methods("GET")
	r.HandleFunc("/products/{productId}", productHandler.GetPublicProductByID).Methods("GET")
	r.HandleFunc("/coupons", couponHandler.GetAllCoupons).Methods("GET")

	r.HandleFunc("/home/payment", paymentHandler.RenderPaymentPage).Methods("GET")
	r.HandleFunc("/home/razorpay-payment", paymentHandler.ProcessRazorpayPayment).Methods("POST")

	// sales report
	r.HandleFunc("/admin/sales-report", chainMiddleware(jwtAuth, adminAuth)(salesHandler.GetSalesReport)).Methods("GET")

	log.Println("Router setup complete")
	return r
}
