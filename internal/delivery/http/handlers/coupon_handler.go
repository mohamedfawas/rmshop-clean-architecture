package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
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
	response, err := h.couponUseCase.ApplyCoupon(r.Context(), userID, checkoutID, input.CouponCode)
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

func (h *CouponHandler) GetAllCoupons(w http.ResponseWriter, r *http.Request) {
	// Check if the user is an admin
	userRole, ok := r.Context().Value(middleware.UserRoleKey).(string)
	if !ok || userRole != "admin" {
		api.SendResponse(w, http.StatusForbidden, "Access denied", nil, "Admin privileges required")
		return
	}

	// Parse query parameters
	params := parseGetCouponsQueryParams(r)

	// Call use case method
	coupons, totalCount, err := h.couponUseCase.GetAllCoupons(r.Context(), params)
	if err != nil {
		api.SendResponse(w, http.StatusInternalServerError, "Failed to retrieve coupons", nil, err.Error())
		return
	}

	// Prepare response
	response := map[string]interface{}{
		"coupons":     coupons,
		"total_count": totalCount,
		"page":        params.Page,
		"limit":       params.Limit,
		"total_pages": (totalCount + int64(params.Limit) - 1) / int64(params.Limit),
	}

	api.SendResponse(w, http.StatusOK, "Coupons retrieved successfully", response, "")
}

func parseGetCouponsQueryParams(r *http.Request) domain.CouponQueryParams {
	params := domain.CouponQueryParams{
		Page:  1,
		Limit: 10,
	}

	if page, err := strconv.Atoi(r.URL.Query().Get("page")); err == nil && page > 0 {
		params.Page = page
	}

	if limit, err := strconv.Atoi(r.URL.Query().Get("limit")); err == nil && limit > 0 {
		params.Limit = limit
	}

	params.Sort = r.URL.Query().Get("sort")
	params.Order = r.URL.Query().Get("order")
	params.Status = r.URL.Query().Get("status")
	params.Search = r.URL.Query().Get("search")

	if minDiscount, err := strconv.ParseFloat(r.URL.Query().Get("min_discount"), 64); err == nil {
		params.MinDiscount = &minDiscount
	}

	if maxDiscount, err := strconv.ParseFloat(r.URL.Query().Get("max_discount"), 64); err == nil {
		params.MaxDiscount = &maxDiscount
	}

	return params
}

func (h *CouponHandler) UpdateCoupon(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	couponID, err := strconv.ParseInt(vars["coupon_id"], 10, 64)
	if err != nil {
		api.SendResponse(w, http.StatusBadRequest, "Failed to update coupon", nil, "Invalid coupon ID")
		return
	}

	var updateInput domain.CouponUpdateInput
	err = json.NewDecoder(r.Body).Decode(&updateInput)
	if err != nil {
		api.SendResponse(w, http.StatusBadRequest, "Failed to update coupon", nil, "Invalid request body")
		return
	}

	updatedCoupon, err := h.couponUseCase.UpdateCoupon(r.Context(), couponID, updateInput)
	if err != nil {
		switch err {
		case utils.ErrCouponNotFound:
			api.SendResponse(w, http.StatusNotFound, "Failed to update coupon", nil, "Coupon not found")
		case utils.ErrInvalidCouponCode:
			api.SendResponse(w, http.StatusBadRequest, "Failed to update coupon", nil, "Invalid coupon code format")
		case utils.ErrInvalidDiscountPercentage:
			api.SendResponse(w, http.StatusBadRequest, "Failed to update coupon", nil, "Invalid discount percentage")
		case utils.ErrInvalidMinOrderAmount:
			api.SendResponse(w, http.StatusBadRequest, "Failed to update coupon", nil, "Invalid minimum order amount")
		case utils.ErrInvalidExpiryDate:
			api.SendResponse(w, http.StatusBadRequest, "Failed to update coupon", nil, "Invalid expiry date")
		default:
			api.SendResponse(w, http.StatusInternalServerError, "Failed to update coupon", nil, "An unexpected error occurred")
		}
		return
	}

	api.SendResponse(w, http.StatusOK, "Coupon updated successfully", updatedCoupon, "")
}

func (h *CouponHandler) SoftDeleteCoupon(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	couponID, err := strconv.ParseInt(vars["coupon_id"], 10, 64)
	if err != nil {
		api.SendResponse(w, http.StatusBadRequest, "Failed to delete coupon", nil, "Invalid coupon ID format")
		return
	}

	err = h.couponUseCase.SoftDeleteCoupon(r.Context(), couponID)
	if err != nil {
		switch err {
		case utils.ErrCouponNotFound:
			api.SendResponse(w, http.StatusNotFound, "Failed to delete coupon", nil, "Coupon not found")
		case utils.ErrCouponAlreadyDeleted:
			api.SendResponse(w, http.StatusBadRequest, "Failed to delete coupon", nil, "Coupon is already soft deleted")
		case utils.ErrCouponInUse:
			api.SendResponse(w, http.StatusBadRequest, "Failed to delete coupon", nil, "Cannot delete coupon as it is currently in use")
		default:
			log.Printf("error : %v", err)
			api.SendResponse(w, http.StatusInternalServerError, "Failed to delete coupon", nil, "An unexpected error occurred")
		}
		return
	}

	api.SendResponse(w, http.StatusOK, "Coupon successfully soft deleted", nil, "")
}
