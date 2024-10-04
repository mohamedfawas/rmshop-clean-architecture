package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/mohamedfawas/rmshop-clean-architecture/internal/delivery/http/middleware"
	"github.com/mohamedfawas/rmshop-clean-architecture/internal/usecase"
	"github.com/mohamedfawas/rmshop-clean-architecture/pkg/api"
	"github.com/mohamedfawas/rmshop-clean-architecture/pkg/utils"
)

type ReturnHandler struct {
	returnUseCase usecase.ReturnUseCase
}

func NewReturnHandler(returnUseCase usecase.ReturnUseCase) *ReturnHandler {
	return &ReturnHandler{returnUseCase: returnUseCase}
}

func (h *ReturnHandler) InitiateReturn(w http.ResponseWriter, r *http.Request) {
	// Extract the user id from the context
	userID, ok := r.Context().Value(middleware.UserIDKey).(int64)
	if !ok {
		api.SendResponse(w, http.StatusUnauthorized, "Failed to initiate return", nil, "User not authenticated")
		return
	}

	// extract the order id from the url
	vars := mux.Vars(r)
	orderID, err := strconv.ParseInt(vars["orderId"], 10, 64)
	if err != nil {
		api.SendResponse(w, http.StatusBadRequest, "Failed to initiate return", nil, "Invalid order ID")
		return
	}

	// input request body indicating reason for return
	var input struct {
		Reason string `json:"reason"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		api.SendResponse(w, http.StatusBadRequest, "Failed to initiate return", nil, "Invalid request body")
		return
	}

	// Call the usecase method to record initiate order return from user side
	returnRequest, err := h.returnUseCase.InitiateReturn(r.Context(), userID, orderID, input.Reason)
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

func (h *ReturnHandler) GetReturnRequestByOrderID(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(int64)
	if !ok {
		api.SendResponse(w, http.StatusUnauthorized, "Failed to get return request", nil, "User not authenticated")
		return
	}

	vars := mux.Vars(r)
	orderID, err := strconv.ParseInt(vars["orderId"], 10, 64)
	if err != nil {
		api.SendResponse(w, http.StatusBadRequest, "Failed to get return request", nil, "Invalid order ID")
		return
	}

	returnRequest, err := h.returnUseCase.GetReturnRequestByOrderID(r.Context(), userID, orderID)
	if err != nil {
		switch err {
		case utils.ErrOrderNotFound:
			api.SendResponse(w, http.StatusNotFound, "Failed to get return request", nil, "Order not found")
		case utils.ErrUnauthorized:
			api.SendResponse(w, http.StatusForbidden, "Failed to get return request", nil, "You don't have permission to access this order")
		case utils.ErrReturnRequestNotFound:
			api.SendResponse(w, http.StatusNotFound, "Failed to get return request", nil, "Return request not found for this order")
		default:
			api.SendResponse(w, http.StatusInternalServerError, "Failed to get return request", nil, "An unexpected error occurred")
		}
		return
	}

	api.SendResponse(w, http.StatusOK, "Return request retrieved successfully", returnRequest, "")
}

func (h *ReturnHandler) GetUserReturnRequests(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(int64)
	if !ok {
		api.SendResponse(w, http.StatusUnauthorized, "Failed to get return requests", nil, "User not authenticated")
		return
	}

	returnRequests, err := h.returnUseCase.GetUserReturnRequests(r.Context(), userID)
	if err != nil {
		api.SendResponse(w, http.StatusInternalServerError, "Failed to get return requests", nil, "An unexpected error occurred")
		return
	}

	api.SendResponse(w, http.StatusOK, "Return requests retrieved successfully", returnRequests, "")
}

func (h *ReturnHandler) UpdateReturnRequest(w http.ResponseWriter, r *http.Request) {
	// Extract the return id from the url
	vars := mux.Vars(r)
	returnID, err := strconv.ParseInt(vars["returnId"], 10, 64)
	if err != nil {
		api.SendResponse(w, http.StatusBadRequest, "Failed to update return request", nil, "Invalid return request ID")
		return
	}

	// fetch the request body
	var input struct {
		IsApproved bool `json:"is_approved"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		api.SendResponse(w, http.StatusBadRequest, "Failed to update return request", nil, "Invalid request body")
		return
	}

	// Call the usecase method
	updatedReturn, err := h.returnUseCase.UpdateReturnRequest(r.Context(), returnID, input.IsApproved)
	if err != nil {
		switch err {
		case utils.ErrReturnRequestNotFound:
			api.SendResponse(w, http.StatusNotFound, "Failed to update return request", nil, "Return request not found")
		case utils.ErrReturnRequestAlreadyProcessed:
			api.SendResponse(w, http.StatusBadRequest, "Failed to update return request", nil, "Return request has already been processed")
		default:
			api.SendResponse(w, http.StatusInternalServerError, "Failed to update return request", nil, "An unexpected error occurred")
		}
		return
	}

	message := "Return request approved successfully"
	if !input.IsApproved {
		message = "Return request rejected"
	}

	api.SendResponse(w, http.StatusOK, message, updatedReturn, "")
}

