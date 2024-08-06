package http

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/mohamedfawas/rmshop-clean-architecture/internal/delivery/http/handlers"
	"github.com/mohamedfawas/rmshop-clean-architecture/internal/delivery/http/middleware"
)

func NewRouter(userHandler *handlers.UserHandler, adminHandler *handlers.AdminHandler, categoryHandler *handlers.CategoryHandler, subCategoryHandler *handlers.SubCategoryHandler) http.Handler {
	r := mux.NewRouter()

	// User routes
	r.HandleFunc("/user/register", userHandler.Register)
	r.HandleFunc("/user/login", userHandler.Login)
	r.HandleFunc("/user/logout", middleware.JWTAuthMiddleware(userHandler.Logout))

	// Admin routes
	r.HandleFunc("/admin/login", adminHandler.Login)
	r.HandleFunc("/admin/logout", middleware.JWTAuthMiddleware(adminHandler.Logout))

	// Category routes
	r.HandleFunc("/admin/categories", middleware.JWTAuthMiddleware(middleware.AdminAuthMiddleware(categoryHandler.CreateCategory)))

	// Subcategory routes
	r.HandleFunc("/admin/categories/{categoryId}/subcategories",
		middleware.JWTAuthMiddleware(middleware.AdminAuthMiddleware(subCategoryHandler.CreateSubCategory))).Methods("POST")

	// Wrap the entire mux with the logging middleware
	return middleware.LoggingMiddleware(r)
}
