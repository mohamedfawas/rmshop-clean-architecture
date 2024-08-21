package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/mohamedfawas/rmshop-clean-architecture/internal/usecase"
	"github.com/mohamedfawas/rmshop-clean-architecture/pkg/api"
	"github.com/mohamedfawas/rmshop-clean-architecture/pkg/utils"
	"github.com/mohamedfawas/rmshop-clean-architecture/pkg/validator"
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

func (h *AdminHandler) Login(w http.ResponseWriter, r *http.Request) {
	var input AdminLoginInput
	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		api.SendResponse(w, http.StatusBadRequest, "Login failed", nil, "Invalid request body")
		return
	}

	input.Username = strings.ToLower(strings.TrimSpace(input.Username))
	input.Password = strings.TrimSpace(input.Password)

	err = validator.ValidateAdminCredentials(input.Username, input.Password)
	if err != nil {
		switch err {
		case utils.ErrInvalidAdminCredentials:
			api.SendResponse(w, http.StatusBadRequest, "Login failed", nil, "Please provide both username and password")
		case utils.ErrAdminUsernameTooLong:
			api.SendResponse(w, http.StatusBadRequest, "Login failed", nil, "Provided username is too long")
		case utils.ErrAdminPasswordTooLong:
			api.SendResponse(w, http.StatusBadRequest, "Login failed", nil, "Provided password is too long")
		default:
			api.SendResponse(w, http.StatusInternalServerError, "Login failed", nil, "An unexpected error occurred")
		}
		return
	}

	token, err := h.adminUseCase.Login(r.Context(), input.Username, input.Password)
	if err != nil {
		switch err {
		case utils.ErrInvalidAdminCredentials:
			api.SendResponse(w, http.StatusUnauthorized, "Login failed", nil, "Invalid username or password")
		case utils.ErrRetreivingAdminUsername:
			api.SendResponse(w, http.StatusInternalServerError, "Login failed", nil, "Error while retrieving data of the given username")
		case utils.ErrGenerateJWTTokenWithRole:
			api.SendResponse(w, http.StatusInternalServerError, "Login failed", nil, "Error while generating JWT token")
		default:
			api.SendResponse(w, http.StatusInternalServerError, "Login failed", nil, "An unexpected error occurred")
		}
		return
	}

	response := AdminLoginResponse{Token: token}
	api.SendResponse(w, http.StatusOK, "Login successful", response, "")
}

func (h *AdminHandler) Logout(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	token, err := validator.ValidateAuthHeaderAndReturnToken(authHeader)
	if err != nil {
		switch err {
		case utils.ErrMissingAuthToken:
			api.SendResponse(w, http.StatusUnauthorized, "Logout failed", nil, "Missing authorization token")
		case utils.ErrAuthHeaderFormat:
			api.SendResponse(w, http.StatusUnauthorized, "Logout failed", nil, "Invalid authorization header format")
		case utils.ErrEmptyToken:
			api.SendResponse(w, http.StatusUnauthorized, "Logout failed", nil, "Empty token")
		default:
			api.SendResponse(w, http.StatusInternalServerError, "Logout failed", nil, "An unexpected error occured")
		}
		return
	}

	err = h.adminUseCase.Logout(r.Context(), token)
	if err != nil {
		switch err {
		case utils.ErrInvalidAdminToken:
			api.SendResponse(w, http.StatusUnauthorized, "Logout failed", nil, "Invalid token")
		case utils.ErrTokenAlreadyBlacklisted:
			api.SendResponse(w, http.StatusBadRequest, "Logout failed", nil, "Invalid token: token already invalidated")
		case utils.ErrCheckTokenBlacklisted:
			api.SendResponse(w, http.StatusInternalServerError, "Logout failed", nil, "Error while checking whether token is blacklisted")
		case utils.ErrTokenExpired:
			api.SendResponse(w, http.StatusBadRequest, "Logout failed", nil, "Invalid token: token expired")
		case utils.ErrInvalidExpirationClaim:
			api.SendResponse(w, http.StatusBadRequest, "Logout failed", nil, "Invalid token: expiration claim is invalid")
		default:
			api.SendResponse(w, http.StatusInternalServerError, "Logout failed", nil, "An unexpected error occured")
		}
		return
	}

	api.SendResponse(w, http.StatusOK, "Admin logged out successfully", nil, "")
}
