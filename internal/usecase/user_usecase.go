package usecase

import (
	"context"

	"github.com/mohamedfawas/rmshop-clean-architecture/internal/domain"
	"github.com/mohamedfawas/rmshop-clean-architecture/internal/repository"
)

// UserUseCase defines the interface for user-related use cases
type UserUseCase interface {
	Register(ctx context.Context, user *domain.User) error
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
	return u.userRepo.Create(ctx, user)
}
