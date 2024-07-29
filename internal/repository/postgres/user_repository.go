package postgres

import (
	"context"
	"database/sql"

	"github.com/mohamedfawas/rmshop-clean-architecture/internal/domain"
	"golang.org/x/crypto/bcrypt"
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
	//Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	query := `INSERT INTO users (name, email,password_hash, date_of_birth,phone_number,is_blocked, created_at)
				VALUES ($1, $2, $3, $4, $5, $6, NOW())
				RETURNING id, created_at`

	err = r.db.QueryRowContext(ctx, query,
		user.Name, user.Email, string(hashedPassword), user.DOB, user.PhoneNumber, user.IsBlocked).
		Scan(&user.ID, &user.CreatedAt)
	return err
}
