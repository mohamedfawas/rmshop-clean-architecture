package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/mohamedfawas/rmshop-clean-architecture/internal/delivery/http/middleware"
	"github.com/mohamedfawas/rmshop-clean-architecture/internal/domain"
	"github.com/mohamedfawas/rmshop-clean-architecture/internal/usecase"
	"github.com/mohamedfawas/rmshop-clean-architecture/pkg/api"
	"github.com/mohamedfawas/rmshop-clean-architecture/pkg/utils"
	"github.com/mohamedfawas/rmshop-clean-architecture/pkg/validator"
)

type UserHandler struct {
	userUseCase usecase.UserUseCase
}

func NewUserHandler(userUseCase usecase.UserUseCase) *UserHandler {
	return &UserHandler{userUseCase: userUseCase}
}

type LoginInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token string `json:"token"`
}

// Login handles the HTTP request for user login
func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	var input LoginInput                          // Define a struct to parse the login input from JSON
	err := json.NewDecoder(r.Body).Decode(&input) // Decode the JSON request body into the input struct
	if err != nil {
		// If there's an error in parsing, return a 400 Bad Request error
		api.SendResponse(w, http.StatusBadRequest, "Failed to parse request", nil, "Invalid request body")
		return
	}

	// Trim and convert to lowercase email
	input.Email = strings.TrimSpace(strings.ToLower(input.Email))
	input.Password = strings.TrimSpace(input.Password)

	// validate credentials
	if err := validator.ValidateUserLoginCredentials(input.Email, input.Password); err != nil {
		api.SendResponse(w, http.StatusBadRequest, "Validation failed", nil, "Please provide valid email and password")
		return
	}

	// Validate email format
	if err = validator.ValidateUserEmail(input.Email); err != nil {
		api.SendResponse(w, http.StatusBadRequest, "Validation failed", nil, "Invalid email format")
		return
	}

	token, err := h.userUseCase.Login(r.Context(), input.Email, input.Password)
	if err != nil {
		switch err {
		case utils.ErrInvalidCredentials:
			api.SendResponse(w, http.StatusUnauthorized, "Login failed", nil, "Invalid email or password")
		case utils.ErrUserBlocked:
			api.SendResponse(w, http.StatusUnauthorized, "Login failed", nil, "User is blocked")
		case utils.ErrUpdateLastLogin:
			api.SendResponse(w, http.StatusInternalServerError, "Login failed", nil, "Error while updating last login time")
		case utils.ErrGenerateJWTTokenWithRole:
			api.SendResponse(w, http.StatusInternalServerError, "Login failed", nil, "Error while generating jwt token with role")
		default:
			api.SendResponse(w, http.StatusInternalServerError, "Login failed", nil, "An unexpected error occured")
		}
		return
	}

	response := LoginResponse{Token: token}
	api.SendResponse(w, http.StatusOK, "Login successful", response, "")
}

func (h *UserHandler) Logout(w http.ResponseWriter, r *http.Request) {
	// Extract the token from the Authorization header
	authHeader := r.Header.Get("Authorization")
	token, err := validator.ValidateAuthHeaderAndReturnToken(authHeader)
	if err != nil {
		switch err {
		case utils.ErrMissingAuthToken:
			api.SendResponse(w, http.StatusUnauthorized, "Authentication failed", nil, "Missing authorization token")
		case utils.ErrAuthHeaderFormat:
			api.SendResponse(w, http.StatusUnauthorized, "Authentication failed", nil, "Invalid authorization header format")
		case utils.ErrEmptyToken:
			api.SendResponse(w, http.StatusUnauthorized, "Authentication failed", nil, "Empty token")
		default:
			api.SendResponse(w, http.StatusInternalServerError, "Authentication failed", nil, "An unexpected error occured")
		}
		return
	}

	// Call the Logout method of the userUseCase, passing the token
	err = h.userUseCase.Logout(r.Context(), token)
	if err != nil {
		switch err {
		case utils.ErrInvalidToken:
			api.SendResponse(w, http.StatusUnauthorized, "Logout failed", nil, "Invalid token")
		case utils.ErrTokenAlreadyBlacklisted:
			api.SendResponse(w, http.StatusBadRequest, "Logout failed", nil, "Token already invalidated")
		case utils.ErrFailedToCheckBlacklisted:
			api.SendResponse(w, http.StatusInternalServerError, "Logout failed", nil, "Failed to check if token is blacklisted")
		default:
			api.SendResponse(w, http.StatusInternalServerError, "Logout failed", nil, "An unexpected error occured")
		}
		return
	}

	api.SendResponse(w, http.StatusOK, "Logged out successfully", nil, "")
}

