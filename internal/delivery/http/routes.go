package http

import (
	"net/http"

	"github.com/mohamedfawas/rmshop-clean-architecture/internal/delivery/http/handlers"
	"github.com/mohamedfawas/rmshop-clean-architecture/internal/delivery/http/middleware"
)

func NewRouter(userHandler *handlers.UserHandler, adminHandler *handlers.AdminHandler) http.Handler {
	mux := http.NewServeMux()

	// User routes
	mux.HandleFunc("/user/register", userHandler.Register)
	mux.HandleFunc("/user/login", userHandler.Login)
	mux.HandleFunc("/user/logout", middleware.JWTAuthMiddleware(userHandler.Logout))

	// Admin routes
	mux.HandleFunc("/admin/login", adminHandler.Login)
	mux.HandleFunc("/admin/logout", middleware.JWTAuthMiddleware(adminHandler.Logout))

	// Wrap the entire mux with the logging middleware
	return middleware.LoggingMiddleware(mux)
}
