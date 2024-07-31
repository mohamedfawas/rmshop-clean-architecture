package adminseeder

import (
	"database/sql"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// Config holds the configuration for the admin seeder
type Config struct {
	DB       *sql.DB
	Username string
	Password string
}

// SeedAdmin creates or updates an admin account
func SeedAdmin(cfg Config) error {
	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(cfg.Password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Insert admin into the database
	_, err = cfg.DB.Exec(`
		INSERT INTO admins (username, password_hash)
		VALUES ($1, $2)
		ON CONFLICT (username) DO UPDATE
		SET password_hash = EXCLUDED.password_hash
	`, cfg.Username, string(hashedPassword))
	if err != nil {
		return fmt.Errorf("failed to insert/update admin: %w", err)
	}

	return nil
}
