package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
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

	// trim whitespace, case insensitive
	input.Code = strings.ToLower(strings.TrimSpace(input.Code))

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

func (h *CouponHandler) GetAllCoupons(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	params := parseGetCouponsQueryParams(r)

	// Call use case method
	coupons, totalCount, err := h.couponUseCase.GetAllCoupons(r.Context(), params)
	if err != nil {
		log.Printf("error : %v", err)
		api.SendResponse(w, http.StatusInternalServerError, "Failed to retrieve coupons", nil, "An unexpected error occured")
		return
	}
	response := map[string]interface{}{
		"coupons":     coupons,
		"total_count": totalCount,
		"page":        params.Page,
		"limit":       params.Limit,
		"total_pages": (totalCount + int64(params.Limit) - 1) / int64(params.Limit), //addition of - 1 ensures that you round up to the next page if there are any remaining items that don't fit perfectly into the full pages
		// The formula (a + b - 1) / b is a common way to perform integer division with rounding up
	}

	api.SendResponse(w, http.StatusOK, "Coupons retrieved successfully", response, "")
}

func parseGetCouponsQueryParams(r *http.Request) domain.CouponQueryParams {
	// Initialize the query parameters struct with default values
	// Default page is set to 1, and default limit (results per page) is set to 10.
	params := domain.CouponQueryParams{
		Page:  1,
		Limit: 10,
	}

	// Parse the "page" query parameter from the request URL.
	// If it's a valid integer and greater than 0, assign it to the params.Page.
	// Otherwise, keep the default value of 1.
	if page, err := strconv.Atoi(r.URL.Query().Get("page")); err == nil && page > 0 {
		params.Page = page
	}

	// Parse the "limit" query parameter from the request URL.
	// If it's a valid integer and greater than 0, assign it to the params.Limit.
	// Otherwise, keep the default value of 10.
	if limit, err := strconv.Atoi(r.URL.Query().Get("limit")); err == nil && limit > 0 {
		params.Limit = limit
	}

	// Assign the "sort" query parameter to the params.Sort field.
	// This defines which field to sort the results by (e.g., "created_at").
	params.Sort = r.URL.Query().Get("sort")

	// Assign the "order" query parameter to the params.Order field.
	// This defines the sorting order (e.g., "asc" or "desc").
	params.Order = r.URL.Query().Get("order")

	// Assign the "status" query parameter to the params.Status field.
	// This allows filtering by the coupon's status (e.g., "active")
	params.Status = r.URL.Query().Get("status")

	// Assign the "search" query parameter to the params.Search field.
	// This allows searching for coupons by a search term, such as a coupon code or description.
	params.Search = r.URL.Query().Get("search")

	// Parse the "min_discount" query parameter from the request URL.
	// If it's a valid float, assign it to the params.MinDiscount field.
	// This sets a minimum discount filter for the coupons.
	if minDiscount, err := strconv.ParseFloat(r.URL.Query().Get("min_discount"), 64); err == nil {
		params.MinDiscount = &minDiscount
	}

	// Parse the "max_discount" query parameter from the request URL.
	// If it's a valid float, assign it to the params.MaxDiscount field.
	// This sets a maximum discount filter for the coupons.
	if maxDiscount, err := strconv.ParseFloat(r.URL.Query().Get("max_discount"), 64); err == nil {
		params.MaxDiscount = &maxDiscount
	}

	// Return the populated CouponQueryParams struct with all the parsed values
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
