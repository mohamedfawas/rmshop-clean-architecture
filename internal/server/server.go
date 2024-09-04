package server

import (
	"database/sql"
	"log"
	"net/http"

	httpDelivery "github.com/mohamedfawas/rmshop-clean-architecture/internal/delivery/http"
	"github.com/mohamedfawas/rmshop-clean-architecture/internal/delivery/http/handlers"
	"github.com/mohamedfawas/rmshop-clean-architecture/internal/repository/postgres"
	"github.com/mohamedfawas/rmshop-clean-architecture/internal/usecase"
	"github.com/mohamedfawas/rmshop-clean-architecture/pkg/auth"
	"github.com/mohamedfawas/rmshop-clean-architecture/pkg/cloudinary"
	email "github.com/mohamedfawas/rmshop-clean-architecture/pkg/emailVerify"
)

// Server struct holds the router which will be used to handle HTTP requests
type Server struct {
	router http.Handler
}

// NewServer creates and returns a new Server instance
func NewServer(db *sql.DB, emailSender *email.Sender, cloudinaryService *cloudinary.CloudinaryService, tokenBlacklist *auth.TokenBlacklist) *Server {
	log.Println("Initializing server components...")

	// User components initialization
	userRepo := postgres.NewUserRepository(db)
	userUseCase := usecase.NewUserUseCase(userRepo, emailSender, tokenBlacklist)
	userHandler := handlers.NewUserHandler(userUseCase)
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

	// cart components
	cartRepo := postgres.NewCartRepository(db)
	cartUseCase := usecase.NewCartUseCase(cartRepo, productRepo, userRepo)
	cartHandler := handlers.NewCartHandler(cartUseCase)
	log.Println("Cart components initialized")

	// checkour repo initialized
	checkoutRepo := postgres.NewCheckoutRepository(db)
	log.Println("Checkout repository initialized")

	orderRepo := postgres.NewOrderRepository(db)
	log.Println("Order repository initialized")

	// coupon components
	couponRepo := postgres.NewCouponRepository(db)
	couponUseCase := usecase.NewCouponUseCase(couponRepo, checkoutRepo)
	couponHandler := handlers.NewCouponHandler(couponUseCase)
	log.Println("Coupon components initialized")

	checkoutUseCase := usecase.NewCheckoutUseCase(checkoutRepo, productRepo, cartRepo, couponRepo, userRepo, orderRepo)
	checkoutHandler := handlers.NewCheckoutHandler(checkoutUseCase, couponUseCase)
	log.Println("Checkout components initialized")

	orderUseCase := usecase.NewOrderUseCase(orderRepo)
	orderHandler := handlers.NewOrderHandler(orderUseCase)
	log.Println("Order components initialized")

	// Initialize the router with all handlers
	router := httpDelivery.NewRouter(
		userHandler,
		adminHandler,
		categoryHandler,
		subCategoryHandler,
		productHandler,
		tokenBlacklist,
		cartHandler,
		couponHandler,
		checkoutHandler,
		orderHandler,
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
