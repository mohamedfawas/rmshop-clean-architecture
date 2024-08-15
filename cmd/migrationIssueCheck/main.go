package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

func main() {
	// Replace with your actual database connection details
	db, err := sql.Open("postgres", "postgres://postgres:manojsir@1@localhost:5432/rmshop_db?sslmode=disable")
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

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

	version, dirty, err := m.Version()
	if err != nil {
		if err == migrate.ErrNilVersion {
			fmt.Println("No migrations have been applied yet.")
			return
		}
		log.Fatalf("Failed to get migration version: %v", err)
	}

	fmt.Printf("Current migration version: %d\n", version)
	fmt.Printf("Is the database in a dirty state? %v\n", dirty)
}