type InitiateSignUpInput struct {
	Name        string `json:"name"`
	Email       string `json:"email"`
	Password    string `json:"password"`
	DateOfBirth string `json:"date_of_birth"`
	PhoneNumber string `json:"phone_number"`
}

func (h *UserHandler) InitiateSignUp(w http.ResponseWriter, r *http.Request) {
	// Check Content-Type : input should be json type
	if r.Header.Get("Content-Type") != "application/json" {
		api.SendResponse(w, http.StatusUnsupportedMediaType, "Unsupported Media Type", nil, "Content-Type must be application/json")
		return
	}

	var input InitiateSignUpInput
	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		api.SendResponse(w, http.StatusBadRequest, "Invalid request body", nil, "Failed to decode request body")
		return
	}

	//trim trailing and leading spaces and convert to lower case to make it case insensitive
	input.Name = strings.ToLower(strings.TrimSpace(input.Name))
	input.Email = strings.ToLower(strings.TrimSpace(input.Email))
	input.PhoneNumber = strings.TrimSpace(input.PhoneNumber)
	input.DateOfBirth = strings.TrimSpace(input.DateOfBirth)

	//validate user name
	err = validator.ValidateUserName(input.Name)
	if err != nil {
		switch err {
		case utils.ErrInvalidUserName:
			api.SendResponse(w, http.StatusBadRequest, "Invalid name", nil, "Please provide a valid name")
		case utils.ErrUserNameTooShort:
			api.SendResponse(w, http.StatusBadRequest, "Name too short", nil, "Name should have at least two characters")
		case utils.ErrUserNameTooLong:
			api.SendResponse(w, http.StatusBadRequest, "Name too long", nil, "Name should have less than 200 characters")
		case utils.ErrUserNameWithNumericVals:
			api.SendResponse(w, http.StatusBadRequest, "Invalid name", nil, "Name should not contain numeric characters")
		default:
			api.SendResponse(w, http.StatusInternalServerError, "Internal server error", nil, "An unexpected error occured")
		}
		return
	}

	//validate email
	err = validator.ValidateUserEmail(input.Email)
	if err != nil {
		switch err {
		case utils.ErrMissingEmail:
			api.SendResponse(w, http.StatusBadRequest, "Missing email", nil, "The email field is required and must be a valid email address")
		case utils.ErrInvalidEmail:
			api.SendResponse(w, http.StatusBadRequest, "Invalid email", nil, "Please provide a valid email address")
		default:
			api.SendResponse(w, http.StatusInternalServerError, "Internal server error", nil, "An unexpected error occured")
		}
		return
	}

	//validate password
	err = validator.ValidatePassword(input.Password)
	if err != nil {
		switch err {
		case utils.ErrPasswordInvalid:
			api.SendResponse(w, http.StatusBadRequest, "Invalid password", nil, "Please provide a valid password")
		case utils.ErrPasswordTooShort:
			api.SendResponse(w, http.StatusBadRequest, "Password too short", nil, "Password should have at least 8 characters")
		case utils.ErrPasswordTooLong:
			api.SendResponse(w, http.StatusBadRequest, "Password too long", nil, "Password should not have more than 64 characters")
		case utils.ErrPasswordSecurity:
			api.SendResponse(w, http.StatusBadRequest, "Weak password", nil, "Password should have at least one upper case letter, one lower case letter, one number and one special character")
		default:
			api.SendResponse(w, http.StatusInternalServerError, "Internal server error", nil, "An unexpected error occured")
		}
		return
	}

	// validate dob
	err = validator.ValidateUserDOB(input.DateOfBirth)
	if err != nil {
		api.SendResponse(w, http.StatusBadRequest, "Invalid date of birth", nil, "Please provide date of birth in YYYY-MM-DD format")
		return
	}
	dob, _ := time.Parse("2006-01-02", input.DateOfBirth) //parse dob

	//validate phone number
	err = validator.ValidatePhoneNumber(input.PhoneNumber)
	if err != nil {
		api.SendResponse(w, http.StatusBadRequest, "Invalid phone number", nil, "Phone number should have 10 digits")
		return
	}

	user := &domain.User{
		Name:        input.Name,
		Email:       input.Email,
		Password:    input.Password,
		DOB:         dob,
		PhoneNumber: input.PhoneNumber,
	}

	err = h.userUseCase.InitiateSignUp(r.Context(), user)
	if err != nil {
		switch err {
		case utils.ErrDuplicateEmail:
			api.SendResponse(w, http.StatusConflict, "Failed to initiate sign up", nil, "Email already registered")
		case utils.ErrHashingPassword:
			api.SendResponse(w, http.StatusInternalServerError, "Failed to initiate sign up", nil, "Error while hashing the password")
		case utils.ErrGenerateOTP:
			api.SendResponse(w, http.StatusInternalServerError, "Failed to initiate sign up", nil, "Error while generating OTP")
		case utils.ErrCreateVericationEntry:
			api.SendResponse(w, http.StatusInternalServerError, "Failed to initiate sign up", nil, "Error while creating the verification entry")
		case utils.ErrSendingOTP:
			api.SendResponse(w, http.StatusInternalServerError, "Failed to initiate sign up", nil, "Error while sending OTP")
		default:
			api.SendResponse(w, http.StatusInternalServerError, "Failed to initiate sign up", nil, "An unexpected error occured")
		}
		return
	}

	api.SendResponse(w, http.StatusOK, "Sign up initiated successfully", map[string]string{"message": "OTP sent to your email"}, "")
}

