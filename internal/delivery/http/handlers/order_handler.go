package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

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

func (h *OrderHandler) GetOrderDetails(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from context
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

	// Call use case method to get the order with all details
	order, err := h.orderUseCase.GetOrderByID(r.Context(), userID, orderID)
	if err != nil {
		switch err {
		case utils.ErrOrderNotFound:
			api.SendResponse(w, http.StatusNotFound, "Failed to get order", nil, "Order not found")
		case utils.ErrUnauthorized:
			api.SendResponse(w, http.StatusForbidden, "Failed to get order", nil, "You don't have permission to access this order")
		default:
			log.Printf("Error retrieving order: %v", err)
			api.SendResponse(w, http.StatusInternalServerError, "Failed to get order", nil, "An unexpected error occurred")
		}
		return
	}

	// Prepare response data
	responseData := domain.OrderResponse{
		OrderID:        order.ID,
		UserID:         order.UserID,
		TotalAmount:    order.TotalAmount,
		DiscountAmount: order.DiscountAmount,
		FinalAmount:    order.FinalAmount,
		DeliveryStatus: order.DeliveryStatus,
		OrderStatus:    order.OrderStatus,
		CreatedAt:      order.CreatedAt,
		UpdatedAt:      order.UpdatedAt,
		Items:          order.Items,
		Payment:        order.Payment,
	}
	api.SendResponse(w, http.StatusOK, "Order retrieved successfully", responseData, "")
}

func (h *OrderHandler) GetUserOrders(w http.ResponseWriter, r *http.Request) {
	// Extract the user id from the context
	userID, ok := r.Context().Value(middleware.UserIDKey).(int64)
	if !ok {
		api.SendResponse(w, http.StatusUnauthorized, "Failed to get orders", nil, "User not authenticated")
		return
	}

	// Get the page number given in the query
	page, err := strconv.Atoi(r.URL.Query().Get("page"))
	// If page number is not given in the query or page number given is less than 1, then set the default value
	if err != nil || page < 1 {
		page = 1
	}

	// Call the usecase method
	orders, totalCount, err := h.orderUseCase.GetUserOrders(r.Context(), userID, page)
	if err != nil {
		api.SendResponse(w, http.StatusInternalServerError, "Failed to get orders", nil, "An unexpected error occurred")
		return
	}

	response := map[string]interface{}{
		"orders":      orders,
		"total_count": totalCount,
		"page":        page,
		"limit":       10,
		"total_pages": (totalCount + 9) / 10, // Ceiling division by 10
	}

	api.SendResponse(w, http.StatusOK, "Orders retrieved successfully", response, "")
}

func (h *OrderHandler) GetOrders(w http.ResponseWriter, r *http.Request) {
	// Get the page number from the query parameters
	pageStr := r.URL.Query().Get("page")
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1 // Default to first page if invalid
	}

	// Call use case to get orders
	orders, totalOrders, err := h.orderUseCase.GetOrders(r.Context(), page)
	if err != nil {
		api.SendResponse(w, http.StatusInternalServerError, "Failed to retrieve orders", nil, "An unexpected error occurred")
		return
	}

	// Calculate total pages
	ordersPerPage := 10
	totalPages := (totalOrders + int64(ordersPerPage) - 1) / int64(ordersPerPage)

	response := map[string]interface{}{
		"orders":      orders,
		"total_count": totalOrders,
		"page":        page,
		"total_pages": totalPages,
	}

	api.SendResponse(w, http.StatusOK, "Orders retrieved successfully", response, "")
}

