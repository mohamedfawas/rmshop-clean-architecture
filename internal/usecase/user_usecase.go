package usecase

import (
	"context"
	"errors"
	"fmt"
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
	ErrInvalidInput         = errors.New("invalid input")
	ErrDatabaseUnavailable  = errors.New("database unavailable")
	ErrSMTPServerIssue      = errors.New("SMTP server issue")
)

// UserUseCase defines the interface for user-related use cases
type UserUseCase interface {
	Login(ctx context.Context, email, password string) (string, error)
	Logout(ctx context.Context, token string) error
	InitiateSignUp(ctx context.Context, user *domain.User) error
	VerifyOTP(ctx context.Context, email, otp string) error
	ResendOTP(ctx context.Context, email string) error
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

	// Check if a verified user already exists with this email
	existingUser, err := u.userRepo.GetByEmail(ctx, user.Email)
	if err == nil && existingUser.IsEmailVerified {
		log.Printf("Error checking existing user: %v", err)
		return ErrDuplicateEmail
	}

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return ErrInvalidInput
	}
	//user.PasswordHash = string(hashedPassword)
	user.Password = "" // Clear the plain text password

	// Generate OTP
	otp, err := otputil.GenerateOTP(6) //generate an otp of length 6
	if err != nil {
		log.Printf("Error generating OTP: %v", err)
		return err
	}
	log.Printf("OTP generated for email: %s", user.Email)

	//otp expiration time : 15 minute
	expiresAt := time.Now().Add(15 * time.Minute)

	// Create a temporary verification entry
	verificationEntry := &domain.VerificationEntry{
		Email:        user.Email,
		OTPCode:      otp,
		UserData:     user,
		PasswordHash: string(hashedPassword), // Store the hashed password here
		ExpiresAt:    expiresAt,
		IsVerified:   false,
	}

	err = u.userRepo.CreateVerificationEntry(ctx, verificationEntry)
	if err != nil {
		log.Printf("Error creating verification entry: %v", err)
		return fmt.Errorf("error creating verification entry: %w", err)
	}

	// Send OTP email
	err = u.emailSender.SendOTP(user.Email, otp)
	if err != nil {
		log.Printf("Error sending OTP email: %v", err)
		return fmt.Errorf("error sending OTP email: %w", err)
	}

	return nil
}

func (u *userUseCase) VerifyOTP(ctx context.Context, email, otp string) error {
	verificationEntry, err := u.userRepo.GetVerificationEntryByEmail(ctx, email)
	if err != nil {
		return ErrOTPNotFound
	}

	if verificationEntry.OTPCode != otp {
		return ErrInvalidOTP
	}

	if time.Now().After(verificationEntry.ExpiresAt) {
		return ErrExpiredOTP
	}

	// Create the user account
	user := verificationEntry.UserData
	user.IsEmailVerified = true
	user.PasswordHash = verificationEntry.PasswordHash

	// Ensure the hashed password is set
	if user.PasswordHash == "" {
		return errors.New("password hash is missing")
	}

	err = u.userRepo.Create(ctx, user)
	if err != nil {
		return ErrDatabaseUnavailable
	}

	// Mark the verification entry as verified
	verificationEntry.IsVerified = true
	err = u.userRepo.UpdateVerificationEntry(ctx, verificationEntry)
	if err != nil {
		return ErrDatabaseUnavailable
	}

	return nil
}

func (u *userUseCase) ResendOTP(ctx context.Context, email string) error {
	// Check if there's an existing unverified entry
	existingEntry, err := u.userRepo.GetVerificationEntryByEmail(ctx, email)
	if err != nil {
		if err == ErrUserNotFound {
			return ErrUserNotFound
		}
		return err
	}

	if existingEntry.IsVerified {
		return ErrEmailAlreadyVerified
	}

	// Generate new OTP
	newOTP, err := otputil.GenerateOTP(6)
	if err != nil {
		return ErrInvalidInput
	}

	// Update existing entry or create a new one
	existingEntry.OTPCode = newOTP
	existingEntry.ExpiresAt = time.Now().Add(15 * time.Minute)

	err = u.userRepo.UpdateVerificationEntry(ctx, existingEntry)
	if err != nil {
		return ErrDatabaseUnavailable
	}

	// Send new OTP email
	err = u.emailSender.SendOTP(email, newOTP)
	if err != nil {
		return ErrSMTPServerIssue
	}

	return nil
}
