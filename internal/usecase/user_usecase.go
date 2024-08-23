package usecase

import (
	"context"
	"time"

	"github.com/mohamedfawas/rmshop-clean-architecture/internal/domain"
	"github.com/mohamedfawas/rmshop-clean-architecture/internal/repository"
	"github.com/mohamedfawas/rmshop-clean-architecture/pkg/auth"
	email "github.com/mohamedfawas/rmshop-clean-architecture/pkg/emailVerify"
	otputil "github.com/mohamedfawas/rmshop-clean-architecture/pkg/otpUtility"
	"github.com/mohamedfawas/rmshop-clean-architecture/pkg/utils"
	"golang.org/x/crypto/bcrypt"
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
		if err == utils.ErrUserNotFound {
			// If the user is not found, return an invalid credentials error
			return "", utils.ErrInvalidCredentials
		}
		// For any other error, return it as is
		return "", err
	}

	if user.IsBlocked {
		return "", utils.ErrUserBlocked
	}

	// Check if the provided password matches the stored password
	if !user.CheckPassword(password) {
		// If passwords don't match, return an invalid credentials error
		return "", utils.ErrInvalidCredentials
	}
	// Update the user's last login time
	err = u.userRepo.UpdateLastLogin(ctx, user.ID)
	if err != nil {
		// If updating last login fails, return the error
		return "", utils.ErrUpdateLastLogin
	}

	// Generate a JWT token for the authenticated user
	token, err := auth.GenerateTokenWithRole(user.ID, "user")
	if err != nil {
		// If token generation fails, return the error
		return "", utils.ErrGenerateJWTTokenWithRole
	}
	// Return the generated token
	return token, nil
}

func (u *userUseCase) Logout(ctx context.Context, token string) error {
	// Validate the token
	claims, err := auth.ValidateUserToken(token)
	if err != nil {
		return utils.ErrInvalidToken
	}

	// Check if the token is already blacklisted
	isBlacklisted, err := u.userRepo.IsTokenBlacklisted(ctx, token)
	if err != nil {
		return utils.ErrFailedToCheckBlacklisted
	}
	if isBlacklisted {
		return utils.ErrTokenAlreadyBlacklisted
	}

	// Get token expiration time
	exp, ok := claims["exp"].(float64)
	if !ok {
		return utils.ErrInvalidToken
	}
	expiresAt := time.Unix(int64(exp), 0)

	// Blacklist the token
	return u.userRepo.BlacklistToken(ctx, token, expiresAt)
}

func (u *userUseCase) InitiateSignUp(ctx context.Context, user *domain.User) error {
	existingUser, err := u.userRepo.GetByEmail(ctx, user.Email) // Check if a verified user (which is not soft deleted) already exists with this email
	if err == nil && existingUser.IsEmailVerified {
		return utils.ErrDuplicateEmail
	}

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return utils.ErrHashingPassword
	}
	user.Password = "" // Clear the plain text password

	// Generate OTP
	otp, err := otputil.GenerateOTP(6) //generate an otp of length 6
	if err != nil {
		//log.Printf("Error generating OTP: %v", err)
		return utils.ErrGenerateOTP
	}
	//log.Printf("OTP generated for email: %s", user.Email)

	//otp expiration time : 15 minutes
	expiresAt := time.Now().UTC().Add(15 * time.Minute)

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
		return utils.ErrCreateVericationEntry
	}

	// Send OTP email
	err = u.emailSender.SendOTP(user.Email, otp)
	if err != nil {
		// log.Printf("Error sending OTP email: %v", err)
		return utils.ErrSendingOTP
	}

	return nil
}

func (u *userUseCase) VerifyOTP(ctx context.Context, email, otp string) error {
	// Get the verification entry
	entry, err := u.userRepo.GetVerificationEntryByEmail(ctx, email)
	if err != nil {
		if err == utils.ErrVerificationEntryNotFound {
			return utils.ErrNonExEmail
		}
		return err
	}

	// Check if OTP is expired
	if time.Now().UTC().After(entry.ExpiresAt) {
		return utils.ErrExpiredOTP
	}

	// Verify OTP
	if entry.OTPCode != otp {
		return utils.ErrInvalidOTP
	}

	// Check if the email is already verified
	if entry.IsVerified {
		return utils.ErrEmailAlreadyVerified
	}

	// Create the user
	user := entry.UserData
	user.IsEmailVerified = true
	user.PasswordHash = entry.PasswordHash

	err = u.userRepo.Create(ctx, user)
	if err != nil {
		return utils.ErrCreateUser
	}

	// Mark the verification entry as verified
	entry.IsVerified = true
	err = u.userRepo.UpdateVerificationEntry(ctx, entry)
	if err != nil {
		return utils.ErrUpdateVerificationEntry
	}

	// delete the verification entry
	err = u.userRepo.DeleteVerificationEntry(ctx, email)
	if err != nil {
		return utils.ErrDeleteVerificationEntry
	}

	return nil
}

func (u *userUseCase) ResendOTP(ctx context.Context, email string) error {
	// Get the verification entry
	entry, err := u.userRepo.GetVerificationEntryByEmail(ctx, email)
	if err != nil {
		if err == utils.ErrVerificationEntryNotFound {
			return utils.ErrNonExEmail
		}
		return err
	}

	// Check if signup process has expired
	if time.Now().UTC().Sub(entry.CreatedAt) > utils.SignupExpiration {
		return utils.ErrSignupExpired
	}

	// Check rate limiting
	resendCount, lastResendTime, err := u.userRepo.GetOTPResendInfo(ctx, email)
	if err != nil {
		return utils.ErrRetrieveOTPResendInfo
	}

	if resendCount >= utils.MaxResendAttempts && time.Now().UTC().Sub(lastResendTime) < utils.ResendCooldown {
		return utils.ErrTooManyResendAttempts
	}

	// Generate new OTP
	newOTP, err := otputil.GenerateOTP(6)
	if err != nil {
		return utils.ErrGenerateOTP
	}

	// Update verification entry
	entry.OTPCode = newOTP
	entry.ExpiresAt = time.Now().UTC().Add(15 * time.Minute)

	err = u.userRepo.UpdateVerificationEntryAfterResendOTP(ctx, entry)
	if err != nil {
		return utils.ErrUpdateVerficationAfterResend
	}

	// Update resend info
	err = u.userRepo.UpdateOTPResendInfo(ctx, email)
	if err != nil {
		return utils.ErrUpdateOTPResendTable
	}

	// Send new OTP email
	err = u.emailSender.SendOTP(email, newOTP)
	if err != nil {
		return utils.ErrSMTPServerIssue
	}

	return nil
}
