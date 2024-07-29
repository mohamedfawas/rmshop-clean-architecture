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

func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	var input LoginInput
	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	token, err := h.userUseCase.Login(r.Context(), input.Email, input.Password)
	if err != nil {
		switch err {
		case usecase.ErrInvalidCredentials:
			http.Error(w, "Invalid email or password", http.StatusUnauthorized)
		default:
			log.Printf("Login error: %v", err) // Log the actual error
			http.Error(w, fmt.Sprintf("Internal server error: %v", err), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(LoginResponse{Token: token})
}

func (h *UserHandler) Logout(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("Authorization")
	if token == "" {
		http.Error(w, "Missing authorization token", http.StatusUnauthorized)
		return
	}

	// Remove "Bearer " prefix if present
	if len(token) > 7 && token[:7] == "Bearer " {
		token = token[7:]
	}

	err := h.userUseCase.Logout(r.Context(), token)
	if err != nil {
		if err == usecase.ErrInvalidToken {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
		} else {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Logged out successfully"})
}
