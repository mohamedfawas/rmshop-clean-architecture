package repository

import (
	"context"

	"github.com/mohamedfawas/rmshop-clean-architecture/internal/domain"
)

type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error //Create user method
	//GetByID(ctx context.Context, id int64) (*domain.User, error) //Get user details using user id
	//GetByEmail(ctx context.Context, email string) (*domain.User, error) //Get user details using email
	//Update(ctx context.Context, user *domain.User) error                //Update user details
	//Delete(ctx context.Context, id int64) error                         //Delete user details
	//UpdateLastLogin(ctx context.Context, userID int64) error //ecord the most recent time a user successfully authenticated (logged in) to the system
}
