package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/mohamedfawas/rmshop-clean-architecture/internal/domain"
	"github.com/mohamedfawas/rmshop-clean-architecture/internal/repository"
	"github.com/mohamedfawas/rmshop-clean-architecture/pkg/auth"
)

var (
	ErrDuplicateEmail     = errors.New("email already exists")
	ErrUserNotFound       = errors.New("user not found")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidToken       = errors.New("invalid token")
)

// UserUseCase defines the interface for user-related use cases
type UserUseCase interface {
	Register(ctx context.Context, user *domain.User) error
	Login(ctx context.Context, email, password string) (string, error)
	Logout(ctx context.Context, token string) error
	// Add other user-related use case methods here as needed, for example:
	// GetByID(ctx context.Context, id int64) (*domain.User, error)
	// Update(ctx context.Context, user *domain.User) error
	// Delete(ctx context.Context, id int64) error
}

// userUseCase implements the UserUseCase interface
type userUseCase struct {
	userRepo repository.UserRepository
}

// NewUserUseCase creates a new instance of UserUseCase
func NewUserUseCase(userRepo repository.UserRepository) UserUseCase {
	return &userUseCase{userRepo: userRepo}
}

// Register implements the user registration use case
func (u *userUseCase) Register(ctx context.Context, user *domain.User) error {
	// Add any business logic here (e.g., validation)
	err := u.userRepo.Create(ctx, user)
	if err != nil {
		if err == ErrDuplicateEmail {
			return ErrDuplicateEmail
		}
		return err
	}
	return nil
}

func (u *userUseCase) Login(ctx context.Context, email, password string) (string, error) {
	user, err := u.userRepo.GetByEmail(ctx, email)
	if err != nil {
		if err == ErrUserNotFound {
			return "", ErrInvalidCredentials
		}
		return "", err
	}

	if !user.CheckPassword(password) {
		return "", ErrInvalidCredentials
	}
	err = u.userRepo.UpdateLastLogin(ctx, user.ID)
	if err != nil {
		return "", err
	}

	// Generate JWT token
	token, err := auth.GenerateToken(user.ID)
	if err != nil {
		return "", err
	}
	return token, nil
}

func (u *userUseCase) Logout(ctx context.Context, token string) error {
	// Validate the token
	_, err := auth.ValidateToken(token)
	if err != nil {
		return ErrInvalidToken
	}

	// Get token expiration time
	claims, err := auth.GetTokenClaims(token)
	if err != nil {
		return ErrInvalidToken
	}

	// Convert the expiration time to int64
	expFloat, ok := claims["exp"].(float64)
	if !ok {
		return errors.New("invalid expiration claim")
	}
	expiresAt := time.Unix(int64(expFloat), 0)

	// Blacklist the token
	return u.userRepo.BlacklistToken(ctx, token, expiresAt)
}