func (h *OrderHandler) InitiateReturn(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from context (set by auth middleware)
	userID, ok := r.Context().Value(middleware.UserIDKey).(int64)
	if !ok {
		api.SendResponse(w, http.StatusUnauthorized, "Failed to initiate return", nil, "User not authenticated")
		return
	}

	// Extract order ID from URL
	vars := mux.Vars(r)
	orderID, err := strconv.ParseInt(vars["orderId"], 10, 64)
	if err != nil {
		api.SendResponse(w, http.StatusBadRequest, "Failed to initiate return", nil, "Invalid order ID")
		return
	}

	// Parse request body
	var input struct {
		Reason string `json:"reason"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		api.SendResponse(w, http.StatusBadRequest, "Failed to initiate return", nil, "Invalid request body")
		return
	}

	// Call use case method to initiate return
	returnRequest, err := h.orderUseCase.InitiateReturn(r.Context(), userID, orderID, input.Reason)
	if err != nil {
		switch err {
		case utils.ErrOrderNotFound:
			api.SendResponse(w, http.StatusNotFound, "Failed to initiate return", nil, "Order not found")
		case utils.ErrUnauthorized:
			api.SendResponse(w, http.StatusForbidden, "Failed to initiate return", nil, "You don't have permission to return this order")
		case utils.ErrOrderNotEligibleForReturn:
			api.SendResponse(w, http.StatusBadRequest, "Failed to initiate return", nil, "Order is not eligible for return")
		case utils.ErrReturnWindowExpired:
			api.SendResponse(w, http.StatusBadRequest, "Failed to initiate return", nil, "Return window has expired")
		case utils.ErrReturnAlreadyRequested:
			api.SendResponse(w, http.StatusConflict, "Failed to initiate return", nil, "Return request already exists for this order")
		case utils.ErrInvalidReturnReason:
			api.SendResponse(w, http.StatusBadRequest, "Failed to initiate return", nil, "Invalid return reason")
		default:
			api.SendResponse(w, http.StatusInternalServerError, "Failed to initiate return", nil, "An unexpected error occurred")
		}
		return
	}

	api.SendResponse(w, http.StatusCreated, "Return request initiated successfully", returnRequest, "")
}

func (h *OrderHandler) PlaceOrderCOD(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from context key
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

	// Call use case method to place the order
	order, err := h.orderUseCase.PlaceOrderCOD(r.Context(), userID, checkoutID)
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
			api.SendResponse(w, http.StatusConflict, "Failed to place order", nil, "Order has already been placed for this checkout")
		case utils.ErrCODLimitExceeded:
			api.SendResponse(w, http.StatusBadRequest, "Failed to place order", nil, "COD is not available for orders above Rs 1000")
		default:
			api.SendResponse(w, http.StatusInternalServerError, "Failed to place order", nil, "An unexpected error occurred")
		}
		return
	}

	api.SendResponse(w, http.StatusCreated, "Order placed successfully", order, "")
}

func (h *OrderHandler) GetOrderInvoice(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from context
	userID, ok := r.Context().Value(middleware.UserIDKey).(int64)
	if !ok {
		api.SendResponse(w, http.StatusUnauthorized, "Failed to generate invoice", nil, "User not authenticated")
		return
	}

	// Extract order ID from URL
	vars := mux.Vars(r)
	orderID, err := strconv.ParseInt(vars["orderId"], 10, 64)
	if err != nil {
		api.SendResponse(w, http.StatusBadRequest, "Failed to generate invoice", nil, "Invalid order ID")
		return
	}

	// Generate invoice by calling th usecase method
	pdfBytes, err := h.orderUseCase.GenerateInvoice(r.Context(), userID, orderID)
	if err != nil {
		switch err {
		case utils.ErrOrderNotFound:
			api.SendResponse(w, http.StatusNotFound, "Failed to generate invoice", nil, "Order not found")
		case utils.ErrUnauthorized:
			api.SendResponse(w, http.StatusForbidden, "Failed to generate invoice", nil, "You don't have permission to access this order")
		case utils.ErrCancelledOrder:
			api.SendResponse(w, http.StatusBadRequest, "Failed to generate invoice", nil, "Cannot generate invoice for a cancelled order")
		case utils.ErrUnpaidOrder:
			api.SendResponse(w, http.StatusBadRequest, "Failed to generate invoice", nil, "Cannot generate invoice for an unpaid order")
		case utils.ErrEmptyOrder:
			api.SendResponse(w, http.StatusBadRequest, "Failed to generate invoice", nil, "Cannot generate invoice for an empty order")
		case utils.ErrOrderCancelled:
			api.SendResponse(w, http.StatusBadRequest, "Failed to generate invoice", nil, "Cannot generate invoice for a cancelled or returned order")
		default:
			api.SendResponse(w, http.StatusInternalServerError, "Failed to generate invoice", nil, "An unexpected error occurred")
		}
		return
	}

	// Set headers for PDF download
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=invoice_%d.pdf", orderID))
	w.Header().Set("Content-Length", strconv.Itoa(len(pdfBytes)))

	// Write PDF bytes to response
	_, err = w.Write(pdfBytes)
	if err != nil {
		api.SendResponse(w, http.StatusInternalServerError, "Failed to send invoice", nil, "Error writing response")
		return
	}
}

func (h *OrderHandler) UpdateOrderDeliveryStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orderID, err := strconv.ParseInt(vars["orderId"], 10, 64)
	if err != nil {
		api.SendResponse(w, http.StatusBadRequest, "Failed to update delivery status", nil, "Invalid order ID")
		return
	}

	var input struct {
		DeliveryStatus string `json:"delivery_status"`
		OrderStatus    string `json:"order_status"`
		PaymentStatus  string `json:"payment_status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		api.SendResponse(w, http.StatusBadRequest, "Failed to update delivery status", nil, "Invalid request body")
		return
	}

	err = h.orderUseCase.UpdateOrderDeliveryStatus(r.Context(), orderID, input.DeliveryStatus, input.OrderStatus, input.PaymentStatus)
	if err != nil {
		switch err {
		case utils.ErrInvalidDeliveryStatus:
			api.SendResponse(w, http.StatusBadRequest, "Failed to update delivery status", nil, "Invalid delivery status")
		case utils.ErrInvalidOrderStatus:
			api.SendResponse(w, http.StatusBadRequest, "Failed to update delivery status", nil, "Invalid order status")
		case utils.ErrOrderNotFound:
			api.SendResponse(w, http.StatusNotFound, "Failed to update delivery status", nil, "Order not found")
		case utils.ErrOrderAlreadyDelivered:
			api.SendResponse(w, http.StatusBadRequest, "Failed to update delivery status", nil, "Order is already delivered")
		case utils.ErrMissingPaymentStatus:
			api.SendResponse(w, http.StatusBadRequest, "Failed to update delivery status", nil, "Payment status is required for COD orders")
		case utils.ErrInvalidPaymentStatus:
			api.SendResponse(w, http.StatusBadRequest, "Failed to update delivery status", nil, "Invalid payment status for COD orders")
		default:
			api.SendResponse(w, http.StatusInternalServerError, "Failed to update delivery status", nil, "An unexpected error occurred")
		}
		return
	}

	api.SendResponse(w, http.StatusOK, "Delivery status updated successfully", nil, "")
}

