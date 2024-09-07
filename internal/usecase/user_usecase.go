package usecase

import (
	"context"
	"log"
	"time"

	"github.com/mohamedfawas/rmshop-clean-architecture/internal/domain"
	"github.com/mohamedfawas/rmshop-clean-architecture/internal/repository"
	"github.com/mohamedfawas/rmshop-clean-architecture/pkg/auth"
	email "github.com/mohamedfawas/rmshop-clean-architecture/pkg/emailVerify"
	otputil "github.com/mohamedfawas/rmshop-clean-architecture/pkg/otpUtility"
	"github.com/mohamedfawas/rmshop-clean-architecture/pkg/utils"
	"github.com/mohamedfawas/rmshop-clean-architecture/pkg/validator"
	"golang.org/x/crypto/bcrypt"
)

// UserUseCase defines the interface for user-related use cases
type UserUseCase interface {
	Login(ctx context.Context, email, password string) (string, error)
	Logout(ctx context.Context, token string) error
	InitiateSignUp(ctx context.Context, user *domain.User) error
	VerifyOTP(ctx context.Context, email, otp string) error
	ResendOTP(ctx context.Context, email string) error
	GetUserProfile(ctx context.Context, userID int64) (*domain.User, error)                                    //fz
	UpdateProfile(ctx context.Context, userID int64, updateData *domain.UserUpdatedData) (*domain.User, error) //fz
	ForgotPassword(ctx context.Context, email string) error
	ResetPassword(ctx context.Context, email, otp, newPassword string) error   //fz
	AddUserAddress(ctx context.Context, userAddress *domain.UserAddress) error //fz
	UpdateUserAddress(ctx context.Context, userID, addressID int64, updatedAddressData *domain.UserAddressUpdate) (*domain.UserAddress, error)
	GetUserAddresses(ctx context.Context, userID int64) ([]*domain.UserAddress, error)
	DeleteUserAddress(ctx context.Context, userID, addressID int64) error
}

// userUseCase implements the UserUseCase interface
type userUseCase struct {
	userRepo       repository.UserRepository
	emailSender    email.EmailSender
	tokenBlacklist *auth.TokenBlacklist
}

// NewUserUseCase creates a new instance of UserUseCase
func NewUserUseCase(userRepo repository.UserRepository, emailSender email.EmailSender, tokenBlacklist *auth.TokenBlacklist) UserUseCase {
	return &userUseCase{userRepo: userRepo,
		emailSender:    emailSender,
		tokenBlacklist: tokenBlacklist}
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
	_, err := auth.ValidateToken(token)
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
	claims, err := auth.GetClaimsFromToken(token)
	if err != nil {
		return utils.ErrInvalidToken
	}

	exp, ok := claims["exp"].(float64)
	if !ok {
		return utils.ErrInvalidToken
	}
	expiresAt := time.Unix(int64(exp), 0)

	// Blacklist the token
	return u.userRepo.BlacklistToken(ctx, token, expiresAt)
}

