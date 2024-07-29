package server

import (
	"database/sql"
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
	// Handler -> UseCase -> Repository.
	userRepo := postgres.NewUserRepository(db)          // repository is responsible for handling data persistence and retrieval for user-related operations.
	userUseCase := usecase.NewUserUseCase(userRepo)     // Use cases contain the business logic of the application, user use case is initialized with the user repository, allowing it to perform data operations as needed.
	userHandler := handlers.NewUserHandler(userUseCase) // Handlers are responsible for processing HTTP requests and responses.

	router := httpDelivery.NewRouter(userHandler)

	return &Server{
		router: router,
	}
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}
