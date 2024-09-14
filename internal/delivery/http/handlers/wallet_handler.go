package handlers

import (
	"net/http"

	"github.com/mohamedfawas/rmshop-clean-architecture/internal/delivery/http/middleware"
	"github.com/mohamedfawas/rmshop-clean-architecture/internal/usecase"
	"github.com/mohamedfawas/rmshop-clean-architecture/pkg/api"
	"github.com/mohamedfawas/rmshop-clean-architecture/pkg/utils"
)

type WalletHandler struct {
	walletUseCase usecase.WalletUseCase
}

func NewWalletHandler(walletUseCase usecase.WalletUseCase) *WalletHandler {
	return &WalletHandler{walletUseCase: walletUseCase}
}

func (h *WalletHandler) GetWalletBalance(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(int64)
	if !ok {
		api.SendResponse(w, http.StatusUnauthorized, "Failed to get wallet balance", nil, "User not authenticated")
		return
	}

	balance, err := h.walletUseCase.GetBalance(r.Context(), userID)
	if err != nil {
		switch err {
		case utils.ErrUserNotFound:
			api.SendResponse(w, http.StatusNotFound, "Failed to get wallet balance", nil, "User not found")
		case utils.ErrWalletNotInitialized:
			api.SendResponse(w, http.StatusOK, "Wallet not initialized", map[string]interface{}{"balance": 0.00}, "")
		default:
			api.SendResponse(w, http.StatusInternalServerError, "Failed to get wallet balance", nil, "An unexpected error occurred")
		}
		return
	}

	api.SendResponse(w, http.StatusOK, "Wallet balance retrieved successfully", map[string]interface{}{"balance": balance}, "")
}