func (u *userUseCase) InitiateSignUp(ctx context.Context, user *domain.User) error {
	// check if the given email is already registered for a verified user
	existingUser, err := u.userRepo.GetByEmail(ctx, user.Email)
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

	err = u.userRepo.CreateUserSignUpVerifcationEntry(ctx, verificationEntry)
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
	entry, err := u.userRepo.FindSignUpVerificationEntryByEmail(ctx, email)
	if err != nil {
		if err == utils.ErrVerificationEntryNotFound {
			return utils.ErrNonExEmail
		}
		return err
	}

	// Check if the email is already verified
	if entry.IsVerified {
		return utils.ErrEmailAlreadyVerified
	}

	// Check if OTP is expired
	if time.Now().UTC().After(entry.ExpiresAt) {
		return utils.ErrExpiredOTP
	}

	// Verify OTP
	if entry.OTPCode != otp {
		return utils.ErrInvalidOTP
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
	err = u.userRepo.UpdateSignUpVerificationEntry(ctx, entry)
	if err != nil {
		return utils.ErrUpdateVerificationEntry
	}

	// delete the verification entry
	err = u.userRepo.DeleteSignUpVerificationEntry(ctx, email) // deletes all the verifcation entries made using the given email
	if err != nil {
		return utils.ErrDeleteVerificationEntry
	}

	return nil
}

func (u *userUseCase) ResendOTP(ctx context.Context, email string) error {
	// Get the verification entry
	entry, err := u.userRepo.FindSignUpVerificationEntryByEmail(ctx, email)
	if err != nil {
		if err == utils.ErrVerificationEntryNotFound {
			return utils.ErrNonExEmail
		}
		return err
	}

	// Check if the email is already verified
	if entry.IsVerified {
		return utils.ErrEmailAlreadyVerified
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

	err = u.userRepo.UpdateSignUpVerificationEntryAfterResendOTP(ctx, entry)
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

// GetUserProfile retrieves the profile of a user by their userID.
// It checks if the user exists and is not blocked. If the user is found and not blocked,
// it returns the user profile. If the user is not found or is blocked, it returns the
// appropriate error.
//
// Parameters:
//
//	ctx (context.Context) : The context for managing request-scoped values, deadlines, and cancellation.
//	userID (int64)        : The ID of the user whose profile is being retrieved.
//
// Returns:
//
//	(*domain.User, error) : A pointer to the user profile on success, or an error if the operation fails.
//
// Possible errors:
//   - utils.ErrUserNotFound    : If the user with the provided userID does not exist.
//   - utils.ErrUserBlocked     : If the user is found but is blocked and their profile cannot be accessed.
//   - utils.ErrInternalServer  : For unexpected errors during the operation.
func (u *userUseCase) GetUserProfile(ctx context.Context, userID int64) (*domain.User, error) {
	user, err := u.userRepo.GetByID(ctx, userID)
	if err != nil {
		if err == utils.ErrUserNotFound {
			return nil, utils.ErrUserNotFound
		}
		log.Printf("Error: %v", err)
		return nil, utils.ErrInternalServer
	}

	if user.IsBlocked {
		return nil, utils.ErrUserBlocked
	}

	return user, nil
}

// UpdateProfile updates the user's profile with the provided update data.
// It first retrieves the user by their ID, checks if the user is blocked,
// and then updates their name and phone number if provided.
// It also sets the current time as the 'updated_at' field and commits the changes to the repository.
//
// Parameters:
//   - ctx: Context for controlling execution and managing timeouts.
//   - userID: ID of the user whose profile is being updated.
//   - updateData: Pointer to a UserUpdatedData struct containing the fields to be updated.
//
// Returns:
//   - *domain.User: The updated User object.
//   - error: Returns ErrUserNotFound if the user is not found, ErrUserBlocked if the user is blocked, or another error if the update fails.
func (u *userUseCase) UpdateProfile(ctx context.Context, userID int64, updateData *domain.UserUpdatedData) (*domain.User, error) {
	user, err := u.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, utils.ErrUserNotFound
	}

	if user.IsBlocked {
		return nil, utils.ErrUserBlocked
	}

	if updateData.Name != "" {
		user.Name = updateData.Name
	}

	if updateData.PhoneNumber != "" {
		user.PhoneNumber = updateData.PhoneNumber
	}

	// update the 'updated_at' time
	user.UpdatedAt = time.Now().UTC()

	err = u.userRepo.Update(ctx, user)
	if err != nil {
		log.Printf("error : %v", err)
		return nil, err
	}
	return user, nil
}

func (u *userUseCase) ForgotPassword(ctx context.Context, email string) error {
	// check if the user exists
	user, err := u.userRepo.GetByEmail(ctx, email)
	if err != nil {
		// user not found error and internal server error
		return err
	}

	// check if user is blocked
	if user.IsBlocked {
		return utils.ErrUserBlocked
	}

	// Generate reset token (OTP)
	resetToken, err := otputil.GenerateOTP(6)
	if err != nil {
		return utils.ErrGenerateOTP
	}

	//create a verification entry
	expiresAt := time.Now().UTC().Add(15 * time.Minute)
	passwordResetEntry := &domain.PasswordResetEntry{
		Email:     user.Email,
		OTPCode:   resetToken,
		ExpiresAt: expiresAt,
	}

	err = u.userRepo.CreatePasswordResetEntry(ctx, passwordResetEntry)
	if err != nil {
		return utils.ErrCreateVericationEntry
	}

	//send reset token email
	err = u.emailSender.SendPasswordResetToken(user.Email, resetToken)
	if err != nil {
		return utils.ErrSendingResetToken
	}

	return nil
}

func (u *userUseCase) ResetPassword(ctx context.Context, email, otp, newPassword string) error {
	// check if the user exists
	user, err := u.userRepo.GetByEmail(ctx, email)
	if err != nil {
		if err == utils.ErrUserNotFound {
			log.Println("User not found during password reset attempt")
			return utils.ErrUserNotFound
		}
		log.Printf("Unexpected error during password reset : %v", err)
		return err
	}

	// check if the user is blocked
	if user.IsBlocked {
		return utils.ErrUserBlocked
	}

	// Get the verification entry
	entry, err := u.userRepo.FindPasswordResetEntryByEmail(ctx, email)
	if err != nil {
		if err == utils.ErrVerificationEntryNotFound {
			return utils.ErrOTPNotRequested
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

	// Check if the new password is the same as the old one
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(newPassword)); err == nil {
		return utils.ErrSamePassword
	}

	// Hash the new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return utils.ErrHashingPassword
	}

	// Update the password
	err = u.userRepo.UpdatePassword(ctx, user.ID, string(hashedPassword))
	if err != nil {
		return err
	}

	// Delete the verification entry
	err = u.userRepo.DeletePasswordResetVerificationEntry(ctx, email)
	if err != nil {
		return err
	}

	return nil
}

// AddUserAddress adds a new address for a user, ensuring that the user exists, is not blocked,
// and that the address is not a duplicate. It first checks if the user exists and is active (not blocked),
// verifies that the address is unique, and then inserts the new address into the database.
//
// Parameters:
//   - ctx: Context for controlling execution and managing timeouts.
//   - userAddress: Pointer to a UserAddress struct containing the user's address details to be added.
//
// Returns:
//   - error: Returns an error if the user does not exist, the user is blocked, or the address already exists.
//     Also returns an error if any operation fails during the process
func (u *userUseCase) AddUserAddress(ctx context.Context, userAddress *domain.UserAddress) error {
	// check if the user exists and whether the user is blocked
	user, err := u.userRepo.GetByID(ctx, userAddress.UserID)
	if err != nil {
		if err == utils.ErrUserNotFound {
			return utils.ErrUserNotFound
		}
		return err
	}

	if user.IsBlocked {
		return utils.ErrUserBlocked
	}

	// check for duplicate address
	exists, err := u.userRepo.UserAddressExists(ctx, userAddress)
	if err != nil {
		log.Printf("error : %v", err)
		return err
	}
	if exists {
		return utils.ErrUserAddressAlreadyExists
	}

	// add the address
	err = u.userRepo.AddUserAddress(ctx, userAddress)
	if err != nil {
		log.Printf("error : %v", err)
		return err
	}

	return nil
}

// UpdateUserAddress updates the user's address in the system with the provided updated data.
// It retrieves the existing address by addressID, checks if the address belongs to the user,
// validates and updates the address fields, and persists the changes in the repository.
//
// Parameters:
//
//	ctx (context.Context)      : The context for managing request-scoped values, deadlines, and cancellation.
//	userID (int64)             : The ID of the user attempting to update the address.
//	addressID (int64)          : The ID of the address to be updated.
//	updatedAddressData (*domain.UserAddressUpdate) : A pointer to a struct containing the new address data to update.
//
// Returns:
//
//	(*domain.UserAddress, error) : A pointer to the updated user address on success, or an error if the operation fails.
//
// Possible errors:
//   - utils.ErrAddressNotFound         : If the address with the provided addressID is not found.
//   - utils.ErrUnauthorized            : If the address does not belong to the user with the provided userID.
//   - utils.ErrInvalidUserCityEntry    : If the city in the updated data is invalid.
//   - utils.ErrInvalidUserStateEntry   : If the state in the updated data is invalid.
//   - utils.ErrInvalidPinCode          : If the provided pin code in the updated data is invalid. (more than than 6 digits)
//   - utils.ErrInvalidPhoneNumber      : If the phone number in the updated data is invalid. (more than 10 digits)
//   - Other repository-related errors  : If updating the address in the repository fails.
func (u *userUseCase) UpdateUserAddress(ctx context.Context, userID, addressID int64, updatedAddressData *domain.UserAddressUpdate) (*domain.UserAddress, error) {
	// retrieve the existing address
	userAddress, err := u.userRepo.GetUserAddressByID(ctx, addressID)
	if err != nil {
		return nil, utils.ErrAddressNotFound
	}

	// check if the address belongs to the user
	if userAddress.UserID != userID {
		return nil, utils.ErrUnauthorized
	}

	// update fields if provided
	if updatedAddressData.AddressLine1 != nil {
		err = validator.ValidateAddressLine(*updatedAddressData.AddressLine1)
		if err != nil {
			return nil, err // verify errors in address validation
		}
		userAddress.AddressLine1 = *updatedAddressData.AddressLine1
	}

	if updatedAddressData.AddressLine2 != nil {
		userAddress.AddressLine2 = *updatedAddressData.AddressLine2
	}

	if updatedAddressData.City != nil {
		err = validator.ValidateCity(*updatedAddressData.City)
		if err != nil {
			return nil, utils.ErrInvalidUserCityEntry
		}
		userAddress.City = *updatedAddressData.City
	}

	if updatedAddressData.State != nil {
		err = validator.ValidateState(*updatedAddressData.State)
		if err != nil {
			return nil, utils.ErrInvalidUserStateEntry
		}
		userAddress.State = *updatedAddressData.State
	}

	if updatedAddressData.PinCode != nil {
		err = validator.ValidatePinCode(*updatedAddressData.PinCode)
		if err != nil {
			return nil, utils.ErrInvalidPinCode
		}
		userAddress.PinCode = *updatedAddressData.PinCode
	}

	if updatedAddressData.Landmark != nil {
		userAddress.Landmark = *updatedAddressData.Landmark
	}
	if updatedAddressData.PhoneNumber != nil {
		if err := validator.ValidatePhoneNumber(*updatedAddressData.PhoneNumber); err != nil {
			return nil, utils.ErrInvalidPhoneNumber
		}
		userAddress.PhoneNumber = *updatedAddressData.PhoneNumber
	}

	// change update time
	userAddress.UpdatedAt = time.Now().UTC()

	// update the address in the repository
	err = u.userRepo.UpdateUserAddress(ctx, userAddress)
	if err != nil {
		return nil, err
	}

	return userAddress, nil
}

// GetUserAddresses retrieves all addresses associated with a given user ID.
// It first verifies that the user exists and is not blocked. If the user is not found or is blocked,
// it returns the appropriate error. If the user is valid and not blocked, it proceeds to fetch and return
// the user's addresses from the repository.
//
// Parameters:
//
//	ctx (context.Context) : The context for managing request-scoped values, deadlines, and cancellation.
//	userID (int64)        : The ID of the user whose addresses are being retrieved.
//
// Returns:
//
//	([]*domain.UserAddress, error) : A slice of pointers to user addresses on success, or an error if the operation fails.
//
// Possible errors:
//   - utils.ErrUserNotFound          : If the user with the provided userID does not exist.
//   - utils.ErrUserBlocked           : If the user is found but is blocked and cannot access the addresses.
//   - Other repository-related errors : If retrieving addresses from the repository fails.
func (u *userUseCase) GetUserAddresses(ctx context.Context, userID int64) ([]*domain.UserAddress, error) {
	// Check if the user exists and is not blocked
	user, err := u.userRepo.GetByID(ctx, userID)
	if err != nil {
		if err == utils.ErrUserNotFound {
			return nil, utils.ErrUserNotFound
		}
		return nil, err
	}

	if user.IsBlocked {
		return nil, utils.ErrUserBlocked
	}

	// Retrieve user addresses
	addresses, err := u.userRepo.GetUserAddresses(ctx, userID)
	if err != nil {
		log.Printf("error : %v", err)
		return nil, err
	}

	return addresses, nil
}

// DeleteUserAddress deletes a user address if it exists and belongs to the specified user.
// The function first checks if the address exists and belongs to the user. If the address is not found or
// does not belong to the user, it returns the appropriate error. It also ensures that the address being deleted
// is not the last address for the user. If it is the last address, it returns an error to prevent deletion.
// If all checks pass, it performs a soft delete of the address.
//
// Parameters:
//
//	ctx (context.Context) : The context for managing request-scoped values, deadlines, and cancellation.
//	userID (int64)        : The ID of the user attempting to delete the address.
//	addressID (int64)     : The ID of the address to be deleted.
//
// Returns:
//
//	error : An error if the operation fails. Possible errors include:
//	  - utils.ErrAddressNotFound : If the address with the given addressID does not exist.
//	  - utils.ErrUnauthorized    : If the address does not belong to the user with the given userID.
//	  - utils.ErrLastAddress     : If the address being deleted is the last address for the user.
//	  - Other repository-related errors if any of the repository operations fail.
func (u *userUseCase) DeleteUserAddress(ctx context.Context, userID, addressID int64) error {
	// Check if the address exists and belongs to the user
	address, err := u.userRepo.GetUserAddressByID(ctx, addressID)
	if err != nil {
		return err
	}

	if address == nil {
		return utils.ErrAddressNotFound
	}

	if address.UserID != userID {
		return utils.ErrUnauthorized
	}

	// Check if this is the last address
	addressCount, err := u.userRepo.GetUserAddressCount(ctx, userID)
	if err != nil {
		return err
	}

	if addressCount == 1 {
		return utils.ErrLastAddress
	}

	// Delete the address
	err = u.userRepo.DeleteUserAddress(ctx, addressID)
	if err != nil {
		return err
	}

	return nil
}
