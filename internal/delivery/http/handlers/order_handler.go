package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/mohamedfawas/rmshop-clean-architecture/internal/delivery/http/middleware"
	"github.com/mohamedfawas/rmshop-clean-architecture/internal/domain"
	"github.com/mohamedfawas/rmshop-clean-architecture/internal/usecase"
	"github.com/mohamedfawas/rmshop-clean-architecture/pkg/api"
	"github.com/mohamedfawas/rmshop-clean-architecture/pkg/utils"
)

type OrderHandler struct {
	orderUseCase usecase.OrderUseCase
}

func NewOrderHandler(orderUseCase usecase.OrderUseCase) *OrderHandler {
	return &OrderHandler{orderUseCase: orderUseCase}
}

func (h *OrderHandler) GetOrderConfirmation(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from context (set by auth middleware)
	userID, ok := r.Context().Value(middleware.UserIDKey).(int64)
	if !ok {
		api.SendResponse(w, http.StatusUnauthorized, "Failed to get order", nil, "User not authenticated")
		return
	}

	// Extract order ID from URL
	vars := mux.Vars(r)
	orderID, err := strconv.ParseInt(vars["order_id"], 10, 64)
	if err != nil {
		api.SendResponse(w, http.StatusBadRequest, "Failed to get order", nil, "Invalid order ID")
		return
	}

	// Call use case method to get the order
	order, err := h.orderUseCase.GetOrderByID(r.Context(), userID, orderID)
	if err != nil {
		switch err {
		case utils.ErrOrderNotFound:
			api.SendResponse(w, http.StatusNotFound, "Failed to get order", nil, "Order not found")
		case utils.ErrUnauthorized:
			api.SendResponse(w, http.StatusForbidden, "Failed to get order", nil, "You don't have permission to access this order")
		default:
			api.SendResponse(w, http.StatusInternalServerError, "Failed to get order", nil, "An unexpected error occurred")
		}
		return
	}

	api.SendResponse(w, http.StatusOK, "Order retrieved successfully", order, "")
}

func (h *OrderHandler) GetUserOrders(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from context (set by auth middleware)
	userID, ok := r.Context().Value(middleware.UserIDKey).(int64)
	if !ok {
		api.SendResponse(w, http.StatusUnauthorized, "Failed to get orders", nil, "User not authenticated")
		return
	}

	// Parse query parameters
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	sortBy := r.URL.Query().Get("sort")
	order := r.URL.Query().Get("order")
	status := r.URL.Query().Get("status")

	// Set default values if not provided
	if page == 0 {
		page = 1
	}
	if limit == 0 {
		limit = 10
	}

	// Call use case method to get the orders
	orders, totalCount, err := h.orderUseCase.GetUserOrders(r.Context(), userID, page, limit, sortBy, order, status)
	if err != nil {
		switch err {
		case utils.ErrInvalidPaginationParams:
			api.SendResponse(w, http.StatusBadRequest, "Failed to get orders", nil, "Invalid pagination parameters")
		default:
			api.SendResponse(w, http.StatusInternalServerError, "Failed to get orders", nil, "An unexpected error occurred")
		}
		return
	}

	// Prepare pagination metadata
	totalPages := (totalCount + int64(limit) - 1) / int64(limit)
	nextPage := ""
	if int64(page) < totalPages {
		nextPage = fmt.Sprintf("/user/orders?page=%d&limit=%d", page+1, limit)
	}

	response := map[string]interface{}{
		"orders":      orders,
		"total_count": totalCount,
		"page":        page,
		"limit":       limit,
		"total_pages": totalPages,
		"next_page":   nextPage,
	}

	api.SendResponse(w, http.StatusOK, "Orders retrieved successfully", response, "")
}

