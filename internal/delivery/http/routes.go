package http

import (
	"net/http"

	"github.com/mohamedfawas/rmshop-clean-architecture/internal/delivery/http/handlers"
	"github.com/mohamedfawas/rmshop-clean-architecture/internal/delivery/http/middleware"
)

func NewRouter(userHandler *handlers.UserHandler) http.Handler {
	mux := http.NewServeMux()

	// User routes
	mux.HandleFunc("/user/register", userHandler.Register)

	// Wrap the entire mux with the logging middleware
	return middleware.LoggingMiddleware(mux)
}
