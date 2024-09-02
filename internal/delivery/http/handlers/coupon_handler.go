package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

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