func (h *OrderHandler) PlaceOrderRazorpay(w http.ResponseWriter, r *http.Request) {
	// extract the user id from the context key
	userID, ok := r.Context().Value(middleware.UserIDKey).(int64)
	if !ok {
		api.SendResponse(w, http.StatusUnauthorized, "Failed to place order", nil, "User not authenticated")
		return
	}

	// Extract the checkout id from the url
	vars := mux.Vars(r)
	checkoutID, err := strconv.ParseInt(vars["checkout_id"], 10, 64)
	if err != nil {
		api.SendResponse(w, http.StatusBadRequest, "Failed to place order", nil, "Invalid checkout ID")
		return
	}

	// Call the method in the usecase layer
	order, err := h.orderUseCase.PlaceOrderRazorpay(r.Context(), userID, checkoutID)
	if err != nil {
		switch err {
		case utils.ErrCheckoutNotFound:
			api.SendResponse(w, http.StatusNotFound, "Failed to place order", nil, "Checkout not found")
		case utils.ErrUnauthorized:
			api.SendResponse(w, http.StatusForbidden, "Failed to place order", nil, "Unauthorized access to this checkout")
		case utils.ErrEmptyCheckout:
			api.SendResponse(w, http.StatusBadRequest, "Failed to place order", nil, "Cannot place order with empty checkout")
		case utils.ErrInsufficientStock:
			api.SendResponse(w, http.StatusBadRequest, "Failed to place order", nil, "Insufficient stock for one or more items")
		case utils.ErrInvalidAddress:
			api.SendResponse(w, http.StatusBadRequest, "Failed to place order", nil, "Invalid or missing delivery address")
		case utils.ErrOrderAlreadyPlaced:
			api.SendResponse(w, http.StatusConflict, "Failed to place order", nil, "Order has already been placed for this checkout")
		default:
			api.SendResponse(w, http.StatusInternalServerError, "Failed to place order", nil, "An unexpected error occurred")
		}
		return
	}

	api.SendResponse(w, http.StatusCreated, "Order placed successfully", order, "")
}