func (h *ReturnHandler) InitiateRefund(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	returnID, err := strconv.ParseInt(vars["returnId"], 10, 64)
	if err != nil {
		api.SendResponse(w, http.StatusBadRequest, "Failed to initiate refund", nil, "Invalid return request ID")
		return
	}

	refundDetails, err := h.returnUseCase.InitiateRefund(r.Context(), returnID)
	if err != nil {
		switch err {
		case utils.ErrReturnRequestNotFound:
			api.SendResponse(w, http.StatusNotFound, "Failed to initiate refund", nil, "Return request not found")
		case utils.ErrReturnRequestNotApproved:
			api.SendResponse(w, http.StatusBadRequest, "Failed to initiate refund", nil, "Return request not approved")
		case utils.ErrRefundAlreadyInitiated:
			api.SendResponse(w, http.StatusConflict, "Failed to initiate refund", nil, "Refund already initiated for this return request")
		case utils.ErrInsufficientBalance:
			api.SendResponse(w, http.StatusBadRequest, "Failed to initiate refund", nil, "Insufficient balance to process refund")
		case utils.ErrInvalidRefundAmount:
			api.SendResponse(w, http.StatusBadRequest, "Failed to initiate refund", nil, "Invalid refund amount")
		case utils.ErrOrderCancelled:
			api.SendResponse(w, http.StatusBadRequest, "Failed to initiate refund", nil, "Cannot refund a cancelled order")
		case utils.ErrNotEligibleForRefund:
			api.SendResponse(w, http.StatusBadRequest, "Failed to initiate refund", nil, "Given order not elgible for refund")
		default:
			api.SendResponse(w, http.StatusInternalServerError, "Failed to initiate refund", nil, "An unexpected error occurred")
		}
		return
	}

	api.SendResponse(w, http.StatusOK, "Refund initiated successfully", refundDetails, "")
}

func (h *ReturnHandler) MarkOrderReturnedToSeller(w http.ResponseWriter, r *http.Request) {
	// Extract return ID from URL
	vars := mux.Vars(r)
	returnID, err := strconv.ParseInt(vars["returnId"], 10, 64)
	if err != nil {
		api.SendResponse(w, http.StatusBadRequest, "Failed to mark order as returned", nil, "Invalid return ID")
		return
	}

	// Call the use case method
	returnRequest, err := h.returnUseCase.MarkOrderReturnedToSeller(r.Context(), returnID)
	if err != nil {
		switch err {
		case utils.ErrReturnRequestNotFound:
			api.SendResponse(w, http.StatusNotFound, "Failed to mark order as returned", nil, "Return request not found")
		case utils.ErrReturnRequestNotApproved:
			api.SendResponse(w, http.StatusBadRequest, "Failed to mark order as returned", nil, "Return request not approved")
		case utils.ErrAlreadyMarkedAsReturned:
			api.SendResponse(w, http.StatusBadRequest, "Failed to mark order as returned", nil, "Order already marked as returned to seller")
		case utils.ErrNoStockUpdated:
			api.SendResponse(w, http.StatusBadRequest, "Failed to mark order as returned", nil, "No stock updated, possibly due to empty order")
		default:
			api.SendResponse(w, http.StatusInternalServerError, "Failed to mark order as returned", nil, "An unexpected error occurred")
		}
		return
	}

	responseMessage := map[string]interface{}{
		"return_id":               returnRequest.ID,
		"order_id":                returnRequest.OrderID,
		"IsOrderReachedTheSeller": returnRequest.IsOrderReachedTheSeller,
		"IsStockUpdated":          returnRequest.IsStockUpdated,
	}

	// Send success response
	api.SendResponse(w, http.StatusOK, "Order marked as returned to seller successfully", responseMessage, "")
}

func (h *ReturnHandler) GetPendingReturnRequests(w http.ResponseWriter, r *http.Request) {
	pageStr := r.URL.Query().Get("page")
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	requests, totalCount, err := h.returnUseCase.GetPendingReturnRequests(r.Context(), page)
	if err != nil {
		api.SendResponse(w, http.StatusInternalServerError, "Failed to retrieve return requests", nil, "An unexpected error occurred")
		return
	}

	response := map[string]interface{}{
		"return_requests": requests,
		"total_count":     totalCount,
		"page":            page,
		"limit":           10,
		"total_pages":     (totalCount + 9) / 10, // Ceiling division by 10
	}

	api.SendResponse(w, http.StatusOK, "Return requests retrieved successfully", response, "")
}
