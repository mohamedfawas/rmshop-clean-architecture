package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

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

	//trim space and convert to lower case
	updateData.Name = strings.ToLower(strings.TrimSpace(updateData.Name))
	updateData.PhoneNumber = strings.TrimSpace(updateData.PhoneNumber)

	//no input data
	if updateData.Name == "" && updateData.PhoneNumber == "" {
		api.SendResponse(w, http.StatusBadRequest, "Failed to update user profile", nil, "No update data provided")
		return
	}

	//validate the updated name
	if updateData.Name != "" {
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