func (h *OrderHandler) CancelOrder(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from context (set by auth middleware)
	userID, ok := r.Context().Value(middleware.UserIDKey).(int64)
	if !ok {
		api.SendResponse(w, http.StatusUnauthorized, "Failed to cancel order", nil, "User not authenticated")
		return
	}

	// Extract order ID from URL
	vars := mux.Vars(r)
	orderID, err := strconv.ParseInt(vars["orderId"], 10, 64)
	if err != nil {
		api.SendResponse(w, http.StatusBadRequest, "Failed to cancel order", nil, "Invalid order ID format")
		return
	}

	// Call use case method to cancel the order
	result, err := h.orderUseCase.CancelOrder(r.Context(), userID, orderID)
	if err != nil {
		switch err {
		case utils.ErrOrderNotFound:
			api.SendResponse(w, http.StatusNotFound, "Failed to cancel order", nil, "Order not found")
		case utils.ErrUnauthorized:
			api.SendResponse(w, http.StatusForbidden, "Failed to cancel order", nil, "You are not authorized to cancel this order")
		case utils.ErrOrderAlreadyCancelled:
			api.SendResponse(w, http.StatusBadRequest, "Failed to cancel order", nil, "Order is already cancelled")
		case utils.ErrOrderNotCancellable:
			api.SendResponse(w, http.StatusBadRequest, "Failed to cancel order", nil, "Order cannot be cancelled in its current state")
		case utils.ErrCancellationPeriodExpired:
			api.SendResponse(w, http.StatusBadRequest, "Failed to cancel order", nil, "Cancellation period has expired for this order")
		default:
			api.SendResponse(w, http.StatusInternalServerError, "Failed to cancel order", nil, "An unexpected error occurred")
		}
		return
	}

	api.SendResponse(w, http.StatusOK, "Order cancelled successfully", result, "")
}

func (h *OrderHandler) GetOrders(w http.ResponseWriter, r *http.Request) {
	// Extract query parameters
	params := domain.OrderQueryParams{
		Page:      1,
		Limit:     10,
		SortBy:    "created_at",
		SortOrder: "desc",
	}

	if page, err := strconv.Atoi(r.URL.Query().Get("page")); err == nil && page > 0 {
		params.Page = page
	}

	if limit, err := strconv.Atoi(r.URL.Query().Get("limit")); err == nil && limit > 0 {
		params.Limit = limit
	}

	params.SortBy = r.URL.Query().Get("sort")
	params.SortOrder = r.URL.Query().Get("order")
	params.Status = r.URL.Query().Get("status")

	if customerID, err := strconv.ParseInt(r.URL.Query().Get("customer_id"), 10, 64); err == nil {
		params.CustomerID = customerID
	}

	if startDate, err := time.Parse("2006-01-02", r.URL.Query().Get("start_date")); err == nil {
		params.StartDate = &startDate
	}

	if endDate, err := time.Parse("2006-01-02", r.URL.Query().Get("end_date")); err == nil {
		params.EndDate = &endDate
	}

	if fields := r.URL.Query().Get("fields"); fields != "" {
		params.Fields = strings.Split(fields, ",")
	}

	// Call use case
	orders, total, err := h.orderUseCase.GetOrders(r.Context(), params)
	if err != nil {
		switch err {
		case utils.ErrInvalidPaginationParams:
			api.SendResponse(w, http.StatusBadRequest, "Invalid pagination parameters", nil, err.Error())
		default:
			api.SendResponse(w, http.StatusInternalServerError, "Failed to retrieve orders", nil, "An unexpected error occurred")
		}
		return
	}

	response := map[string]interface{}{
		"orders":      orders,
		"total_count": total,
		"page":        params.Page,
		"limit":       params.Limit,
		"total_pages": (total + int64(params.Limit) - 1) / int64(params.Limit),
	}

	api.SendResponse(w, http.StatusOK, "Orders retrieved successfully", response, "")
}

func (h *OrderHandler) UpdateOrderStatus(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	orderID, err := strconv.ParseInt(vars["orderId"], 10, 64)
	if err != nil {
		api.SendResponse(w, http.StatusBadRequest, "Failed to update order status", nil, "Invalid order ID")
		return
	}

	var input struct {
		Status string `json:"order_status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		api.SendResponse(w, http.StatusBadRequest, "Failed to update order status", nil, "Invalid request body")
		return
	}

	result, err := h.orderUseCase.UpdateOrderStatus(r.Context(), orderID, input.Status)
	if err != nil {
		switch err {
		case utils.ErrOrderNotFound:
			api.SendResponse(w, http.StatusNotFound, "Failed to update order status", nil, "Order not found")
		case utils.ErrUnauthorized:
			api.SendResponse(w, http.StatusForbidden, "Failed to update order status", nil, "Unauthorized action")
		case utils.ErrInvalidOrderStatus:
			api.SendResponse(w, http.StatusBadRequest, "Failed to update order status", nil, "Invalid order status")
		case utils.ErrOrderNotCancellable:
			api.SendResponse(w, http.StatusBadRequest, "Failed to cancel order", nil, "Order cannot be cancelled in its current state")
		case utils.ErrOrderAlreadyCancelled:
			api.SendResponse(w, http.StatusBadRequest, "Failed to cancel order", nil, "Order is already cancelled")
		default:
			api.SendResponse(w, http.StatusInternalServerError, "Failed to update order status", nil, "An unexpected error occurred")
		}
		return
	}

	api.SendResponse(w, http.StatusOK, "Order status updated successfully", result, "")
}
