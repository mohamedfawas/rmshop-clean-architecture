package handlers

import (
	"encoding/json"
	"net/http"

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

	session, err := h.checkoutUseCase.CreateOrUpdateCheckout(r.Context(), userID)
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

	api.SendResponse(w, http.StatusCreated, "Checkout created or updated successfully", session, "")
}

func (h *CheckoutHandler) ApplyCoupon(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userID, ok := r.Context().Value(middleware.UserIDKey).(int64)
	if !ok {
		api.SendResponse(w, http.StatusUnauthorized, "Failed to apply coupon", nil, "User not authenticated")
		return
	}

	// Parse request body
	var input domain.ApplyCouponInput
	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		api.SendResponse(w, http.StatusBadRequest, "Failed to apply coupon", nil, "Invalid request body")
		return
	}

	// Ensure coupon code is provided
	if input.CouponCode == "" {
		api.SendResponse(w, http.StatusBadRequest, "Failed to apply coupon", nil, "Coupon code is required")
		return
	}

	// Apply coupon
	response, err := h.checkoutUseCase.ApplyCoupon(r.Context(), userID, input.CouponCode)
	if err != nil {
		switch err {
		case utils.ErrEmptyCart:
			api.SendResponse(w, http.StatusBadRequest, "Failed to apply coupon", nil, "Cannot apply coupon to an empty cart")
		case utils.ErrCouponAlreadyApplied:
			api.SendResponse(w, http.StatusBadRequest, "Failed to apply coupon", nil, "A coupon is already applied to this cart")
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
	// Extract the user id from the context values
	userID, ok := r.Context().Value(middleware.UserIDKey).(int64)
	if !ok {
		api.SendResponse(w, http.StatusUnauthorized, "Failed to update address", nil, "User not authenticated")
		return
	}

	// Input address id to use for shipping address
	var input struct {
		AddressID int64 `json:"address_id"`
	}
	// extract given address id from the request body
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		api.SendResponse(w, http.StatusBadRequest, "Failed to update address", nil, "Invalid request body")
		return
	}

	// validate given address id
	if input.AddressID <= 0 {
		api.SendResponse(w, http.StatusBadRequest, "Failed to update address", nil, "Invalid address ID")
		return
	}

	updatedCheckout, err := h.checkoutUseCase.UpdateCheckoutAddress(r.Context(), userID, input.AddressID)
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
	// Extract the user id from the context values
	userID, ok := r.Context().Value(middleware.UserIDKey).(int64)
	if !ok {
		api.SendResponse(w, http.StatusUnauthorized, "Failed to get checkout summary", nil, "User not authenticated")
		return
	}

	// Call the GetCheckoutSummary method from the usecase layer
	summary, err := h.checkoutUseCase.GetCheckoutSummary(r.Context(), userID)
	if err != nil {
		switch err {
		case utils.ErrUnauthorized:
			api.SendResponse(w, http.StatusForbidden, "Failed to get checkout summary", nil, "You don't have permission to access this checkout")
		case utils.ErrCheckoutNotFound:
			api.SendResponse(w, http.StatusForbidden, "Failed to get checkout summary", nil, "Checkout not created for the authenticated user")
		case utils.ErrCartUpdatedAfterCreatingCheckoutSession:
			api.SendResponse(w, http.StatusConflict, "Failed to get checkout summary", nil, "Cart is updated after creating checkout session, please create the checkout session again.")
		default:
			api.SendResponse(w, http.StatusInternalServerError, "Failed to get checkout summary", nil, "An unexpected error occurred")
		}
		return
	}

	api.SendResponse(w, http.StatusOK, "Checkout summary retrieved successfully", summary, "")
}

func (h *CheckoutHandler) RemoveAppliedCoupon(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from context (set by auth middleware)
	userID, ok := r.Context().Value(middleware.UserIDKey).(int64)
	if !ok {
		api.SendResponse(w, http.StatusUnauthorized, "Failed to remove coupon", nil, "User not authenticated")
		return
	}

	// Call use case method to remove the coupon
	updatedCheckout, err := h.checkoutUseCase.RemoveAppliedCoupon(r.Context(), userID)
	if err != nil {
		switch err {
		case utils.ErrCheckoutNotFound:
			api.SendResponse(w, http.StatusNotFound, "Failed to remove coupon", nil, "No active checkout session found")
		case utils.ErrNoCouponApplied:
			api.SendResponse(w, http.StatusBadRequest, "Failed to remove coupon", nil, "No coupon is applied to this checkout")
		case utils.ErrCheckoutCompleted:
			api.SendResponse(w, http.StatusBadRequest, "Failed to remove coupon", nil, "Checkout is already completed")
		default:
			api.SendResponse(w, http.StatusInternalServerError, "Failed to remove coupon", nil, "An unexpected error occurred")
		}
		return
	}

	api.SendResponse(w, http.StatusOK, "Coupon removed successfully", updatedCheckout, "")
}
