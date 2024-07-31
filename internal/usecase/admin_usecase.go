package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/mohamedfawas/rmshop-clean-architecture/internal/repository"
	"github.com/mohamedfawas/rmshop-clean-architecture/pkg/auth"
)

var (
	ErrAdminNotFound           = errors.New("admin not found")
	ErrInvalidAdminCredentials = errors.New("invalid admin credentials")
	ErrInvalidAdminToken       = errors.New("invalid admin token")
)

type AdminUseCase interface {
	Login(ctx context.Context, username, password string) (string, error)
	Logout(ctx context.Context, token string) error
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

	// Generate JWT token with admin role
	token, err := auth.GenerateTokenWithRole(admin.ID, "admin")
	if err != nil {
		return "", err
	}
	return token, nil
}

func (u *adminUseCase) Logout(ctx context.Context, token string) error {
	// Validate the token
	_, err := auth.ValidateToken(token)
	if err != nil {
		return ErrInvalidAdminToken
	}

	// Get token expiration time
	claims, err := auth.GetTokenClaims(token)
	if err != nil {
		return ErrInvalidAdminToken
	}

	// Convert the expiration time to int64
	expFloat, ok := claims["exp"].(float64)
	if !ok {
		return errors.New("invalid expiration claim")
	}
	expiresAt := time.Unix(int64(expFloat), 0)

	// Blacklist the token
	return u.adminRepo.BlacklistToken(ctx, token, expiresAt)
}
