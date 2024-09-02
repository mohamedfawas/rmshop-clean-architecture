package handlers

import (
	"net/http"

	"github.com/mohamedfawas/rmshop-clean-architecture/internal/delivery/http/middleware"
	"github.com/mohamedfawas/rmshop-clean-architecture/internal/usecase"
	"github.com/mohamedfawas/rmshop-clean-architecture/pkg/api"
	"github.com/mohamedfawas/rmshop-clean-architecture/pkg/utils"
)

type CheckoutHandler struct {
	checkoutUseCase usecase.CheckoutUseCase
}

func NewCheckoutHandler(checkoutUseCase usecase.CheckoutUseCase) *CheckoutHandler {
	return &CheckoutHandler{checkoutUseCase: checkoutUseCase}
}

func (h *CheckoutHandler) CreateCheckout(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(int64)
	if !ok {
		api.SendResponse(w, http.StatusUnauthorized, "Failed to create checkout", nil, "User not authenticated")
		return
	}

	session, err := h.checkoutUseCase.CreateCheckout(r.Context(), userID)
	if err != nil {
		switch err {
		case utils.ErrEmptyCart:
			api.SendResponse(w, http.StatusBadRequest, "Failed to create checkout", nil, "Cart is empty")
		case utils.ErrInsufficientStock:
			api.SendResponse(w, http.StatusBadRequest, "Failed to create checkout", nil, "Insufficient stock for one or more items")
		default:
			api.SendResponse(w, http.StatusInternalServerError, "Failed to create checkout", nil, "An unexpected error occurred")
		}
		return
	}

	api.SendResponse(w, http.StatusCreated, "Checkout created successfully", session, "")
}
