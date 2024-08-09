package main

import (
	"log"
	"net/http"

	"github.com/mohamedfawas/rmshop-clean-architecture/internal/config"
	"github.com/mohamedfawas/rmshop-clean-architecture/internal/server"
	"github.com/mohamedfawas/rmshop-clean-architecture/pkg/database"
	email "github.com/mohamedfawas/rmshop-clean-architecture/pkg/emailVerify"
)

func main() {
	log.Println("Starting application...")
	//configuration is loaded from config.yaml
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	//db connection is established using the loaded configuration
	db, err := database.NewPostgresConnection(cfg.DB.Host, cfg.DB.Port, cfg.DB.User, cfg.DB.Password, cfg.DB.Name)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close() // Ensure the database connection is closed when the application exits

	// Run migrations , Migrations are run to ensure the database schema is up to date.
	err = database.RunMigrations(db, "./migrations")
	if err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Initialize email sender for OTP functionality
	emailSender := email.NewSender(cfg.SMTP.Host, cfg.SMTP.Port, cfg.SMTP.Username, cfg.SMTP.Password)

	// Create a new server instance with the database connection and email sender
	srv := server.NewServer(db, emailSender)

	// Start the HTTP server
	log.Printf("Starting server on : %s", cfg.Server.Port)
	if err := http.ListenAndServe(":"+cfg.Server.Port, srv); err != nil {
		log.Fatalf("Failed to start the server : %v", err)
	}
}
