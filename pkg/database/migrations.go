package database

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/golang-migrate/migrate/v4"
	postgresMigration "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

// RunMigrations executes all pending migrations
func RunMigrations(db *sql.DB, migrationPath string) error {
	//creating a database driver for PostgreSQL to be used with a migration tool, likely the golang-migrate/migrate library.
	driver, err := postgresMigration.WithInstance(db, &postgresMigration.Config{}) //&postgresMigration.Config{}: This is a configuration struct for the PostgreSQL driver. In this case, it's empty, meaning it's using default settings.
	//driver: This is the created PostgreSQL driver that can be used with the migration tool.
	if err != nil {
		return fmt.Errorf("could not create the postgres driver: %v", err)
	}

	m, err := migrate.NewWithDatabaseInstance( //function is part of the golang-migrate library, creates a new migration instance that will be used to apply or rollback database migrations.
		"file://"+migrationPath, // migrationPath variable should contain the path to the directory where your migration files are stored
		"postgres", driver)      // postgres : type of database you're using , driver : database driver instance that provides a connection to your PostgreSQL database.
	//m: This is the migration instance that you'll use to perform migration operations.
	if err != nil {
		return fmt.Errorf("migrate: failed to create new instance: %v", err)
	}

	// m.Up() is called to run all available "up" migrations.
	if err := m.Up(); err != nil && err != migrate.ErrNoChange { //If there is an error, but it's not migrate.ErrNoChange, it means something went wrong during the migration process.
		//migrate.ErrNoChange is a special error that indicates no migrations were applied because the database was already up to date. This is not considered a failure condition.
		return fmt.Errorf("migrate: failed to run up migrations: %v", err)
	}

	log.Println("Migrations executed successfully")
	return nil
}
