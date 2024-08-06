package postgres

import (
	"context"
	"database/sql"
	"log"
	"time"

	"github.com/lib/pq"
	"github.com/mohamedfawas/rmshop-clean-architecture/internal/domain"
	"github.com/mohamedfawas/rmshop-clean-architecture/internal/usecase"
)

// this code sets up a structure that will handle database operations related to users.
type userRepository struct {
	db *sql.DB // pointer to sql db object, represents a connection to the database.
}

// constructor function for creating new userRepository instances.
func NewUserRepository(db *sql.DB) *userRepository {
	return &userRepository{db: db} //creates a new userRepository with the provided database connection and returns it.
}

//This approach follows the dependency injection principle, where the database connection is provided from outside rather than created within the repository. This makes the code more flexible and easier to test.

func (r *userRepository) Create(ctx context.Context, user *domain.User) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Reset the sequence
	_, err = tx.ExecContext(ctx, "SELECT reset_users_id_seq()")
	if err != nil {
		log.Printf("Error resetting sequence: %v", err)
		return err
	}

	query := `INSERT INTO users (name, email, password_hash, date_of_birth, phone_number, is_blocked, created_at)
              VALUES ($1, $2, $3, $4, $5, $6, $7)
              RETURNING id, created_at`

	err = tx.QueryRowContext(ctx, query,
		user.Name, user.Email, user.PasswordHash, user.DOB, user.PhoneNumber, user.IsBlocked, user.CreatedAt).
		Scan(&user.ID, &user.CreatedAt)
	if err != nil {
		pqErr, ok := err.(*pq.Error)
		if ok && pqErr.Code == "23505" { // Unique violation error code
			return usecase.ErrDuplicateEmail
		}
		return err
	}

	if err = tx.Commit(); err != nil {
		log.Printf("Error committing transaction: %v", err)
		return err
	}

	return nil
}

func (r *userRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	// SQL query to select user details by email
	query := `SELECT id, name, email, password_hash, date_of_birth, phone_number, is_blocked, created_at, updated_at, last_login 
              FROM users WHERE email = $1`

	var user domain.User
	var lastLogin sql.NullTime
	// Execute the query and scan the result into the user struct
	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID, &user.Name, &user.Email, &user.PasswordHash, &user.DOB,
		&user.PhoneNumber, &user.IsBlocked, &user.CreatedAt, &user.UpdatedAt, &lastLogin,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			// If no user is found, return a specific error
			return nil, usecase.ErrUserNotFound
		}
		// For any other error, return it as is
		return nil, err
	}

	// If lastLogin is not null, set it in the user struct
	if lastLogin.Valid {
		user.LastLogin = lastLogin.Time
	}

	// Return the user struct and nil error
	return &user, nil
}

func (r *userRepository) UpdateLastLogin(ctx context.Context, userID int64) error {
	// SQL query to update the last_login field for a user
	query := `UPDATE users SET last_login = NOW() WHERE id = $1`

	// Execute the update query
	_, err := r.db.ExecContext(ctx, query, userID)
	return err
}

func (r *userRepository) BlacklistToken(ctx context.Context, token string, expiresAt time.Time) error {
	// SQL query to insert the blacklisted token
	query := `INSERT INTO blacklisted_tokens (token, expires_at) VALUES ($1, $2)`
	// Execute the insert query
	_, err := r.db.ExecContext(ctx, query, token, expiresAt)
	return err
}

func (r *userRepository) IsTokenBlacklisted(ctx context.Context, token string) (bool, error) {
	// SQL query to check if a token is blacklisted and not expired
	query := `SELECT EXISTS(SELECT 1 FROM blacklisted_tokens WHERE token = $1 AND expires_at > NOW())`
	var exists bool
	// Execute the query and scan the result
	err := r.db.QueryRowContext(ctx, query, token).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}
