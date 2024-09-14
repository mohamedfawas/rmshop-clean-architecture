package server

import (
	"database/sql"
	"log"
	"net/http"
	"path/filepath"
	"text/template"

	"github.com/mohamedfawas/rmshop-clean-architecture/internal/config"
	httpDelivery "github.com/mohamedfawas/rmshop-clean-architecture/internal/delivery/http"
	"github.com/mohamedfawas/rmshop-clean-architecture/internal/delivery/http/handlers"
	"github.com/mohamedfawas/rmshop-clean-architecture/internal/repository/postgres"
	"github.com/mohamedfawas/rmshop-clean-architecture/internal/usecase"
	"github.com/mohamedfawas/rmshop-clean-architecture/pkg/auth"
	"github.com/mohamedfawas/rmshop-clean-architecture/pkg/cloudinary"
	email "github.com/mohamedfawas/rmshop-clean-architecture/pkg/emailVerify"
	"github.com/mohamedfawas/rmshop-clean-architecture/pkg/payment/razorpay"
)

// Server struct holds the router which will be used to handle HTTP requests
type Server struct {
	router    http.Handler
	templates *template.Template
}

// NewServer creates and returns a new Server instance
func NewServer(db *sql.DB, emailSender *email.Sender, cloudinaryService *cloudinary.CloudinaryService, tokenBlacklist *auth.TokenBlacklist, cfg *config.Config) *Server {
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

	// wishlist
	wishlistRepo := postgres.NewWishlistRepository(db)
	wishlistUseCase := usecase.NewWishlistUseCase(wishlistRepo, productRepo, userRepo)
	wishlistHandler := handlers.NewWishlistHandler(wishlistUseCase)

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

	razorpayService := razorpay.NewService(cfg.Razorpay.KeyID, cfg.Razorpay.KeySecret)

	checkoutUseCase := usecase.NewCheckoutUseCase(checkoutRepo, productRepo, cartRepo, couponRepo, userRepo, orderRepo, razorpayService)
	checkoutHandler := handlers.NewCheckoutHandler(checkoutUseCase, couponUseCase)
	log.Println("Checkout components initialized")

	walletRepo := postgres.NewWalletRepository(db)
	walletUseCase := usecase.NewWalletUseCase(walletRepo, userRepo)
	walletHandler := handlers.NewWalletHandler(walletUseCase)
	log.Println("wallet components initialized")

	orderUseCase := usecase.NewOrderUseCase(orderRepo, checkoutRepo, productRepo, cartRepo, cfg.Razorpay.KeySecret, cfg.Razorpay.KeySecret)
	orderHandler := handlers.NewOrderHandler(orderUseCase)
	log.Println("Order components initialized")

	inventoryRepo := postgres.NewInventoryRepository(db)
	inventoryUseCase := usecase.NewInventoryUseCase(inventoryRepo, productRepo)
	inventoryHandler := handlers.NewInventoryHandler(inventoryUseCase)
	log.Println("Inventory components initialized")

	templates := setupTemplates()
	paymentHandler := handlers.NewPaymentHandler(orderUseCase, cfg.Razorpay.KeyID, cfg.Razorpay.KeySecret, templates)

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
		inventoryHandler,
		paymentHandler,
		wishlistHandler,
		walletHandler,
		templates,
	)
	log.Println("Router initialized")

	// Return a new Server instance with the initialized router
	return &Server{
		router:    router,
		templates: templates,
	}
}

// ServeHTTP makes the Server struct implement the http.Handler interface
// This method is called for every HTTP request to the server
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

func setupTemplates() *template.Template {
	templatesDir := "./static/html" // Adjust this path as needed
	pattern := filepath.Join(templatesDir, "*.html")
	return template.Must(template.ParseGlob(pattern))
}
