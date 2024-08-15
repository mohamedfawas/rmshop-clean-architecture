package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
	"github.com/mohamedfawas/rmshop-clean-architecture/internal/config"
	"github.com/mohamedfawas/rmshop-clean-architecture/pkg/database"
)

func main() {
	// Define command-line flags
	direction := flag.String("direction", "", "Migration direction: up or down")
	steps := flag.Int("steps", 0, "Number of migrations to apply (0 means all)")
	force := flag.Int("force", -1, "Force set version (use with caution)")
	flag.Parse()

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

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		log.Fatalf("Failed to create database driver: %v", err)
	}

	// Get the current working directory
	currentDir, err := os.Getwd()
	if err != nil {
		log.Fatalf("Failed to get current directory: %v", err)
	}

	// Construct the migration source URL
	migrationPath := filepath.Join(currentDir, "migrations")
	migrationSource := fmt.Sprintf("file://%s", filepath.ToSlash(migrationPath))
	fmt.Printf("Migration source: %s\n", migrationSource)

	m, err := migrate.NewWithDatabaseInstance(
		migrationSource,
		"postgres", driver)
	if err != nil {
		log.Fatalf("Failed to create migrate instance: %v", err)
	}

	// Check current version before migration
	version, dirty, err := m.Version()
	if err != nil && err != migrate.ErrNilVersion {
		log.Fatalf("Failed to get migration version: %v", err)
	}

	fmt.Printf("Current migration version: %d\n", version)
	fmt.Printf("Is the database in a dirty state? %v\n", dirty)

	// Force set version if specified
	if *force >= 0 {
		err = m.Force(*force)
		if err != nil {
			log.Fatalf("Failed to force version: %v", err)
		}
		fmt.Printf("Forced version to: %d\n", *force)
		return
	}

	// Run migrations based on the direction flag
	switch *direction {
	case "up":
		if *steps > 0 {
			err = m.Steps(*steps)
		} else {
			err = m.Up()
		}
	case "down":
		if *steps > 0 {
			err = m.Steps(-(*steps))
		} else {
			err = m.Down()
		}
	case "":
		fmt.Println("No direction specified. Use -direction=up or -direction=down")
		return
	default:
		log.Fatal("Invalid direction. Use 'up' or 'down'.")
	}

	if err != nil && err != migrate.ErrNoChange {
		log.Fatalf("Migration failed: %v", err)
	}

	// Check version after migration
	newVersion, newDirty, err := m.Version()
	if err != nil && err != migrate.ErrNilVersion {
		log.Fatalf("Failed to get new migration version: %v", err)
	}

	fmt.Printf("New migration version: %d\n", newVersion)
	fmt.Printf("Is the database in a dirty state? %v\n", newDirty)

	fmt.Println("Migration completed successfully")
}
