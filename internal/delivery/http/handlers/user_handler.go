package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/mohamedfawas/rmshop-clean-architecture/internal/domain"
	"github.com/mohamedfawas/rmshop-clean-architecture/internal/usecase"
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

	//check for empty fields
	if input.Name == "" || input.Email == "" || input.Password == "" || input.DateOfBirth == "" || input.PhoneNumber == "" {
		http.Error(w, "All fields are required", http.StatusBadRequest)
		return
	}

	//trim trailing and leading spaces
	input.Name = strings.TrimSpace(input.Name)
	input.Email = strings.TrimSpace(input.Email)
	input.PhoneNumber = strings.TrimSpace(input.PhoneNumber)
	input.DateOfBirth = strings.TrimSpace(input.DateOfBirth)

	//validate email format
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(input.Email) {
		http.Error(w, "Invalid Email Format", http.StatusBadRequest)
		return
	}

	//validate password length
	if len(input.Password) < 8 {
		http.Error(w, "Password must be at least 8 characters long", http.StatusBadRequest)
		return
	}

	// Parse the date of birth
	dob, err := time.Parse("2006-01-02", input.DateOfBirth)
	if err != nil {
		log.Printf("Error parsing date of birth: %v", err)
		http.Error(w, "Invalid date format for date_of_birth", http.StatusBadRequest)
		return
	}

	//validate phone number
	phoneRegex := regexp.MustCompile(`^\d{10}$`)
	if !phoneRegex.MatchString(input.PhoneNumber) {
		http.Error(w, "Invalid phone number (it should have 10 digits)", http.StatusBadRequest)
		return
	}

	//max name length criteria
	if len(input.Name) > 100 {
		http.Error(w, "Name is too long (maximum 100 characters)", http.StatusBadRequest)
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
			http.Error(w, "Email already exists", http.StatusConflict)
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
		log.Printf("Error decoding request body: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	err = h.userUseCase.VerifyOTP(r.Context(), input.Email, input.OTP)
	if err != nil {
		log.Printf("Error verifying OTP: %v", err)
		switch err {
		case usecase.ErrInvalidOTP:
			http.Error(w, "Invalid OTP", http.StatusBadRequest)
		case usecase.ErrExpiredOTP:
			http.Error(w, "OTP has expired", http.StatusBadRequest)
		case usecase.ErrOTPNotFound:
			http.Error(w, "OTP not found", http.StatusNotFound)
		default:
			log.Printf("Unexpected error in VerifyOTP: %v", err)
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