type VerifyOTPInput struct {
	Email string `json:"email"`
	OTP   string `json:"otp"`
}

func (h *UserHandler) VerifyOTP(w http.ResponseWriter, r *http.Request) {
	var input VerifyOTPInput
	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		api.SendResponse(w, http.StatusBadRequest, "Failed to verify OTP", nil, "Invalid request body")
		return
	}

	// Trim whitespace and make lower case
	input.Email = strings.ToLower(strings.TrimSpace(input.Email))
	input.OTP = strings.TrimSpace(input.OTP)

	// Validate email
	err = validator.ValidateUserEmail(input.Email)
	if err != nil {
		switch err {
		case utils.ErrMissingEmail:
			api.SendResponse(w, http.StatusBadRequest, "Failed to verify OTP", nil, "The email field is required and must be a valid email address")
		case utils.ErrInvalidEmail:
			api.SendResponse(w, http.StatusBadRequest, "Failed to verify OTP", nil, "Give a valid email")
		default:
			api.SendResponse(w, http.StatusInternalServerError, "Failed to verify OTP", nil, err.Error())
		}
		return
	}

	// Validate OTP
	if err := validator.ValidateOTP(input.OTP); err != nil {
		switch err {
		case utils.ErrMissingOTP:
			api.SendResponse(w, http.StatusBadRequest, "Failed to verify OTP", nil, "Give a valid OTP")
		case utils.ErrOtpLength:
			api.SendResponse(w, http.StatusBadRequest, "Failed to verify OTP", nil, "OTP must have 6 digits")
		case utils.ErrOtpNums:
			api.SendResponse(w, http.StatusBadRequest, "Failed to verify OTP", nil, "OTP must contain digits only")
		default:
			api.SendResponse(w, http.StatusInternalServerError, "Failed to verify OTP", nil, err.Error())
		}
		return
	}

	err = h.userUseCase.VerifyOTP(r.Context(), input.Email, input.OTP)
	if err != nil {
		switch err {
		case utils.ErrInvalidOTP:
			api.SendResponse(w, http.StatusBadRequest, "Failed to verify OTP", nil, "Invalid OTP")
		case utils.ErrExpiredOTP:
			api.SendResponse(w, http.StatusBadRequest, "Failed to verify OTP", nil, "OTP has expired")
		case utils.ErrNonExEmail:
			api.SendResponse(w, http.StatusNotFound, "Failed to verify OTP", nil, "OTP not generated for the given email")
		case utils.ErrEmailAlreadyVerified:
			api.SendResponse(w, http.StatusBadRequest, "Failed to verify OTP", nil, "Email already verified")
		case utils.ErrCreateUser:
			api.SendResponse(w, http.StatusInternalServerError, "Failed to verify OTP", nil, "Error while creating the user")
		case utils.ErrUpdateVerificationEntry:
			api.SendResponse(w, http.StatusInternalServerError, "Failed to verify OTP", nil, "Error while updating the verification entry")
		case utils.ErrDeleteVerificationEntry:
			api.SendResponse(w, http.StatusInternalServerError, "Failed to verify OTP", nil, "Error while deleting the verification entry")
		default:
			api.SendResponse(w, http.StatusInternalServerError, "Failed to verify OTP", nil, err.Error())
		}
		return
	}

	api.SendResponse(w, http.StatusOK, "Email verified successfully", nil, "")
}

type ResendOTPInput struct {
	Email string `json:"email"`
}

