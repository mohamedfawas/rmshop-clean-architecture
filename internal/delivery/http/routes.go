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

	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log.Printf("Incoming request: %s %s", r.Method, r.URL.Path)
			next.ServeHTTP(w, r)
		})
	})
	r.HandleFunc("/admin/categories", func(w http.ResponseWriter, r *http.Request) {
		log.Println("Matched route: /admin/categories")
		middleware.JWTAuthMiddleware(middleware.AdminAuthMiddleware(categoryHandler.GetAllCategories))(w, r)
	}).Methods("GET")

	// User routes
	r.HandleFunc("/user/register", userHandler.Register)
	r.HandleFunc("/user/login", userHandler.Login)
	r.HandleFunc("/user/logout", middleware.JWTAuthMiddleware(userHandler.Logout))

	// Admin routes
	r.HandleFunc("/admin/login", adminHandler.Login)
	r.HandleFunc("/admin/logout", middleware.JWTAuthMiddleware(adminHandler.Logout))

	// Category routes
	r.HandleFunc("/admin/categories", middleware.JWTAuthMiddleware(middleware.AdminAuthMiddleware(categoryHandler.CreateCategory)))
	//r.HandleFunc("/admin/categories", middleware.JWTAuthMiddleware(middleware.AdminAuthMiddleware(categoryHandler.GetAllCategories))).Methods("GET")
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

	// Subcategory routes
	r.HandleFunc("/admin/categories/{categoryId}/subcategories",
		middleware.JWTAuthMiddleware(middleware.AdminAuthMiddleware(subCategoryHandler.CreateSubCategory))).Methods("POST")

	// Product routes
	r.HandleFunc("/admin/products", middleware.JWTAuthMiddleware(middleware.AdminAuthMiddleware(productHandler.CreateProduct))).Methods("POST")
	r.HandleFunc("/admin/products", middleware.JWTAuthMiddleware(middleware.AdminAuthMiddleware(productHandler.GetAllProducts))).Methods("GET")

	log.Println("Router setup complete")
	// Wrap the entire mux with the logging middleware
	return middleware.LoggingMiddleware(r)
}
