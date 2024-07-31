package handlers

import (
	"encoding/json"
	"net/http"

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

func (h *AdminHandler) Login(w http.ResponseWriter, r *http.Request) {
	var input AdminLoginInput
	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	token, err := h.adminUseCase.Login(r.Context(), input.Username, input.Password)
	if err != nil {
		switch err {
		case usecase.ErrInvalidAdminCredentials:
			http.Error(w, "Invalid username or password", http.StatusUnauthorized)
		default:
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
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
		if err == usecase.ErrInvalidAdminToken {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
		} else {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Admin logged out successfully"})
}
