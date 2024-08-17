package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/mohamedfawas/rmshop-clean-architecture/internal/repository"
	"github.com/mohamedfawas/rmshop-clean-architecture/pkg/auth"
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
	// Attempt to retrieve the admin by username from the repository
	admin, err := u.adminRepo.GetByUsername(ctx, username)
	if err != nil {
		if err == ErrAdminNotFound {
			// If the admin is not found, return an invalid credentials error
			return "", ErrInvalidAdminCredentials
		}
		// For any other error, return it as is
		return "", err
	}

	// Check if the provided password matches the stored password
	if !admin.CheckPassword(password) {
		// If passwords don't match, return an invalid credentials error
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
	_, role, err := auth.ValidateTokenWithRole(token)
	if err != nil {
		return ErrInvalidAdminToken
	}

	if role != "admin" {
		return ErrInvalidAdminToken
	}

	// Check if token is already blacklisted
	blacklisted, err := u.adminRepo.IsTokenBlacklisted(ctx, token)
	if err != nil {
		return err
	}
	if blacklisted {
		return ErrTokenAlreadyBlacklisted
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