func (h *UserHandler) ResendOTP(w http.ResponseWriter, r *http.Request) {
	var input ResendOTPInput
	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		api.SendResponse(w, http.StatusBadRequest, "Failed to resend OTP", nil, "Invalid request body")
		return
	}

	// Trim whitespace and convert to lower case
	input.Email = strings.ToLower(strings.TrimSpace(input.Email))

	// Validate email
	if err := validator.ValidateUserEmail(input.Email); err != nil {
		switch err {
		case utils.ErrMissingEmail:
			api.SendResponse(w, http.StatusBadRequest, "Failed to resend OTP", nil, "The email field is required and must be a valid email address")
		case utils.ErrInvalidEmail:
			api.SendResponse(w, http.StatusBadRequest, "Failed to resend OTP", nil, "Give a valid email")
		default:
			api.SendResponse(w, http.StatusInternalServerError, "Failed to resend OTP", nil, err.Error())
		}
		return
	}

	err = h.userUseCase.ResendOTP(r.Context(), input.Email)
	if err != nil {
		switch err {
		case utils.ErrVerificationEntryType:
			api.SendResponse(w, http.StatusNotFound, "Failed to resend OTP", nil, "Provided Email hasn't used for sign up process")
		case utils.ErrNonExEmail:
			api.SendResponse(w, http.StatusNotFound, "Failed to resend OTP", nil, "Provided Email hasn't initiated sign up")
		case utils.ErrEmailAlreadyVerified:
			api.SendResponse(w, http.StatusBadRequest, "Failed to resend OTP", nil, "Email already verified")
		case utils.ErrTooManyResendAttempts:
			api.SendResponse(w, http.StatusTooManyRequests, "Failed to resend OTP", nil, "Too many resend attempts. Please try again later.")
		case utils.ErrSignupExpired:
			api.SendResponse(w, http.StatusBadRequest, "Failed to resend OTP", nil, "Initiated signup process has expired. Please start over.")
		case utils.ErrRetrieveOTPResendInfo:
			api.SendResponse(w, http.StatusInternalServerError, "Failed to resend OTP", nil, "Error while retrieving data from otp_resend_info table")
		case utils.ErrGenerateOTP:
			api.SendResponse(w, http.StatusInternalServerError, "Failed to resend OTP", nil, "Error while generating new OTP")
		case utils.ErrUpdateVerficationAfterResend:
			api.SendResponse(w, http.StatusInternalServerError, "Failed to resend OTP", nil, "Error while updating verification entry table after resending OTP")
		case utils.ErrUpdateOTPResendTable:
			api.SendResponse(w, http.StatusInternalServerError, "Failed to resend OTP", nil, "Error while updating otp_resend_info table")
		case utils.ErrSMTPServerIssue:
			api.SendResponse(w, http.StatusInternalServerError, "Failed to resend OTP", nil, "Error while sending OTP due to SMTP server issue")
		default:
			api.SendResponse(w, http.StatusInternalServerError, "Failed to resend OTP", nil, err.Error())
		}
		return
	}

	api.SendResponse(w, http.StatusOK, "OTP resent successfully", map[string]string{"message": "New OTP sent to your email"}, "")
}

func (h *UserHandler) GetUserProfile(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from the context (set by the JWT middleware)
	userID, ok := r.Context().Value(middleware.UserIDKey).(int64)
	if !ok {
		api.SendResponse(w, http.StatusUnauthorized, "Failed to retrieve user profile", nil, "Invalid user id in token")
		return
	}

	// call the use case method to get the user profile
	user, err := h.userUseCase.GetUserProfile(r.Context(), userID)
	if err != nil {
		switch err {
		case utils.ErrUserNotFound:
			api.SendResponse(w, http.StatusNotFound, "Failed to retrieve user profile", nil, "The requested user profile does not exist")
		case utils.ErrUserBlocked:
			api.SendResponse(w, http.StatusForbidden, "Failed to retrieve user profile", nil, "User is blocked")
		default:
			api.SendResponse(w, http.StatusInternalServerError, "Failed to retrieve user profile", nil, "An unexpected error occurred")
		}
		return
	}

	type UserProfileResponse struct {
		ID              int64     `json:"id"`
		Name            string    `json:"name"`
		Email           string    `json:"email"`
		DOB             string    `json:"date_of_birth"`
		PhoneNumber     string    `json:"phone_number"`
		IsEmailVerified bool      `json:"is_email_verified"`
		CreatedAt       time.Time `json:"created_at"`
		LastLogin       time.Time `json:"last_login,omitempty"`
		IsBlocked       bool      `json:"is_blocked"`
	}

	response := UserProfileResponse{
		ID:              user.ID,
		Name:            user.Name,
		Email:           user.Email,
		DOB:             user.DOB.Format("2006-01-02"),
		PhoneNumber:     user.PhoneNumber,
		IsEmailVerified: user.IsEmailVerified,
		IsBlocked:       user.IsBlocked,
		CreatedAt:       user.CreatedAt,
		LastLogin:       user.LastLogin,
	}
	api.SendResponse(w, http.StatusOK, "User profile retrieved successfully", response, "")
}