func (h *OrderHandler) CancelOrder(w http.ResponseWriter, r *http.Request) {
	// Extract the order id from the url
	vars := mux.Vars(r)
	orderID, err := strconv.ParseInt(vars["orderId"], 10, 64)
	if err != nil {
		api.SendResponse(w, http.StatusBadRequest, "Invalid order ID", nil, "Order ID must be a number")
		return
	}

	// Extract the user id from the context
	userID, ok := r.Context().Value(middleware.UserIDKey).(int64)
	if !ok {
		api.SendResponse(w, http.StatusUnauthorized, "Authentication required", nil, "User not authenticated")
		return
	}

	// Call the usecase method to initiate order cancellation from user side
	result, err := h.orderUseCase.CancelOrder(r.Context(), userID, orderID)
	if err != nil {
		switch err {
		case utils.ErrOrderNotFound:
			api.SendResponse(w, http.StatusNotFound, "Order not found", nil, "The specified order does not exist")
		case utils.ErrUnauthorized:
			api.SendResponse(w, http.StatusForbidden, "Unauthorized", nil, "You are not authorized to cancel this order")
		case utils.ErrOrderAlreadyCancelled:
			api.SendResponse(w, http.StatusBadRequest, "Order already cancelled", nil, "This order has already been cancelled")
		case utils.ErrOrderNotCancellable:
			api.SendResponse(w, http.StatusBadRequest, "Order not cancellable", nil, "This order cannot be cancelled in its current state")
		default:
			api.SendResponse(w, http.StatusInternalServerError, "Internal server error", nil, "An unexpected error occurred")
		}
		return
	}

	if result.RequiresAdminReview {
		api.SendResponse(w, http.StatusAccepted, "Cancellation request submitted for admin review", result, "")
	} else {
		api.SendResponse(w, http.StatusOK, "Order cancelled successfully", result, "")
	}
}

