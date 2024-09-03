package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/mohamedfawas/rmshop-clean-architecture/internal/delivery/http/middleware"
	"github.com/mohamedfawas/rmshop-clean-architecture/internal/domain"
	"github.com/mohamedfawas/rmshop-clean-architecture/internal/usecase"
	"github.com/mohamedfawas/rmshop-clean-architecture/pkg/api"
	"github.com/mohamedfawas/rmshop-clean-architecture/pkg/utils"
)

type CheckoutHandler struct {
	checkoutUseCase usecase.CheckoutUseCase
	couponUseCase   usecase.CouponUseCase
}

func NewCheckoutHandler(checkoutUseCase usecase.CheckoutUseCase, couponUseCase usecase.CouponUseCase) *CheckoutHandler {
	return &CheckoutHandler{checkoutUseCase: checkoutUseCase,
		couponUseCase: couponUseCase}
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

func (h *CheckoutHandler) ApplyCoupon(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userID, ok := r.Context().Value(middleware.UserIDKey).(int64)
	if !ok {
		api.SendResponse(w, http.StatusUnauthorized, "Failed to apply coupon", nil, "User not authenticated")
		return
	}

	// Get checkout ID from URL
	vars := mux.Vars(r)
	checkoutID, err := strconv.ParseInt(vars["checkout_id"], 10, 64)
	if err != nil {
		api.SendResponse(w, http.StatusBadRequest, "Failed to apply coupon", nil, "Invalid checkout ID")
		return
	}

	// Parse request body
	var input domain.ApplyCouponInput
	err = json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		api.SendResponse(w, http.StatusBadRequest, "Failed to apply coupon", nil, "Invalid request body")
		return
	}

	// Validate coupon code
	if input.CouponCode == "" {
		api.SendResponse(w, http.StatusBadRequest, "Failed to apply coupon", nil, "Coupon code is required")
		return
	}

	// Apply coupon
	response, err := h.checkoutUseCase.ApplyCoupon(r.Context(), userID, checkoutID, input.CouponCode)
	if err != nil {
		switch err {
		case utils.ErrCheckoutNotFound:
			api.SendResponse(w, http.StatusNotFound, "Failed to apply coupon", nil, "Checkout not found")
		case utils.ErrUnauthorized:
			api.SendResponse(w, http.StatusForbidden, "Failed to apply coupon", nil, "Unauthorized access to this checkout")
		case utils.ErrEmptyCheckout:
			api.SendResponse(w, http.StatusBadRequest, "Failed to apply coupon", nil, "Cannot apply coupon to an empty checkout")
		case utils.ErrCouponAlreadyApplied:
			api.SendResponse(w, http.StatusBadRequest, "Failed to apply coupon", nil, "A coupon is already applied to this checkout")
		case utils.ErrInvalidCouponCode:
			api.SendResponse(w, http.StatusBadRequest, "Failed to apply coupon", nil, "Invalid coupon code")
		case utils.ErrCouponInactive:
			api.SendResponse(w, http.StatusBadRequest, "Failed to apply coupon", nil, "This coupon is no longer active")
		case utils.ErrCouponExpired:
			api.SendResponse(w, http.StatusBadRequest, "Failed to apply coupon", nil, "This coupon has expired")
		case utils.ErrOrderTotalBelowMinimum:
			api.SendResponse(w, http.StatusBadRequest, "Failed to apply coupon", nil, "Order total does not meet the minimum amount for this coupon")
		default:
			api.SendResponse(w, http.StatusInternalServerError, "Failed to apply coupon", nil, "An unexpected error occurred")
		}
		return
	}

	api.SendResponse(w, http.StatusOK, "Coupon applied successfully", response, "")
}
