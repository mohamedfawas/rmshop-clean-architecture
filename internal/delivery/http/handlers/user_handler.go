package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/mohamedfawas/rmshop-clean-architecture/internal/domain"
	"github.com/mohamedfawas/rmshop-clean-architecture/internal/usecase"
	"github.com/mohamedfawas/rmshop-clean-architecture/pkg/validator"
)

type UserHandler struct {
	userUseCase usecase.UserUseCase
}

func NewUserHandler(userUseCase usecase.UserUseCase) *UserHandler {
	return &UserHandler{userUseCase: userUseCase}
}

type UserRegistrationInput struct {
	Name        string `json:"name"`
	Email       string `json:"email"`
	Password    string `json:"password"`
	DateOfBirth string `json:"date_of_birth"`
	PhoneNumber string `json:"phone_number"`
}

func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	var input UserRegistrationInput
	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Parse the date of birth
	dob, err := time.Parse("2006-01-02", input.DateOfBirth)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid date format for date_of_birth: %v", err), http.StatusBadRequest)
		return
	}

	// Validate user input
	err = validator.ValidateUserInput(input.Name, input.Email, input.Password, input.PhoneNumber, input.DateOfBirth)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	user := &domain.User{
		Name:        input.Name,
		Email:       input.Email,
		Password:    input.Password,
		DOB:         dob,
		PhoneNumber: input.PhoneNumber,
	}

	err = h.userUseCase.Register(r.Context(), user) //If the data is read successfully, it tries to register the user
	if err != nil {
		// http.Error(w, err.Error(), http.StatusInternalServerError) //If there's an error during registration, it sends back an error message
		// return
		switch err {
		case usecase.ErrDuplicateEmail:
			http.Error(w, "Email already exists", http.StatusConflict)
		default:
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "User registered successfully"})
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

	// Call the Login method of the userUseCase, passing the email and password
	token, err := h.userUseCase.Login(r.Context(), input.Email, input.Password)
	if err != nil {
		// Handle different types of errors
		switch err {
		case usecase.ErrInvalidCredentials:
			// If credentials are invalid, return a 401 Unauthorized error
			http.Error(w, "Invalid email or password", http.StatusUnauthorized)
		default:
			// Log the actual error for debugging purposes
			log.Printf("Login error: %v", err)
			// For any other error, return a 500 Internal Server Error
			http.Error(w, fmt.Sprintf("Internal server error: %v", err), http.StatusInternalServerError)
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
	var input InitiateSignUpInput
	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Parse the date of birth
	dob, err := time.Parse("2006-01-02", input.DateOfBirth)
	if err != nil {
		http.Error(w, "Invalid date format for date_of_birth", http.StatusBadRequest)
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
		case usecase.ErrDuplicateEmail:
			http.Error(w, "Email already exists", http.StatusConflict)
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

	err = h.userUseCase.VerifyOTP(r.Context(), input.Email, input.OTP)
	if err != nil {
		switch err {
		case usecase.ErrInvalidOTP:
			http.Error(w, "Invalid OTP", http.StatusBadRequest)
		case usecase.ErrExpiredOTP:
			http.Error(w, "OTP has expired", http.StatusBadRequest)
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

	err = h.userUseCase.ResendOTP(r.Context(), input.Email)
	if err != nil {
		switch err {
		case usecase.ErrUserNotFound:
			http.Error(w, "User not found", http.StatusNotFound)
		case usecase.ErrEmailAlreadyVerified:
			http.Error(w, "Email already verified", http.StatusBadRequest)
		default:
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "New OTP sent to your email"})
}