func (h *OrderHandler) AdminApproveCancellation(w http.ResponseWriter, r *http.Request) {
	// Extract order ID from URL
	vars := mux.Vars(r)
	orderID, err := strconv.ParseInt(vars["orderId"], 10, 64)
	if err != nil {
		api.SendResponse(w, http.StatusBadRequest, "Failed to approve cancellation", nil, "Invalid order ID format")
		return
	}

	// Call use case method to approve cancellation
	result, err := h.orderUseCase.ApproveCancellation(r.Context(), orderID)
	if err != nil {
		log.Printf("Error approving cancellation: %v", err)
		switch err {
		case utils.ErrOrderNotFound:
			api.SendResponse(w, http.StatusNotFound, "Failed to approve cancellation", nil, "Order not found")
		case utils.ErrOrderNotPendingCancellation:
			api.SendResponse(w, http.StatusBadRequest, "Failed to approve cancellation", nil, "Order is not in pending cancellation state")
		case utils.ErrCancellationRequestNotFound:
			api.SendResponse(w, http.StatusNotFound, "Failed to approve cancellation", nil, "Cancellation request not found")
		case utils.ErrRefundFailed:
			api.SendResponse(w, http.StatusInternalServerError, "Failed to process refund", map[string]interface{}{
				"order_id":   orderID,
				"new_status": "cancelled",
			}, "Failed to process refund")
		default:
			api.SendResponse(w, http.StatusInternalServerError, "Failed to approve cancellation", nil, "An unexpected error occurred")
		}
		return
	}

	// Prepare the response data
	responseData := map[string]interface{}{
		"order_id":             result.OrderID,
		"updated_order_status": result.UpdatedOrderStatus,
		"refund_status":        result.RefundStatus,
	}

	api.SendResponse(w, http.StatusOK, "Order cancellation approved successfully", responseData, "")
}

func (h *OrderHandler) AdminCancelOrder(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orderID, err := strconv.ParseInt(vars["orderId"], 10, 64)
	if err != nil {
		api.SendResponse(w, http.StatusBadRequest, "Failed to cancel order", nil, "Invalid order ID")
		return
	}

	result, err := h.orderUseCase.AdminCancelOrder(r.Context(), orderID)
	if err != nil {
		switch err {
		case utils.ErrOrderNotFound:
			api.SendResponse(w, http.StatusNotFound, "Failed to cancel order", nil, "Order not found")
		case utils.ErrOrderAlreadyCancelled:
			api.SendResponse(w, http.StatusBadRequest, "Failed to cancel order", nil, "Order is already cancelled")
		case utils.ErrOrderNotCancellable:
			api.SendResponse(w, http.StatusBadRequest, "Failed to cancel order", nil, "Order cannot be cancelled in its current state")
		default:
			log.Printf("Error cancelling order: %v", err)
			api.SendResponse(w, http.StatusInternalServerError, "Failed to cancel order", nil, "An unexpected error occurred")
		}
		return
	}

	responseData := map[string]interface{}{
		"order_id":              result.OrderID,
		"order_status":          result.OrderStatus,
		"refund_initiated":      result.RefundInitiated,
		"requires_admin_review": result.RequiresAdminReview,
	}

	api.SendResponse(w, http.StatusOK, "Order cancelled successfully", responseData, "")
}

func (h *OrderHandler) GetCancellationRequests(w http.ResponseWriter, r *http.Request) {
	// Set default values
	page := 1
	limit := 10

	// Parse page number from query parameter
	if pageStr := r.URL.Query().Get("page"); pageStr != "" { // If a string is provided with page query
		if parsedPage, err := strconv.Atoi(pageStr); err == nil && parsedPage > 0 {
			page = parsedPage
		}
	}

	// Create params struct
	params := domain.CancellationRequestParams{
		Page:  page,
		Limit: limit,
	}

	// Get cancellation requests
	requests, totalCount, err := h.orderUseCase.GetCancellationRequests(r.Context(), params)
	if err != nil {
		log.Printf("Error retrieving cancellation requests: %v", err)
		api.SendResponse(w, http.StatusInternalServerError, "Failed to retrieve cancellation requests", nil, "An unexpected error occurred")
		return
	}

	// Calculate total pages
	totalPages := (totalCount + int64(limit) - 1) / int64(limit)

	// Prepare response
	response := map[string]interface{}{
		"cancellation_requests": requests,
		"total_count":           totalCount,
		"page":                  page,
		"limit":                 limit,
		"total_pages":           totalPages,
	}

	api.SendResponse(w, http.StatusOK, "Cancellation requests retrieved successfully", response, "")
}
