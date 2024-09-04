package http

import (
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/mohamedfawas/rmshop-clean-architecture/internal/delivery/http/handlers"
	"github.com/mohamedfawas/rmshop-clean-architecture/internal/delivery/http/middleware"
	"github.com/mohamedfawas/rmshop-clean-architecture/pkg/auth"
	"golang.org/x/time/rate"
)

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Incoming request: %s %s", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
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
	checkoutHandler *handlers.CheckoutHandler) http.Handler {

	log.Println("Setting up router...")
	r := mux.NewRouter()

	// Set up logging middleware
	r.Use(loggingMiddleware)

	// JWT middleware
	jwtAuth := middleware.JWTAuthMiddleware(tokenBlacklist)

	// Admin auth middleware
	adminAuth := middleware.AdminAuthMiddleware

	// User auth middleware
	userAuth := middleware.UserAuthMiddleware

	// Admin routes
	r.HandleFunc("/admin/login", adminHandler.Login).Methods("POST")
	r.HandleFunc("/admin/logout", chainMiddleware(jwtAuth, adminAuth)(adminHandler.Logout)).Methods("POST")

	// Admin routes: Category routes
	r.HandleFunc("/admin/categories", chainMiddleware(jwtAuth, adminAuth)(categoryHandler.CreateCategory)).Methods("POST")
	r.HandleFunc("/admin/categories", chainMiddleware(jwtAuth, adminAuth)(categoryHandler.GetAllCategories)).Methods("GET")
	r.HandleFunc("/admin/categories/{categoryId}", chainMiddleware(jwtAuth, adminAuth)(categoryHandler.GetActiveCategoryByID)).Methods("GET")
	r.HandleFunc("/admin/categories/{categoryId}", chainMiddleware(jwtAuth, adminAuth)(categoryHandler.UpdateCategory)).Methods("PUT")
	r.HandleFunc("/admin/categories/{categoryId}", chainMiddleware(jwtAuth, adminAuth)(categoryHandler.SoftDeleteCategory)).Methods("DELETE")

	// Admin routes: Subcategory routes
	r.HandleFunc("/admin/categories/{categoryId}/subcategories", chainMiddleware(jwtAuth, adminAuth)(subCategoryHandler.CreateSubCategory)).Methods("POST")
	r.HandleFunc("/admin/categories/{categoryId}/subcategories", chainMiddleware(jwtAuth, adminAuth)(subCategoryHandler.GetSubCategoriesByCategory)).Methods("GET")
	r.HandleFunc("/admin/categories/{categoryId}/subcategories/{subcategoryId}", chainMiddleware(jwtAuth, adminAuth)(subCategoryHandler.GetSubCategoryByID)).Methods("GET")
	r.HandleFunc("/admin/categories/{categoryId}/subcategories/{subcategoryId}", chainMiddleware(jwtAuth, adminAuth)(subCategoryHandler.UpdateSubCategory)).Methods("PUT")
	r.HandleFunc("/admin/categories/{categoryId}/subcategories/{subcategoryId}", chainMiddleware(jwtAuth, adminAuth)(subCategoryHandler.SoftDeleteSubCategory)).Methods("DELETE")

	// Admin routes: Product routes
	r.HandleFunc("/admin/products", chainMiddleware(jwtAuth, adminAuth)(productHandler.CreateProduct)).Methods("POST")
	r.HandleFunc("/admin/products", chainMiddleware(jwtAuth, adminAuth)(productHandler.GetAllProducts)).Methods("GET")
	r.HandleFunc("/admin/products/{productId}", chainMiddleware(jwtAuth, adminAuth)(productHandler.GetProductByID)).Methods("GET")
	r.HandleFunc("/admin/products/{productId}", chainMiddleware(jwtAuth, adminAuth)(productHandler.UpdateProduct)).Methods("PUT")
	r.HandleFunc("/admin/products/{productId}", chainMiddleware(jwtAuth, adminAuth)(productHandler.SoftDeleteProduct)).Methods("DELETE")

	// Product image management
	r.HandleFunc("/admin/products/{productId}/images", chainMiddleware(jwtAuth, adminAuth)(productHandler.AddProductImages)).Methods("POST")
	r.HandleFunc("/admin/products/{productId}/images/{imageId}", chainMiddleware(jwtAuth, adminAuth)(productHandler.DeleteProductImage)).Methods("DELETE")

	// Admin routes: coupon routes
	r.HandleFunc("/admin/coupons", chainMiddleware(jwtAuth, adminAuth)(couponHandler.CreateCoupon)).Methods("POST")

	// User routes
	r.HandleFunc("/user/login", userHandler.Login).Methods("POST")
	r.HandleFunc("/user/signup", userHandler.InitiateSignUp).Methods("POST")
	r.HandleFunc("/user/verify-otp", userHandler.VerifyOTP).Methods("POST")

	// Create a new IP rate limiter
	otpResendLimiter := middleware.NewIPRateLimiter(rate.Every(30*time.Second), 1)

	// Allow 1 request per 30 seconds for OTP resend
	r.HandleFunc("/user/resend-otp", middleware.RateLimitMiddleware(userHandler.ResendOTP, otpResendLimiter)).Methods("POST")

	// Protected user routes
	r.HandleFunc("/user/logout", chainMiddleware(jwtAuth, userAuth)(userHandler.Logout)).Methods("POST")
	r.HandleFunc("/user/profile", chainMiddleware(jwtAuth, userAuth)(userHandler.GetUserProfile)).Methods("GET")
	r.HandleFunc("/user/profile", chainMiddleware(jwtAuth, userAuth)(userHandler.UpdateProfile)).Methods("PUT")
	r.HandleFunc("/user/addresses", chainMiddleware(jwtAuth, userAuth)(userHandler.AddUserAddress)).Methods("POST")
	r.HandleFunc("/user/addresses/{addressId}", chainMiddleware(jwtAuth, userAuth)(userHandler.UpdateUserAddress)).Methods("PATCH")
	r.HandleFunc("/user/addresses", chainMiddleware(jwtAuth, userAuth)(userHandler.GetUserAddresses)).Methods("GET")
	r.HandleFunc("/user/addresses/{addressId}", chainMiddleware(jwtAuth, userAuth)(userHandler.DeleteUserAddress)).Methods("DELETE")
	r.HandleFunc("/user/cart/items", chainMiddleware(jwtAuth, userAuth)(cartHandler.AddToCart)).Methods("POST")
	r.HandleFunc("/user/cart", chainMiddleware(jwtAuth, userAuth)(cartHandler.GetUserCart)).Methods("GET")
	r.HandleFunc("/user/cart/items/{itemId}", chainMiddleware(jwtAuth, userAuth)(cartHandler.UpdateCartItemQuantity)).Methods("PATCH")
	r.HandleFunc("/user/cart/items/{itemId}", chainMiddleware(jwtAuth, userAuth)(cartHandler.DeleteCartItem)).Methods("DELETE")
	r.HandleFunc("/user/cart/apply-coupon", chainMiddleware(jwtAuth, userAuth)(couponHandler.ApplyCoupon)).Methods("POST")
	r.HandleFunc("/user/checkout", chainMiddleware(jwtAuth, userAuth)(checkoutHandler.CreateCheckout)).Methods("POST")
	r.HandleFunc("/user/checkout/{checkout_id}/apply-coupon", chainMiddleware(jwtAuth, userAuth)(checkoutHandler.ApplyCoupon)).Methods("POST")
	r.HandleFunc("/user/checkout/{checkout_id}/address", chainMiddleware(jwtAuth, userAuth)(checkoutHandler.UpdateCheckoutAddress)).Methods("PATCH")

	// Public routes
	r.HandleFunc("/user/forgot-password", middleware.RateLimitMiddleware(userHandler.ForgotPassword, otpResendLimiter)).Methods("POST")
	r.HandleFunc("/user/reset-password", userHandler.ResetPassword).Methods("POST")

	log.Println("Router setup complete")
	return r
}
