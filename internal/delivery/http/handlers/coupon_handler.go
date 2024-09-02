package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/mohamedfawas/rmshop-clean-architecture/internal/delivery/http/middleware"
	"github.com/mohamedfawas/rmshop-clean-architecture/internal/domain"
	"github.com/mohamedfawas/rmshop-clean-architecture/internal/usecase"
	"github.com/mohamedfawas/rmshop-clean-architecture/pkg/api"
	"github.com/mohamedfawas/rmshop-clean-architecture/pkg/utils"
)

type CouponHandler struct {
	couponUseCase usecase.CouponUseCase
}

func NewCouponHandler(couponUseCase usecase.CouponUseCase) *CouponHandler {
	return &CouponHandler{couponUseCase: couponUseCase}
}

func (h *CouponHandler) CreateCoupon(w http.ResponseWriter, r *http.Request) {
	var input domain.CreateCouponInput
	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		api.SendResponse(w, http.StatusBadRequest, "Failed to create coupon", nil, "Invalid request body")
		return
	}

	// trim whitespace
	input.Code = strings.TrimSpace(input.Code)

	coupon, err := h.couponUseCase.CreateCoupon(r.Context(), input)
	if err != nil {
		switch err {
		case utils.ErrInvalidCouponCode:
			api.SendResponse(w, http.StatusBadRequest, "Failed to create coupon", nil, "Invalid coupon code format")
		case utils.ErrInvalidDiscountPercentage:
			api.SendResponse(w, http.StatusBadRequest, "Failed to create coupon", nil, "Invalid discount percentage")
		case utils.ErrInvalidMinOrderAmount:
			api.SendResponse(w, http.StatusBadRequest, "Failed to create coupon", nil, "Invalid minimum order amount")
		case utils.ErrInvalidExpiryDate:
			api.SendResponse(w, http.StatusBadRequest, "Failed to create coupon", nil, "Invalid expiry date")
		case utils.ErrDuplicateCouponCode:
			api.SendResponse(w, http.StatusConflict, "Failed to create coupon", nil, "Coupon code already exists")
		default:
			api.SendResponse(w, http.StatusInternalServerError, "Failed to create coupon", nil, "An unexpected error occurred")
		}
		return
	}

	api.SendResponse(w, http.StatusCreated, "Coupon created successfully", coupon, "")
}

func (h *CouponHandler) ApplyCoupon(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(int64)
	if !ok {
		api.SendResponse(w, http.StatusUnauthorized, "Failed to apply coupon", nil, "User not authenticated")
		return
	}

	var input domain.ApplyCouponInput
	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		api.SendResponse(w, http.StatusBadRequest, "Failed to apply coupon", nil, "Invalid request body")
		return
	}

	if input.CouponCode == "" {
		api.SendResponse(w, http.StatusBadRequest, "Failed to apply coupon", nil, "Coupon code is required")
		return
	}

	response, err := h.couponUseCase.ApplyCoupon(r.Context(), userID, input)
	if err != nil {
		switch err {
		case utils.ErrInvalidCouponCode:
			api.SendResponse(w, http.StatusBadRequest, "Failed to apply coupon", nil, "Invalid coupon code")
		case utils.ErrCouponExpired:
			api.SendResponse(w, http.StatusBadRequest, "Failed to apply coupon", nil, "Coupon has expired")
		case utils.ErrCouponInactive:
			api.SendResponse(w, http.StatusBadRequest, "Failed to apply coupon", nil, "This coupon is no longer active")
		case utils.ErrOrderTotalBelowMinimum:
			api.SendResponse(w, http.StatusBadRequest, "Failed to apply coupon", nil, "Minimum order amount not met for this coupon")
		case utils.ErrCouponAlreadyApplied:
			api.SendResponse(w, http.StatusBadRequest, "Failed to apply coupon", nil, "A coupon is already applied to this cart")
		case utils.ErrEmptyCart:
			api.SendResponse(w, http.StatusBadRequest, "Failed to apply coupon", nil, "Cannot apply coupon to an empty cart")
		default:
			api.SendResponse(w, http.StatusInternalServerError, "Failed to apply coupon", nil, "An unexpected error occurred")
		}
		return
	}

	api.SendResponse(w, http.StatusOK, "Coupon applied successfully", response, "")
}