func (h *UserHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	//extract user id from the jwt token
	userID, ok := r.Context().Value(middleware.UserIDKey).(int64)
	if !ok {
		api.SendResponse(w, http.StatusUnauthorized, "Failed to update profile", nil, "Invalid token")
		return
	}

	var updateData domain.UserUpdatedData
	err := json.NewDecoder(r.Body).Decode(&updateData)
	if err != nil {
		api.SendResponse(w, http.StatusBadRequest, "Failed to update user profile", nil, "Invalid request body")
		return
	}

	//no input data
	if updateData.Name == "" && updateData.PhoneNumber == "" {
		api.SendResponse(w, http.StatusBadRequest, "Failed to update user profile", nil, "No update data provided")
		return
	}

	//validate the updated name
	if updateData.Name != "" {
		updateData.Name = strings.ToLower(strings.TrimSpace(updateData.Name))
		err = validator.ValidateUserName(updateData.Name)
		if err != nil {
			switch err {
			case utils.ErrUserNameTooShort:
				api.SendResponse(w, http.StatusBadRequest, "Failed to update user profile", nil, "Given name is too short: name should have atleast 2 characters")
			case utils.ErrUserNameTooLong:
				api.SendResponse(w, http.StatusBadRequest, "Failed to update user profile", nil, "Given name is too long: name should have less than 200 characters")
			case utils.ErrUserNameWithNumericVals:
				api.SendResponse(w, http.StatusBadRequest, "Failed to update user profile", nil, "Numerical values are not allowed in name")
			default:
				api.SendResponse(w, http.StatusInternalServerError, "Failed to update user profile", nil, "An unexpected error occured")
			}
			return
		}
	}

	if updateData.PhoneNumber != "" {
		updateData.PhoneNumber = strings.TrimSpace(updateData.PhoneNumber)
		//validate updated phone number
		err = validator.ValidatePhoneNumber(updateData.PhoneNumber)
		if err != nil {
			api.SendResponse(w, http.StatusBadRequest, "Failed to update user profile", nil, "Please give a valid phone number with 10 digits")
			return
		}
	}

	updatedUser, err := h.userUseCase.UpdateProfile(r.Context(), userID, &updateData)
	if err != nil {
		switch err {
		case utils.ErrUserNotFound:
			api.SendResponse(w, http.StatusNotFound, "Failed to update user profile", nil, "User not found")
		case utils.ErrUserBlocked:
			api.SendResponse(w, http.StatusForbidden, "Failed to update user profile", nil, "User is blocked")
		default:
			api.SendResponse(w, http.StatusInternalServerError, "Failed to update user profile", nil, "An unexpected error occured")
		}
		return
	}

	type updatedUserResponse struct {
		ID              int64     `json:"id"`
		Name            string    `json:"name"`
		Email           string    `json:"email"`
		DOB             string    `json:"date_of_birth"`
		PhoneNumber     string    `json:"phone_number"`
		IsEmailVerified bool      `json:"is_email_verified"`
		CreatedAt       time.Time `json:"created_at"`
		UpdatedAt       time.Time `json:"updated_at"`
		LastLogin       time.Time `json:"last_login,omitempty"`
		IsBlocked       bool      `json:"is_blocked"`
	}

	response := updatedUserResponse{
		ID:              updatedUser.ID,
		Name:            updatedUser.Name,
		Email:           updatedUser.Email,
		DOB:             updatedUser.DOB.Format("2006-01-02"),
		PhoneNumber:     updatedUser.PhoneNumber,
		IsEmailVerified: updatedUser.IsEmailVerified,
		CreatedAt:       updatedUser.CreatedAt,
		UpdatedAt:       updatedUser.UpdatedAt,
		LastLogin:       updatedUser.LastLogin,
		IsBlocked:       updatedUser.IsBlocked,
	}

	api.SendResponse(w, http.StatusOK, "Successfully updated the user profile", response, "")
}

