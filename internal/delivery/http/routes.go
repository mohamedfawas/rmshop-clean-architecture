package http

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/mohamedfawas/rmshop-clean-architecture/internal/delivery/http/handlers"
	"github.com/mohamedfawas/rmshop-clean-architecture/internal/delivery/http/middleware"
)

func NewRouter(userHandler *handlers.UserHandler, adminHandler *handlers.AdminHandler, categoryHandler *handlers.CategoryHandler, subCategoryHandler *handlers.SubCategoryHandler, productHandler *handlers.ProductHandler) http.Handler {
	log.Println("Setting up router...")
	r := mux.NewRouter()

	// set up a middleware that logs every incoming HTTP request
	r.Use(func(next http.Handler) http.Handler { // This is using the Use method of the router to add middleware

		//Any request passing through this router will go through the middleware before reaching its final handler
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// This creates a new HandlerFunc, which is the actual middleware

			// Log information about the incoming request
			log.Printf("Incoming request: %s %s", r.Method, r.URL.Path) // This line logs the HTTP method (GET, POST, etc.) and the request path

			// Call the next handler in the chain
			next.ServeHTTP(w, r)
			// This line ensures that the request continues to be processed
			// by passing it to the next handler in the middleware chain
			//This is crucial; without this line, the request would not proceed to the intended route handler
		})
	})

	// Admin routes
	r.HandleFunc("/admin/login",
		adminHandler.Login)
	r.HandleFunc("/admin/logout",
		middleware.JWTAuthMiddleware(adminHandler.Logout))

	// Admin routes : Category routes
	r.HandleFunc("/admin/categories", //add category names
		middleware.JWTAuthMiddleware(
			middleware.AdminAuthMiddleware(
				categoryHandler.CreateCategory))).Methods("POST")
	r.HandleFunc("/admin/categories",
		middleware.JWTAuthMiddleware(
			middleware.AdminAuthMiddleware(
				categoryHandler.GetAllCategories))).Methods("GET")
	r.HandleFunc("/admin/categories/{categoryId}",
		middleware.JWTAuthMiddleware(
			middleware.AdminAuthMiddleware(
				categoryHandler.GetActiveCategoryByID))).Methods("GET")
	r.HandleFunc("/admin/categories/{categoryId}",
		middleware.JWTAuthMiddleware(
			middleware.AdminAuthMiddleware(
				categoryHandler.UpdateCategory))).Methods("PUT")
	r.HandleFunc("/admin/categories/{categoryId}",
		middleware.JWTAuthMiddleware(
			middleware.AdminAuthMiddleware(
				categoryHandler.SoftDeleteCategory))).Methods("DELETE")

	// Admin routes :Subcategory routes
	r.HandleFunc("/admin/categories/{categoryId}/subcategories",
		middleware.JWTAuthMiddleware(middleware.AdminAuthMiddleware(
			subCategoryHandler.CreateSubCategory))).Methods("POST")
	r.HandleFunc("/admin/categories/{categoryId}/subcategories",
		middleware.JWTAuthMiddleware(middleware.AdminAuthMiddleware(
			subCategoryHandler.GetSubCategoriesByCategory))).Methods("GET")
	r.HandleFunc("/admin/categories/{categoryId}/subcategories/{subcategoryId}",
		middleware.JWTAuthMiddleware(middleware.AdminAuthMiddleware(
			subCategoryHandler.GetSubCategoryByID))).Methods("GET")
	r.HandleFunc("/admin/categories/{categoryId}/subcategories/{subcategoryId}",
		middleware.JWTAuthMiddleware(middleware.AdminAuthMiddleware(
			subCategoryHandler.UpdateSubCategory))).Methods("PUT")
	r.HandleFunc("/admin/categories/{categoryId}/subcategories/{subcategoryId}",
		middleware.JWTAuthMiddleware(middleware.AdminAuthMiddleware(
			subCategoryHandler.SoftDeleteSubCategory))).Methods("DELETE")

	// Admin routes :Product routes
	r.HandleFunc("/admin/products",
		middleware.JWTAuthMiddleware(middleware.AdminAuthMiddleware(
			productHandler.CreateProduct))).Methods("POST")
	r.HandleFunc("/admin/products",
		middleware.JWTAuthMiddleware(middleware.AdminAuthMiddleware(
			productHandler.GetAllProducts))).Methods("GET")
	r.HandleFunc("/admin/products/{productId}",
		middleware.JWTAuthMiddleware(middleware.AdminAuthMiddleware(
			productHandler.GetProductByID))).Methods("GET")
	r.HandleFunc("/admin/products/{productId}",
		middleware.JWTAuthMiddleware(middleware.AdminAuthMiddleware(
			productHandler.UpdateProduct))).Methods("PUT")
	r.HandleFunc("/admin/products/{productId}",
		middleware.JWTAuthMiddleware(middleware.AdminAuthMiddleware(
			productHandler.SoftDeleteProduct))).Methods("DELETE")
	//admin route : update primary image
	r.HandleFunc("/admin/products/{productId}/primary-image",
		middleware.JWTAuthMiddleware(middleware.AdminAuthMiddleware(
			productHandler.UpdatePrimaryImage))).Methods("PUT")
	//admin route : add new product images
	r.HandleFunc("/admin/products/{productId}/images",
		middleware.JWTAuthMiddleware(middleware.AdminAuthMiddleware(
			productHandler.AddProductImages))).Methods("POST")
	//admin route : delete a product image
	r.HandleFunc("/admin/products/{productId}/images/{imageId}",
		middleware.JWTAuthMiddleware(middleware.AdminAuthMiddleware(
			productHandler.RemoveProductImage))).Methods("DELETE")

	// User routes
	r.HandleFunc("/user/login",
		userHandler.Login).Methods("POST")
	r.HandleFunc("/user/logout",
		middleware.JWTAuthMiddleware(userHandler.Logout))
	r.HandleFunc("/user/signup",
		userHandler.InitiateSignUp).Methods("POST")
	r.HandleFunc("/user/verify-otp",
		userHandler.VerifyOTP).Methods("POST")
	r.HandleFunc("/user/resend-otp",
		userHandler.ResendOTP).Methods("POST")

	//product listing on user side
	r.HandleFunc("/products", middleware.JWTAuthMiddleware(productHandler.GetActiveProducts)).Methods("GET")

	log.Println("Router setup complete")
	// Wrap the entire mux with the logging middleware
	return middleware.LoggingMiddleware(r)
}
