package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/mohamedfawas/rmshop-clean-architecture/internal/domain"
	"github.com/mohamedfawas/rmshop-clean-architecture/internal/usecase"
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
	// Define a struct to parse the login input from JSON
	var input LoginInput
	// Decode the JSON request body into the input struct
	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		// If there's an error in parsing, return a 400 Bad Request error
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Trim and lowercase email
	input.Email = strings.TrimSpace(strings.ToLower(input.Email))
	input.Password = strings.TrimSpace(input.Password)

	// validate credentials
	if err := validator.ValidateUserLoginCredentials(input.Email, input.Password); err != nil {
		http.Error(w, "Please give an input for email and password", http.StatusBadRequest)
		return
	}

	// Validate email format
	if err = validator.ValidateUserEmail(input.Email); err != nil {
		http.Error(w, "Invalid email format", http.StatusBadRequest)
		return
	}

	token, err := h.userUseCase.Login(r.Context(), input.Email, input.Password)
	if err != nil {
		switch err {
		case usecase.ErrInvalidCredentials:
			http.Error(w, "Invalid email or password", http.StatusUnauthorized)
		case usecase.ErrUserBlocked:
			http.Error(w, "User is blocked", http.StatusForbidden)
		default:
			log.Printf("Login error: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	// If login is successful, set the Content-Type header to JSON
	w.Header().Set("Content-Type", "application/json")
	// Set the status code to 200 OK
	w.WriteHeader(http.StatusOK)
	// Encode and send the token in the response body
	json.NewEncoder(w).Encode(LoginResponse{Token: token})
}

func (h *UserHandler) Logout(w http.ResponseWriter, r *http.Request) {
	// Extract the token from the Authorization header
	token := r.Header.Get("Authorization")
	if token == "" {
		// If the token is missing, return a 401 Unauthorized error
		http.Error(w, "Missing authorization token", http.StatusUnauthorized)
		return
	}

	// Remove "Bearer " prefix if present
	if len(token) > 7 && token[:7] == "Bearer " {
		token = token[7:]
	}

	// Call the Logout method of the userUseCase, passing the token
	err := h.userUseCase.Logout(r.Context(), token)
	if err != nil {
		if err == usecase.ErrInvalidToken {
			// If the token is invalid, return a 401 Unauthorized error
			http.Error(w, "Invalid token", http.StatusUnauthorized)
		} else {
			// For any other error, return a 500 Internal Server Error
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	// If logout is successful, set the status code to 200 OK
	w.WriteHeader(http.StatusOK)
	// Send a success message in the response body
	json.NewEncoder(w).Encode(map[string]string{"message": "Logged out successfully"})
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
		http.Error(w, "Unsupported Media Type", http.StatusUnsupportedMediaType)
		return
	}

	var input InitiateSignUpInput
	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		log.Printf("Error decoding request body: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
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
			http.Error(w, "Give a valid name", http.StatusBadRequest)
		case utils.ErrUserNameTooShort:
			http.Error(w, "Username should have atleast two characters", http.StatusBadRequest)
		case utils.ErrUserNameTooLong:
			http.Error(w, "Username should have less than 200 characters", http.StatusBadRequest)
		case utils.ErrUserNameWithNumericVals:
			http.Error(w, "Username should not contain numeric characters", http.StatusBadRequest)
		default:
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	//validate email
	err = validator.ValidateUserEmail(input.Email)
	if err != nil {
		switch err {
		case utils.ErrMissingEmail:
			http.Error(w, "The email field is required and must be a valid email address", http.StatusBadRequest)
		case utils.ErrInvalidEmail:
			http.Error(w, "Give a valid email", http.StatusBadRequest)
		default:
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	//validate password
	err = validator.ValidatePassword(input.Password)
	if err != nil {
		switch err {
		case utils.ErrPasswordInvalid:
			http.Error(w, "Give a valid password", http.StatusBadRequest)
		case utils.ErrPasswordTooShort:
			http.Error(w, "Password should have atleast 8 characters", http.StatusBadRequest)
		case utils.ErrPasswordTooLong:
			http.Error(w, "password should not have greater than 64 characters", http.StatusBadRequest)
		case utils.ErrPasswordSecurity:
			http.Error(w, "Password should have atleast one upper case letter, one lower case letter, one number and one special character", http.StatusBadRequest)
		default:
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	// validate dob
	err = validator.ValidateUserDOB(input.DateOfBirth)
	if err != nil {
		http.Error(w, "Please give dob in YYYY-MM-DD format", http.StatusBadRequest)
		return
	}
	dob, _ := time.Parse("2006-01-02", input.DateOfBirth) //parse dob

	//validate phone number
	err = validator.ValidatePhoneNumber(input.PhoneNumber)
	if err != nil {
		http.Error(w, "Invalid phone number (it should have 10 digits)", http.StatusBadRequest)
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
		log.Printf("Error in InitiateSignUp: %v", err)
		switch err {
		case usecase.ErrDuplicateEmail:
			http.Error(w, "Email already registered", http.StatusConflict)
		case usecase.ErrInvalidInput:
			http.Error(w, err.Error(), http.StatusBadRequest)
		default:
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "OTP sent to your email"})
}

type VerifyOTPInput struct {
	Email string `json:"email"`
	OTP   string `json:"otp"`
}

func (h *UserHandler) VerifyOTP(w http.ResponseWriter, r *http.Request) {
	var input VerifyOTPInput
	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
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
			http.Error(w, "The email field is required and must be a valid email address", http.StatusBadRequest)
		case utils.ErrInvalidEmail:
			http.Error(w, "Give a valid email", http.StatusBadRequest)
		default:
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	// Validate OTP
	if err := validator.ValidateOTP(input.OTP); err != nil {
		switch err {
		case utils.ErrMissingOTP:
			http.Error(w, "Give a valid OTP", http.StatusBadRequest)
		case utils.ErrOtpLength:
			http.Error(w, "OTP must have 6 digits", http.StatusBadRequest)
		case utils.ErrOtpNums:
			http.Error(w, "OTP must contain digits only", http.StatusBadRequest)
		default:
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	err = h.userUseCase.VerifyOTP(r.Context(), input.Email, input.OTP)
	if err != nil {
		switch err {
		case usecase.ErrInvalidOTP:
			http.Error(w, "Invalid OTP", http.StatusBadRequest)
		case usecase.ErrExpiredOTP:
			http.Error(w, "OTP has expired", http.StatusBadRequest)
		case usecase.ErrNonExEmail:
			http.Error(w, "User not found", http.StatusNotFound)
		case usecase.ErrEmailAlreadyVerified:
			http.Error(w, "Email already verified", http.StatusBadRequest)
		default:
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Email verified successfully"})
}

type ResendOTPInput struct {
	Email string `json:"email"`
}

func (h *UserHandler) ResendOTP(w http.ResponseWriter, r *http.Request) {
	var input ResendOTPInput
	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Trim whitespace and make lower case
	input.Email = strings.ToLower(strings.TrimSpace(input.Email))

	// Validate email
	if err := validator.ValidateUserEmail(input.Email); err != nil {
		http.Error(w, "Invalid email format", http.StatusBadRequest)
		return
	}

	err = h.userUseCase.ResendOTP(r.Context(), input.Email)
	if err != nil {
		switch err {
		case usecase.ErrNonExEmail:
			http.Error(w, "Given Email not initiated sign up", http.StatusNotFound)
		case usecase.ErrEmailAlreadyVerified:
			http.Error(w, "Email already verified", http.StatusBadRequest)
		case usecase.ErrTooManyResendAttempts:
			http.Error(w, "Too many resend attempts. Please try again later.", http.StatusTooManyRequests)
		case usecase.ErrSignupExpired:
			http.Error(w, "Signup process has expired. Please start over.", http.StatusBadRequest)
		default:
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "New OTP sent to your email"})
}
