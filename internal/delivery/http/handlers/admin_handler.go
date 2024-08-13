package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/mohamedfawas/rmshop-clean-architecture/internal/usecase"
)

type AdminHandler struct {
	adminUseCase usecase.AdminUseCase
}

func NewAdminHandler(adminUseCase usecase.AdminUseCase) *AdminHandler {
	return &AdminHandler{adminUseCase: adminUseCase}
}

type AdminLoginInput struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type AdminLoginResponse struct {
	Token string `json:"token"`
}

// Login handles the HTTP request for admin login
func (h *AdminHandler) Login(w http.ResponseWriter, r *http.Request) {
	// Define a struct to parse the login input from JSON
	var input AdminLoginInput
	// Decode the JSON request body into the input struct
	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		// If there's an error in parsing, return a 400 Bad Request error
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Trim whitespace from username and password
	input.Username = strings.TrimSpace(input.Username)
	input.Password = strings.TrimSpace(input.Password)

	// Check for empty username or password
	if input.Username == "" || input.Password == "" {
		http.Error(w, "Username and password are required", http.StatusBadRequest)
		return
	}

	// Check username length (assuming max length is 50)
	if len(input.Username) > 50 {
		http.Error(w, "Username is too long", http.StatusBadRequest)
		return
	}

	// Check password length (assuming max length is 100)
	if len(input.Password) > 100 {
		http.Error(w, "Password is too long", http.StatusBadRequest)
		return
	}

	// Call the Login method of the adminUseCase, passing the username and password
	token, err := h.adminUseCase.Login(r.Context(), input.Username, input.Password)
	if err != nil {
		// Handle different types of errors
		switch err {
		case usecase.ErrInvalidAdminCredentials:
			// If credentials are invalid, return a 401 Unauthorized error
			http.Error(w, "Invalid username or password", http.StatusUnauthorized)
		default:
			// For any other error, return a 500 Internal Server Error
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	// If login is successful, set the Content-Type header to JSON
	w.Header().Set("Content-Type", "application/json")
	// Set the status code to 200 OK
	w.WriteHeader(http.StatusOK)
	// Encode and send the token in the response body
	json.NewEncoder(w).Encode(AdminLoginResponse{Token: token})
}

func (h *AdminHandler) Logout(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("Authorization")
	if token == "" {
		http.Error(w, "Missing authorization token", http.StatusUnauthorized)
		return
	}

	// Remove "Bearer " prefix if present
	if len(token) > 7 && token[:7] == "Bearer " {
		token = token[7:]
	}

	err := h.adminUseCase.Logout(r.Context(), token)
	if err != nil {
		switch err {
		case usecase.ErrInvalidAdminToken:
			http.Error(w, "Invalid token", http.StatusUnauthorized)
		case usecase.ErrTokenAlreadyBlacklisted:
			http.Error(w, "Token already invalidated", http.StatusBadRequest)
		default:
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Admin logged out successfully"})
}