func (h *UserHandler) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Email string `json:"email"`
	}

	// parse the request
	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		api.SendResponse(w, http.StatusBadRequest, "Failed to parse request", nil, "Invalid request body")
		return
	}

	input.Email = strings.ToLower(strings.TrimSpace(input.Email))

	if input.Email == "" {
		api.SendResponse(w, http.StatusBadRequest, "Failed to initiate the setup", nil, "Please provide a valid email address")
		return
	}

	err = validator.ValidateUserEmail(input.Email)
	if err != nil {
		api.SendResponse(w, http.StatusBadRequest, "Failed to initiate the setup", nil, "Please provide a valid email address")
		return
	}

	err = h.userUseCase.ForgotPassword(r.Context(), input.Email)
	if err != nil {
		switch err {
		case utils.ErrUserNotFound:
			api.SendResponse(w, http.StatusNotFound, "Failed to initiate password reset", nil, "User not found")
		case utils.ErrGenerateOTP:
			api.SendResponse(w, http.StatusInternalServerError, "Failed to initiate password reset", nil, "Error generating reset token")
		case utils.ErrCreateVerificationEntry:
			api.SendResponse(w, http.StatusInternalServerError, "Failed to initiate password reset", nil, "Error creating verification entry")
		case utils.ErrSendingResetToken:
			api.SendResponse(w, http.StatusInternalServerError, "Failed to initiate password reset", nil, "Error sending reset token")
		case utils.ErrUserBlocked:
			api.SendResponse(w, http.StatusForbidden, "Failed to initiate password reset", nil, "user is blocked")
		default:
			api.SendResponse(w, http.StatusInternalServerError, "Failed to initiate password reset", nil, "An unexpected error occurred")
		}
		return
	}

	api.SendResponse(w, http.StatusOK, "Password reset initiated", nil, "")
}

