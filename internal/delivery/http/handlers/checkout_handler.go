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

func (h *CheckoutHandler) UpdateCheckoutAddress(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(int64)
	if !ok {
		api.SendResponse(w, http.StatusUnauthorized, "Failed to update address", nil, "User not authenticated")
		return
	}

	vars := mux.Vars(r)
	checkoutID, err := strconv.ParseInt(vars["checkout_id"], 10, 64)
	if err != nil {
		api.SendResponse(w, http.StatusBadRequest, "Failed to update address", nil, "Invalid checkout ID")
		return
	}

	var input struct {
		AddressID int64 `json:"address_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		api.SendResponse(w, http.StatusBadRequest, "Failed to update address", nil, "Invalid request body")
		return
	}

	if input.AddressID <= 0 {
		api.SendResponse(w, http.StatusBadRequest, "Failed to update address", nil, "Invalid address ID")
		return
	}

	updatedCheckout, err := h.checkoutUseCase.UpdateCheckoutAddress(r.Context(), userID, checkoutID, input.AddressID)
	if err != nil {
		switch err {
		case utils.ErrCheckoutNotFound:
			api.SendResponse(w, http.StatusNotFound, "Failed to update address", nil, "Checkout not found")
		case utils.ErrUnauthorized:
			api.SendResponse(w, http.StatusForbidden, "Failed to update address", nil, "Unauthorized access")
		case utils.ErrInvalidCheckoutState:
			api.SendResponse(w, http.StatusBadRequest, "Failed to update address", nil, "Checkout is not in a valid state for address update")
		case utils.ErrAddressNotFound:
			api.SendResponse(w, http.StatusNotFound, "Failed to update address", nil, "Address not found")
		case utils.ErrAddressNotBelongToUser:
			api.SendResponse(w, http.StatusForbidden, "Failed to update address", nil, "Address does not belong to the user")
		default:
			api.SendResponse(w, http.StatusInternalServerError, "Failed to update address", nil, "An unexpected error occurred")
		}
		return
	}

	api.SendResponse(w, http.StatusOK, "Address updated successfully", updatedCheckout, "")
}
func (h *CheckoutHandler) GetCheckoutSummary(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(int64)
	if !ok {
		api.SendResponse(w, http.StatusUnauthorized, "Failed to get checkout summary", nil, "User not authenticated")
		return
	}

	vars := mux.Vars(r)
	checkoutID, err := strconv.ParseInt(vars["checkout_id"], 10, 64)
	if err != nil {
		api.SendResponse(w, http.StatusBadRequest, "Failed to get checkout summary", nil, "Invalid checkout ID")
		return
	}

	summary, err := h.checkoutUseCase.GetCheckoutSummary(r.Context(), userID, checkoutID)
	if err != nil {
		switch err {
		case utils.ErrEmptyCart:
			api.SendResponse(w, http.StatusBadRequest, "Failed to get checkout summary", nil, "Checkout is empty")
		case utils.ErrCheckoutNotFound:
			api.SendResponse(w, http.StatusNotFound, "Failed to get checkout summary", nil, "Checkout not found")
		case utils.ErrUnauthorized:
			api.SendResponse(w, http.StatusForbidden, "Failed to get checkout summary", nil, "You don't have permission to access this checkout")
		default:
			api.SendResponse(w, http.StatusInternalServerError, "Failed to get checkout summary", nil, "An unexpected error occurred")
		}
		return
	}

	api.SendResponse(w, http.StatusOK, "Checkout summary retrieved successfully", summary, "")
}

// Add this method to your CheckoutHandler struct
func (h *CheckoutHandler) PlaceOrder(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from context (set by auth middleware)
	userID, ok := r.Context().Value(middleware.UserIDKey).(int64)
	if !ok {
		api.SendResponse(w, http.StatusUnauthorized, "Failed to place order", nil, "User not authenticated")
		return
	}

	// Extract checkout ID from URL
	vars := mux.Vars(r)
	checkoutID, err := strconv.ParseInt(vars["checkout_id"], 10, 64)
	if err != nil {
		api.SendResponse(w, http.StatusBadRequest, "Failed to place order", nil, "Invalid checkout ID")
		return
	}

	// Parse request body
	var input struct {
		PaymentMethod string `json:"payment_method"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		api.SendResponse(w, http.StatusBadRequest, "Failed to place order", nil, "Invalid request body")
		return
	}

	// Validate payment method
	if input.PaymentMethod != "COD" && input.PaymentMethod != "razorpay" {
		api.SendResponse(w, http.StatusBadRequest, "Failed to place order", nil, "Invalid payment method")
		return
	}

	// Call use case method to place the order
	order, err := h.checkoutUseCase.PlaceOrder(r.Context(), userID, checkoutID, input.PaymentMethod)
	if err != nil {
		switch err {
		case utils.ErrCheckoutNotFound:
			api.SendResponse(w, http.StatusNotFound, "Failed to place order", nil, "Checkout not found")
		case utils.ErrUnauthorized:
			api.SendResponse(w, http.StatusForbidden, "Failed to place order", nil, "Unauthorized access to this checkout")
		case utils.ErrEmptyCart:
			api.SendResponse(w, http.StatusBadRequest, "Failed to place order", nil, "Cannot place order with empty cart")
		case utils.ErrInsufficientStock:
			api.SendResponse(w, http.StatusBadRequest, "Failed to place order", nil, "Insufficient stock for one or more items")
		case utils.ErrInvalidAddress:
			api.SendResponse(w, http.StatusBadRequest, "Failed to place order", nil, "Invalid or missing delivery address")
		case utils.ErrOrderAlreadyPlaced:
			api.SendResponse(w, http.StatusConflict, "Failed to place order", nil, "Order has already been placed")
		default:
			api.SendResponse(w, http.StatusInternalServerError, "Failed to place order", nil, "An unexpected error occurred")
		}
		return
	}

	api.SendResponse(w, http.StatusCreated, "Order placed successfully", order, "")
}
