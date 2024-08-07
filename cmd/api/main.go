package main

import (
	"log"
	"net/http"

	"github.com/mohamedfawas/rmshop-clean-architecture/internal/config"
	"github.com/mohamedfawas/rmshop-clean-architecture/internal/server"
	"github.com/mohamedfawas/rmshop-clean-architecture/pkg/database"
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
	defer db.Close()

	// Run migrations , Migrations are run to ensure the database schema is up to date.
	err = database.RunMigrations(db, "./migrations")
	if err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	srv := server.NewServer(db)

	log.Printf("Starting server on : %s", cfg.Server.Port)
	if err := http.ListenAndServe(":"+cfg.Server.Port, srv); err != nil {
		log.Fatalf("Failed to start the server : %v", err)
	}
}
