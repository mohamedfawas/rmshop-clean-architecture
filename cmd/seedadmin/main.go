package main

import (
	"log"

	"github.com/mohamedfawas/rmshop-clean-architecture/internal/config"
	"github.com/mohamedfawas/rmshop-clean-architecture/pkg/adminseeder"
	"github.com/mohamedfawas/rmshop-clean-architecture/pkg/database"
)

func main() {
	// Load the configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Establish a database connection
	db, err := database.NewPostgresConnection(cfg.DB.Host, cfg.DB.Port, cfg.DB.User, cfg.DB.Password, cfg.DB.Name)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close() // Ensure the database connection is closed when the command exits

	// Run migrations before seeding to ensure the database schema is up to date
	err = database.RunMigrations(db, "./migrations")
	if err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Seed the admin account
	err = adminseeder.SeedAdmin(adminseeder.Config{
		DB:       db,
		Username: cfg.Admin.Username,
		Password: cfg.Admin.Password,
	})
	if err != nil {
		log.Fatalf("Failed to seed admin: %v", err)
	}

	log.Println("Migrations run and admin seeded successfully")
}
