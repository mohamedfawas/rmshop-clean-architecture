package database

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/golang-migrate/migrate/v4"
	postgresMigration "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

type migrationLogger struct{}

func (ml *migrationLogger) Printf(format string, v ...interface{}) {
	log.Printf(format, v...)
}

func (ml *migrationLogger) Verbose() bool {
	return true
}

// RunMigrations executes all pending migrations
func RunMigrations(db *sql.DB, migrationPath string) error {
	log.Printf("Starting migrations from path: %s", migrationPath)

	// Set the timezone to UTC before running migrations
	_, err := db.Exec("SET TIME ZONE 'UTC'")
	if err != nil {
		log.Printf("Warning: Failed to set timezone to UTC before migrations: %v", err)
	}

	//creating a database driver for PostgreSQL to be used with a migration tool, likely the golang-migrate/migrate library.
	driver, err := postgresMigration.WithInstance(db, &postgresMigration.Config{}) //&postgresMigration.Config{}: This is a configuration struct for the PostgreSQL driver. In this case, it's empty, meaning it's using default settings.
	//driver: This is the created PostgreSQL driver that can be used with the migration tool.
	if err != nil {
		log.Printf("Error creating postgres driver: %v", err)
		return fmt.Errorf("could not create the postgres driver: %v", err)
	}

	m, err := migrate.NewWithDatabaseInstance( //function is part of the golang-migrate library, creates a new migration instance that will be used to apply or rollback database migrations.
		"file://"+migrationPath, // migrationPath variable should contain the path to the directory where your migration files are stored
		"postgres", driver)      // postgres : type of database you're using , driver : database driver instance that provides a connection to your PostgreSQL database.
	//m: This is the migration instance that you'll use to perform migration operations.
	if err != nil {
		log.Printf("Error creating migration instance: %v", err)
		return fmt.Errorf("migrate: failed to create new instance: %v", err)
	}

	// Enable verbose logging
	m.Log = &migrationLogger{}

	log.Println("Migration instance created, attempting to run migrations...")
	if err := m.Up(); err != nil {
		if err == migrate.ErrNoChange {
			log.Println("No migrations to run")
			return nil
		}
		log.Printf("Error running migrations: %v", err)
		return fmt.Errorf("migrate: failed to run up migrations: %v", err)
	}

	log.Println("Migrations executed successfully")
	return nil
}
