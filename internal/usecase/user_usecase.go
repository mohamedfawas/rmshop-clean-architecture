package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/mohamedfawas/rmshop-clean-architecture/internal/domain"
	"github.com/mohamedfawas/rmshop-clean-architecture/internal/repository"
	"github.com/mohamedfawas/rmshop-clean-architecture/pkg/auth"
	email "github.com/mohamedfawas/rmshop-clean-architecture/pkg/emailVerify"
	otputil "github.com/mohamedfawas/rmshop-clean-architecture/pkg/otpUtility"
)

var (
	ErrDuplicateEmail       = errors.New("email already exists")
	ErrUserNotFound         = errors.New("user not found")
	ErrInvalidCredentials   = errors.New("invalid credentials")
	ErrInvalidToken         = errors.New("invalid token")
	ErrOTPNotFound          = errors.New("OTP not found")
	ErrInvalidOTP           = errors.New("invalid OTP")
	ErrExpiredOTP           = errors.New("OTP has expired")
	ErrEmailAlreadyVerified = errors.New("email already verified")
)

// UserUseCase defines the interface for user-related use cases
type UserUseCase interface {
	Register(ctx context.Context, user *domain.User) error
	Login(ctx context.Context, email, password string) (string, error)
	Logout(ctx context.Context, token string) error
	InitiateSignUp(ctx context.Context, user *domain.User) error
	VerifyOTP(ctx context.Context, email, otp string) error
	ResendOTP(ctx context.Context, email string) error
	// Add other user-related use case methods here as needed, for example:
	// GetByID(ctx context.Context, id int64) (*domain.User, error)
	// Update(ctx context.Context, user *domain.User) error
	// Delete(ctx context.Context, id int64) error
}

// userUseCase implements the UserUseCase interface
type userUseCase struct {
	userRepo    repository.UserRepository
	emailSender *email.Sender
}

// NewUserUseCase creates a new instance of UserUseCase
func NewUserUseCase(userRepo repository.UserRepository, emailSender *email.Sender) UserUseCase {
	return &userUseCase{userRepo: userRepo,
		emailSender: emailSender}
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
	// Attempt to retrieve the user by email from the repository
	user, err := u.userRepo.GetByEmail(ctx, email)
	if err != nil {
		if err == ErrUserNotFound {
			// If the user is not found, return an invalid credentials error
			return "", ErrInvalidCredentials
		}
		// For any other error, return it as is
		return "", err
	}

	// Check if the provided password matches the stored password
	if !user.CheckPassword(password) {
		// If passwords don't match, return an invalid credentials error
		return "", ErrInvalidCredentials
	}
	// Update the user's last login time
	err = u.userRepo.UpdateLastLogin(ctx, user.ID)
	if err != nil {
		// If updating last login fails, return the error
		return "", err
	}

	// Generate a JWT token for the authenticated user
	token, err := auth.GenerateToken(user.ID)
	if err != nil {
		// If token generation fails, return the error
		return "", err
	}
	// Return the generated token
	return token, nil
}

func (u *userUseCase) Logout(ctx context.Context, token string) error {
	// Validate the token
	_, err := auth.ValidateToken(token)
	if err != nil {
		// If the token is invalid, return an error
		return ErrInvalidToken
	}

	// Get token expiration time
	claims, err := auth.GetTokenClaims(token)
	if err != nil {
		// If unable to get token claims, return an error
		return ErrInvalidToken
	}

	// Convert the expiration time to int64
	expFloat, ok := claims["exp"].(float64)
	if !ok {
		// If the expiration claim is not a float64, return an error
		return errors.New("invalid expiration claim")
	}
	expiresAt := time.Unix(int64(expFloat), 0)

	// Blacklist the token
	return u.userRepo.BlacklistToken(ctx, token, expiresAt)
}

func (u *userUseCase) InitiateSignUp(ctx context.Context, user *domain.User) error {
	// Check if user already exists
	_, err := u.userRepo.GetByEmail(ctx, user.Email)
	if err == nil {
		return ErrDuplicateEmail
	} else if err != ErrUserNotFound {
		return err
	}

	// Generate OTP
	otp, err := otputil.GenerateOTP(6)
	if err != nil {
		return err
	}
	expiresAt := time.Now().Add(15 * time.Minute)

	// Create user with unverified email
	user.IsEmailVerified = false
	err = u.userRepo.Create(ctx, user)
	if err != nil {
		return err
	}

	// Create OTP entry
	otpEntry := &domain.OTP{
		UserID:    user.ID,
		Email:     user.Email,
		OTPCode:   otp,
		ExpiresAt: expiresAt,
	}
	err = u.userRepo.CreateOTP(ctx, otpEntry)
	if err != nil {
		return err
	}

	// Send OTP email
	return u.emailSender.SendOTP(user.Email, otp)
}

func (u *userUseCase) VerifyOTP(ctx context.Context, email, otp string) error {
	otpEntry, err := u.userRepo.GetOTPByEmail(ctx, email)
	if err != nil {
		if err == ErrOTPNotFound {
			return ErrOTPNotFound
		}
		return err
	}

	if otpEntry.OTPCode != otp {
		return ErrInvalidOTP
	}

	if time.Now().After(otpEntry.ExpiresAt) {
		return ErrExpiredOTP
	}

	// Mark email as verified
	err = u.userRepo.UpdateEmailVerificationStatus(ctx, otpEntry.UserID, true)
	if err != nil {
		return err
	}

	// Delete OTP entry
	return u.userRepo.DeleteOTP(ctx, email)
}

func (u *userUseCase) ResendOTP(ctx context.Context, email string) error {
	user, err := u.userRepo.GetByEmail(ctx, email)
	if err != nil {
		if err == ErrUserNotFound {
			return ErrUserNotFound
		}
		return err
	}

	if user.IsEmailVerified {
		return ErrEmailAlreadyVerified
	}

	// Delete existing OTP if any
	_ = u.userRepo.DeleteOTP(ctx, email)

	// Generate new OTP
	otp, err := otputil.GenerateOTP(6)
	if err != nil {
		return err
	}
	expiresAt := time.Now().Add(15 * time.Minute)

	// Create new OTP entry
	otpEntry := &domain.OTP{
		UserID:    user.ID,
		Email:     user.Email,
		OTPCode:   otp,
		ExpiresAt: expiresAt,
	}
	err = u.userRepo.CreateOTP(ctx, otpEntry)
	if err != nil {
		return err
	}

	// Send OTP email
	return u.emailSender.SendOTP(email, otp)
}
