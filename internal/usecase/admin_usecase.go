package usecase

import (
	"context"
	"time"

	"github.com/mohamedfawas/rmshop-clean-architecture/internal/repository"
	"github.com/mohamedfawas/rmshop-clean-architecture/pkg/auth"
	"github.com/mohamedfawas/rmshop-clean-architecture/pkg/utils"
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
		if err == utils.ErrAdminNotFound {
			// If the admin is not found, return an invalid credentials error
			return "", utils.ErrInvalidAdminCredentials
		}
		// For any other error, return it as is
		return "", err
	}

	// Check if the provided password matches the stored password
	if !admin.CheckPassword(password) {
		// If passwords don't match, return an invalid credentials error
		return "", utils.ErrInvalidAdminCredentials
	}

	// Generate JWT token with admin role
	token, err := auth.GenerateTokenWithRole(admin.ID, "admin")
	if err != nil {
		return "", utils.ErrGenerateJWTTokenWithRole
	}
	return token, nil
}

func (u *adminUseCase) Logout(ctx context.Context, token string) error {
	// Since we don't have access to TokenBlacklist here, we'll need to use the repository to blacklist the token
	claims, err := auth.GetClaimsFromToken(token)
	if err != nil {
		return utils.ErrInvalidToken
	}

	expFloat, ok := claims["exp"].(float64)
	if !ok {
		return utils.ErrInvalidExpirationClaim
	}
	expiresAt := time.Unix(int64(expFloat), 0)

	return u.adminRepo.BlacklistToken(ctx, token, expiresAt)
}
