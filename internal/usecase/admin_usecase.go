package usecase

import (
	"context"
	"errors"

	"github.com/mohamedfawas/rmshop-clean-architecture/internal/repository"
	"github.com/mohamedfawas/rmshop-clean-architecture/pkg/auth"
)

var (
	ErrAdminNotFound           = errors.New("admin not found")
	ErrInvalidAdminCredentials = errors.New("invalid admin credentials")
)

type AdminUseCase interface {
	Login(ctx context.Context, username, password string) (string, error)
}

type adminUseCase struct {
	adminRepo repository.AdminRepository
}

func NewAdminUseCase(adminRepo repository.AdminRepository) AdminUseCase {
	return &adminUseCase{adminRepo: adminRepo}
}

func (u *adminUseCase) Login(ctx context.Context, username, password string) (string, error) {
	admin, err := u.adminRepo.GetByUsername(ctx, username)
	if err != nil {
		if err == ErrAdminNotFound {
			return "", ErrInvalidAdminCredentials
		}
		return "", err
	}

	if !admin.CheckPassword(password) {
		return "", ErrInvalidAdminCredentials
	}

	// Generate JWT token
	token, err := auth.GenerateToken(admin.ID)
	if err != nil {
		return "", err
	}
	return token, nil
}
