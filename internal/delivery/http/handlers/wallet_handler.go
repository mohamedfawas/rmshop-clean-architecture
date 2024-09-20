package handlers

import (
	"log"
	"net/http"
	"strconv"

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

func (h *WalletHandler) GetWalletTransactions(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(int64)
	if !ok {
		api.SendResponse(w, http.StatusUnauthorized, "Failed to get wallet transactions", nil, "User not authenticated")
		return
	}

	// Parse query parameters
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	sort := r.URL.Query().Get("sort")
	order := r.URL.Query().Get("order")
	transactionType := r.URL.Query().Get("type")

	// Validate and set default values
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}
	// Map "date" to "created_at" for sorting
	if sort == "date" || sort == "" {
		sort = "created_at"
	}
	if order == "" {
		order = "desc"
	}

	// Call use case
	transactions, totalCount, err := h.walletUseCase.GetWalletTransactions(r.Context(), userID, page, limit, sort, order, transactionType)
	if err != nil {
		log.Printf("error : %v", err)
		api.SendResponse(w, http.StatusInternalServerError, "Failed to get wallet transactions", nil, "An unexpected error occurred")
		return
	}

	response := map[string]interface{}{
		"transactions": transactions,
		"total_count":  totalCount,
		"page":         page,
		"limit":        limit,
		"total_pages":  (totalCount + int64(limit) - 1) / int64(limit),
	}

	api.SendResponse(w, http.StatusOK, "Wallet transactions retrieved successfully", response, "")
}
