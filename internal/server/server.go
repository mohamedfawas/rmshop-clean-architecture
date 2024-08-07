package server

import (
	"database/sql"
	"log"
	"net/http"

	httpDelivery "github.com/mohamedfawas/rmshop-clean-architecture/internal/delivery/http"
	"github.com/mohamedfawas/rmshop-clean-architecture/internal/delivery/http/handlers"
	"github.com/mohamedfawas/rmshop-clean-architecture/internal/repository/postgres"
	"github.com/mohamedfawas/rmshop-clean-architecture/internal/usecase"
)

type Server struct {
	router http.Handler
}

func NewServer(db *sql.DB) *Server {
	log.Println("Initializing server components...")
	// User components
	// Handler -> UseCase -> Repository.
	userRepo := postgres.NewUserRepository(db)          // repository is responsible for handling data persistence and retrieval for user-related operations.
	userUseCase := usecase.NewUserUseCase(userRepo)     // Use cases contain the business logic of the application, user use case is initialized with the user repository, allowing it to perform data operations as needed.
	userHandler := handlers.NewUserHandler(userUseCase) // Handlers are responsible for processing HTTP requests and responses.
	log.Println("User components initialized")

	// Admin components
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

	// Product components
	productRepo := postgres.NewProductRepository(db)
	productUseCase := usecase.NewProductUseCase(productRepo)
	productHandler := handlers.NewProductHandler(productUseCase)

	router := httpDelivery.NewRouter(
		userHandler,
		adminHandler,
		categoryHandler,
		subCategoryHandler,
		productHandler,
	)
	log.Println("Router initialized")

	return &Server{
		router: router,
	}
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}
