package http

import (
	"log"
	"net/http"

	"github.com/mohamedfawas/rmshop-clean-architecture/internal/delivery/http/handlers"
	"github.com/mohamedfawas/rmshop-clean-architecture/internal/delivery/http/middleware"
)

func NewRouter(userHandler *handlers.UserHandler) http.Handler {
	mux := http.NewServeMux()

	// User routes
	mux.HandleFunc("/user/register", userHandler.Register)
	mux.HandleFunc("/user/login", userHandler.Login)
	mux.HandleFunc("/user/logout", func(w http.ResponseWriter, r *http.Request) {
		log.Println("Logout route accessed")
		middleware.JWTAuthMiddleware(userHandler.Logout)(w, r)
	})

	// Wrap the entire mux with the logging middleware
	return middleware.LoggingMiddleware(mux)
}
