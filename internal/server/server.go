package server

import (
	"database/sql"
	"log"
	"net/http"

	httpDelivery "github.com/mohamedfawas/rmshop-clean-architecture/internal/delivery/http"
	"github.com/mohamedfawas/rmshop-clean-architecture/internal/delivery/http/handlers"
	"github.com/mohamedfawas/rmshop-clean-architecture/internal/repository/postgres"
	"github.com/mohamedfawas/rmshop-clean-architecture/internal/usecase"
	"github.com/mohamedfawas/rmshop-clean-architecture/pkg/cloudinary"
	email "github.com/mohamedfawas/rmshop-clean-architecture/pkg/emailVerify"
)

// Server struct holds the router which will be used to handle HTTP requests
type Server struct {
	router http.Handler
}

// NewServer creates and returns a new Server instance
func NewServer(db *sql.DB, emailSender *email.Sender, cloudinaryService *cloudinary.CloudinaryService) *Server {
	log.Println("Initializing server components...")

	// Handler -> UseCase -> Repository.
	// User components initialization
	userRepo := postgres.NewUserRepository(db)                   // Create a new user repository with the database connection
	userUseCase := usecase.NewUserUseCase(userRepo, emailSender) // Create a new user use case with the user repository and email sender
	userHandler := handlers.NewUserHandler(userUseCase)          // Create a new user handler with the user use case
	log.Println("User components initialized")

	// Admin components initialization
	adminRepo := postgres.NewAdminRepository(db)
	adminUseCase := usecase.NewAdminUseCase(adminRepo)
	adminHandler := handlers.NewAdminHandler(adminUseCase)
	log.Println("Admin components initialized")

	// Category components
	categoryRepo := postgres.NewCategoryRepository(db)
	categoryUseCase := usecase.NewCategoryUseCase(categoryRepo)
	categoryHandler := handlers.NewCategoryHandler(categoryUseCase)
	log.Println("Category components initialized")

	// Subcategory components
	subCategoryRepo := postgres.NewSubCategoryRepository(db)
	subCategoryUseCase := usecase.NewSubCategoryUseCase(subCategoryRepo, categoryRepo)
	subCategoryHandler := handlers.NewSubCategoryHandler(subCategoryUseCase)
	log.Println("Sub-category components initialized")

	// Product components
	productRepo := postgres.NewProductRepository(db)
	productUseCase := usecase.NewProductUseCase(productRepo, subCategoryRepo, cloudinaryService)
	productHandler := handlers.NewProductHandler(productUseCase)
	log.Println("Product components initialized")

	// Initialize the router with all handlers
	router := httpDelivery.NewRouter(
		userHandler,
		adminHandler,
		categoryHandler,
		subCategoryHandler,
		productHandler,
	)
	log.Println("Router initialized")

	// Return a new Server instance with the initialized router
	return &Server{
		router: router,
	}
}

// ServeHTTP makes the Server struct implement the http.Handler interface
// This method is called for every HTTP request to the server
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}
