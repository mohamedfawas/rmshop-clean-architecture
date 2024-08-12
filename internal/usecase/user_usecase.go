package usecase

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/mohamedfawas/rmshop-clean-architecture/internal/domain"
	"github.com/mohamedfawas/rmshop-clean-architecture/internal/repository"
	"github.com/mohamedfawas/rmshop-clean-architecture/pkg/auth"
	email "github.com/mohamedfawas/rmshop-clean-architecture/pkg/emailVerify"
	otputil "github.com/mohamedfawas/rmshop-clean-architecture/pkg/otpUtility"
	"golang.org/x/crypto/bcrypt"
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
	token, err := auth.GenerateTokenWithRole(user.ID, "user")
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
	log.Printf("Initiating sign up for email: %s", user.Email)

	// Check if user already exists
	_, err := u.userRepo.GetByEmail(ctx, user.Email)
	if err == nil {
		log.Printf("User with email %s already exists", user.Email)
		return ErrDuplicateEmail
	} else if err != ErrUserNotFound {
		log.Printf("Error checking existing user: %v", err)
		return err
	}

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.PasswordHash = string(hashedPassword)
	user.Password = "" // Clear the plain text password

	// Generate OTP
	otp, err := otputil.GenerateOTP(6)
	if err != nil {
		log.Printf("Error generating OTP: %v", err)
		return err
	}
	log.Printf("OTP generated for email: %s", user.Email)

	expiresAt := time.Now().Add(30 * time.Second)

	// Create user with unverified email
	user.IsEmailVerified = false
	err = u.userRepo.Create(ctx, user)
	if err != nil {
		log.Printf("Error creating user: %v", err)
		return err
	}
	log.Printf("User created with ID: %d", user.ID)

	// Create OTP entry
	otpEntry := &domain.OTP{
		UserID:    user.ID,
		Email:     user.Email,
		OTPCode:   otp,
		ExpiresAt: expiresAt,
	}
	err = u.userRepo.CreateOTP(ctx, otpEntry)
	if err != nil {
		log.Printf("Error creating OTP entry: %v", err)
		return err
	}
	log.Printf("OTP entry created for user ID: %d", user.ID)

	// Send OTP email
	err = u.emailSender.SendOTP(user.Email, otp)
	if err != nil {
		log.Printf("Error sending OTP email: %v", err)
		return err
	}
	log.Printf("OTP email sent to: %s", user.Email)

	return nil
}

func (u *userUseCase) VerifyOTP(ctx context.Context, email, otp string) error {
	log.Printf("Attempting to verify OTP for email: %s", email)

	otpEntry, err := u.userRepo.GetOTPByEmail(ctx, email)
	if err != nil {
		log.Printf("Error retrieving OTP for email %s: %v", email, err)
		return err
	}
	log.Printf("Retrieved OTP entry for email %s", email)

	if otpEntry.OTPCode != otp {
		log.Printf("Invalid OTP provided for email %s", email)
		return ErrInvalidOTP
	}

	if time.Now().After(otpEntry.ExpiresAt) {
		log.Printf("Expired OTP for email %s", email)
		return ErrExpiredOTP
	}

	log.Printf("OTP verified successfully for email %s", email)

	// Mark email as verified
	err = u.userRepo.UpdateEmailVerificationStatus(ctx, otpEntry.UserID, true)
	if err != nil {
		log.Printf("Error updating email verification status for user ID %d: %v", otpEntry.UserID, err)
		return err
	}

	// Delete OTP entry
	err = u.userRepo.DeleteOTP(ctx, email)
	if err != nil {
		log.Printf("Error deleting OTP entry for email %s: %v", email, err)
		return err
	}

	log.Printf("OTP verification process completed successfully for email %s", email)
	return nil
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