func (h *UserHandler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Email       string `json:"email"`
		OTP         string `json:"otp"`
		NewPassword string `json:"new_password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		api.SendResponse(w, http.StatusBadRequest, "Failed to parse request", nil, "Invalid request body")
		return
	}

	// Trim and validate input
	input.Email = strings.TrimSpace(strings.ToLower(input.Email))
	input.OTP = strings.TrimSpace(input.OTP)
	input.NewPassword = strings.TrimSpace(input.NewPassword)

	if input.Email == "" || input.OTP == "" || input.NewPassword == "" {
		api.SendResponse(w, http.StatusBadRequest, "Missing required fields", nil, "Email, OTP, and new password are required")
		return
	}

	if err := validator.ValidateUserEmail(input.Email); err != nil {
		api.SendResponse(w, http.StatusBadRequest, "Invalid email", nil, "Please provide a valid email address")
		return
	}

	if err := validator.ValidateOTP(input.OTP); err != nil {
		api.SendResponse(w, http.StatusBadRequest, "Invalid OTP", nil, "Please provide a valid 6-digit OTP")
		return
	}

	if err := validator.ValidatePassword(input.NewPassword); err != nil {
		api.SendResponse(w, http.StatusBadRequest, "Invalid password", nil, "Password does not meet complexity requirements")
		return
	}

	err := h.userUseCase.ResetPassword(r.Context(), input.Email, input.OTP, input.NewPassword)
	if err != nil {
		switch err {
		case utils.ErrUserNotFound:
			api.SendResponse(w, http.StatusNotFound, "Failed to reset password", nil, "User not found")
		case utils.ErrInvalidOTP:
			api.SendResponse(w, http.StatusBadRequest, "Failed to reset password", nil, "Invalid OTP")
		case utils.ErrExpiredOTP:
			api.SendResponse(w, http.StatusBadRequest, "Failed to reset password", nil, "OTP has expired")
		case utils.ErrOTPNotRequested:
			api.SendResponse(w, http.StatusBadRequest, "Failed to reset password", nil, "No OTP was requested for this email")
		case utils.ErrTooManyResetAttempts:
			api.SendResponse(w, http.StatusTooManyRequests, "Failed to reset password", nil, "Too many reset attempts. Please try again later")
		case utils.ErrSamePassword:
			api.SendResponse(w, http.StatusBadRequest, "Failed to reset password", nil, "New password cannot be the same as the old password")
		default:
			api.SendResponse(w, http.StatusInternalServerError, "Failed to reset password", nil, "An unexpected error occurred")
		}
		return
	}

	api.SendResponse(w, http.StatusOK, "Password reset successfully", nil, "")
}

func (h *UserHandler) AddUserAddress(w http.ResponseWriter, r *http.Request) {
	// extract userId from the context
	userID, ok := r.Context().Value(middleware.UserIDKey).(int64)
	if !ok {
		api.SendResponse(w, http.StatusUnauthorized, "Failed to add address", nil, "Invalid user ID in token")
		return
	}

	var userAddress domain.UserAddress
	err := json.NewDecoder(r.Body).Decode(&userAddress)
	if err != nil {
		api.SendResponse(w, http.StatusBadRequest, "Failed to add address", nil, "Invalid request body")
		return
	}

	// assign the user id taken from token to user id for address entry
	userAddress.UserID = userID

	userAddress.AddressLine1 = strings.ToLower(strings.TrimSpace(userAddress.AddressLine1))
	userAddress.AddressLine2 = strings.ToLower(strings.TrimSpace(userAddress.AddressLine2))
	userAddress.State = strings.ToLower(strings.TrimSpace(userAddress.State))
	userAddress.City = strings.ToLower(strings.TrimSpace(userAddress.City))
	userAddress.PinCode = strings.ToLower(strings.TrimSpace(userAddress.PinCode))
	userAddress.Landmark = strings.ToLower(strings.TrimSpace(userAddress.Landmark))
	userAddress.PhoneNumber = strings.ToLower(strings.TrimSpace(userAddress.PhoneNumber))

	// validate not null entries
	if userAddress.AddressLine1 == "" || userAddress.State == "" || userAddress.City == "" || userAddress.PinCode == "" || userAddress.PhoneNumber == "" {
		api.SendResponse(w, http.StatusBadRequest, "Failed to add the address", nil, "Please provide a valid address with all the necessary fields filled")
		return
	}

	// validate address line 1 , not validating address line 2 (optional value)
	err = validator.ValidateAddressLine(userAddress.AddressLine1)
	if err != nil {
		if err == utils.ErrUserAddressTooLong {
			api.SendResponse(w, http.StatusBadRequest, "Validation failed", nil, "Please provide address line with less than 255 characters")
			return
		}
		if err == utils.ErrUserAddressTooShort {
			api.SendResponse(w, http.StatusBadRequest, "Validation failed", nil, "Please provide address line with more than 10 characters")
			return
		}
	}

	err = validator.ValidateState(userAddress.State)
	if err != nil {
		api.SendResponse(w, http.StatusBadRequest, "Validation failed", nil, "Please give a valid state (with less than 100 characters)")
		return
	}

	err = validator.ValidateCity(userAddress.City)
	if err != nil {
		api.SendResponse(w, http.StatusBadRequest, "Validation failed", nil, "Please give a valid city (with less than 150 characters)")
		return
	}

	err = validator.ValidatePinCode(userAddress.PinCode)
	if err != nil {
		api.SendResponse(w, http.StatusBadRequest, "Validation failed", nil, "Please provide a valid pincode")
		return
	}

	err = validator.ValidatePhoneNumber(userAddress.PhoneNumber)
	if err != nil {
		api.SendResponse(w, http.StatusBadRequest, "Validation failed", nil, "Please provide a valid phone number")
		return
	}

	// call the use case to add the address
	err = h.userUseCase.AddUserAddress(r.Context(), &userAddress)
	if err != nil {
		switch err {
		case utils.ErrUserNotFound:
			api.SendResponse(w, http.StatusNotFound, "Failed to add user address", nil, "The given user is not a registered user")
		case utils.ErrUserBlocked:
			api.SendResponse(w, http.StatusForbidden, "Failed to add user address", nil, "The given user is blocked by admin")
		case utils.ErrUserAddressAlreadyExists:
			api.SendResponse(w, http.StatusConflict, "Failed to add user address", nil, "The given address already exists")
		default:
			api.SendResponse(w, http.StatusInternalServerError, "Internal server error", nil, "An unexpected error occured")
		}
		return
	}

	api.SendResponse(w, http.StatusCreated, "Address added successfully", userAddress, "")
}

func (h *UserHandler) UpdateUserAddress(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from the context (set by the JWT middleware)
	userID, ok := r.Context().Value(middleware.UserIDKey).(int64)
	if !ok {
		api.SendResponse(w, http.StatusUnauthorized, "Failed to update address", nil, "Invalid user ID in token")
		return
	}

	// extract address id from url parameters
	vars := mux.Vars(r)
	addressID, err := strconv.ParseInt(vars["addressId"], 10, 64)
	if err != nil {
		api.SendResponse(w, http.StatusBadRequest, "Failed to update address", nil, "Invalid adrress ID")
		return
	}

	// parse the request body
	var updatedAddressData domain.UserAddressUpdate
	err = json.NewDecoder(r.Body).Decode(&updatedAddressData)
	if err != nil {
		api.SendResponse(w, http.StatusBadRequest, "Failed to update address", nil, "Invalid request body")
		return
	}

	// call the use case method to update the address
	updatedUserAddress, err := h.userUseCase.UpdateUserAddress(r.Context(), userID, addressID, &updatedAddressData)
	if err != nil {
		switch err {
		case utils.ErrUserAddressTooShort:
			api.SendResponse(w, http.StatusBadRequest, "Failed to update address", nil, "User address too short (should have atleast 10 characters)")
		case utils.ErrUserAddressTooLong:
			api.SendResponse(w, http.StatusBadRequest, "Failed to update address", nil, "User address too long (should be less than 255 characters)")
		case utils.ErrInvalidUserStateEntry:
			api.SendResponse(w, http.StatusBadRequest, "Failed to update address", nil, "Please provide a valid state name")
		case utils.ErrInvalidUserCityEntry:
			api.SendResponse(w, http.StatusBadRequest, "Failed to update address", nil, "Please provide a valid city name")
		case utils.ErrInvalidPinCode:
			api.SendResponse(w, http.StatusBadRequest, "Failed to update address", nil, "Please provide a valid pincode")
		case utils.ErrInvalidPhoneNumber:
			api.SendResponse(w, http.StatusBadRequest, "Failed to update address", nil, "Please provide a valid phone number")
		case utils.ErrAddressNotFound:
			api.SendResponse(w, http.StatusNotFound, "Failed to update address", nil, "Address not found")
		case utils.ErrUnauthorized:
			api.SendResponse(w, http.StatusForbidden, "Failed to update address", nil, "You are not authorized to update this address")
		default:
			api.SendResponse(w, http.StatusInternalServerError, "Failed to update address", nil, "An unexpected error occurred")
		}
		return
	}

	api.SendResponse(w, http.StatusOK, "Address updated successfully", updatedUserAddress, "")
}

func (h *UserHandler) GetUserAddresses(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from the context (set by the JWT middleware)
	userID, ok := r.Context().Value(middleware.UserIDKey).(int64)
	if !ok {
		api.SendResponse(w, http.StatusUnauthorized, "Failed to retrieve user addresses", nil, "Invalid user ID in token")
		return
	}

	// Call the use case method to get the user addresses
	addresses, err := h.userUseCase.GetUserAddresses(r.Context(), userID)
	if err != nil {
		switch err {
		case utils.ErrUserNotFound:
			api.SendResponse(w, http.StatusNotFound, "Failed to retrieve user addresses", nil, "User not found")
		case utils.ErrUserBlocked:
			api.SendResponse(w, http.StatusForbidden, "Failed to retrieve user addresses", nil, "User is blocked")
		default:
			api.SendResponse(w, http.StatusInternalServerError, "Failed to retrieve user addresses", nil, "An unexpected error occurred")
		}
		return
	}

	if len(addresses) == 0 {
		api.SendResponse(w, http.StatusOK, "No addresses found", []struct{}{}, "")
		return
	}

	api.SendResponse(w, http.StatusOK, "User addresses retrieved successfully", addresses, "")
}

func (h *UserHandler) DeleteUserAddress(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from the context (set by the JWT middleware)
	userID, ok := r.Context().Value(middleware.UserIDKey).(int64)
	if !ok {
		api.SendResponse(w, http.StatusUnauthorized, "Authentication failed", nil, "Invalid user ID in token")
		return
	}

	// Extract address ID from URL parameters
	vars := mux.Vars(r)
	addressID, err := strconv.ParseInt(vars["addressId"], 10, 64)
	if err != nil {
		api.SendResponse(w, http.StatusBadRequest, "Invalid request", nil, "Invalid address ID format")
		return
	}

	// Call the use case method to delete the address
	err = h.userUseCase.DeleteUserAddress(r.Context(), userID, addressID)
	if err != nil {
		switch err {
		case utils.ErrAddressNotFound:
			api.SendResponse(w, http.StatusNotFound, "Address not found", nil, "The requested address does not exist")
		case utils.ErrUnauthorized:
			api.SendResponse(w, http.StatusForbidden, "Unauthorized", nil, "You are not authorized to delete this address")
		case utils.ErrLastAddress:
			api.SendResponse(w, http.StatusBadRequest, "Cannot delete", nil, "Cannot delete the last remaining address")
		default:
			api.SendResponse(w, http.StatusInternalServerError, "Internal server error", nil, "An unexpected error occurred")
		}
		return
	}

	api.SendResponse(w, http.StatusOK, "Address deleted successfully", nil, "")
}
