package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"text/template"

	"github.com/mohamedfawas/rmshop-clean-architecture/internal/delivery/http/middleware"
	"github.com/mohamedfawas/rmshop-clean-architecture/internal/domain"
	"github.com/mohamedfawas/rmshop-clean-architecture/internal/usecase"
	"github.com/mohamedfawas/rmshop-clean-architecture/pkg/api"
	"github.com/mohamedfawas/rmshop-clean-architecture/pkg/payment/razorpay"
)

type PaymentHandler struct {
	orderUseCase    usecase.OrderUseCase
	razorpayKeyID   string
	razorpaySecret  string
	templates       *template.Template
	razorpayService *razorpay.Service
}

func NewPaymentHandler(orderUseCase usecase.OrderUseCase, razorpayKeyID, razorpaySecret string, templates *template.Template) *PaymentHandler {
	return &PaymentHandler{
		orderUseCase:    orderUseCase,
		razorpayKeyID:   razorpayKeyID,
		razorpaySecret:  razorpaySecret,
		templates:       templates,
		razorpayService: razorpay.NewService(razorpayKeyID, razorpaySecret),
	}
}

func (h *PaymentHandler) RenderPaymentPage(w http.ResponseWriter, r *http.Request) {
	// Extract order ID from the request
	orderIDStr := r.URL.Query().Get("order_id")
	if orderIDStr == "" {
		http.Error(w, "Order ID is required", http.StatusBadRequest)
		return
	}

	orderID, err := strconv.ParseInt(orderIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid Order ID", http.StatusBadRequest)
		return
	}

	// Try to extract user ID from the context, but don't require it
	var userID int64
	if id, ok := r.Context().Value(middleware.UserIDKey).(int64); ok {
		userID = id
	}

	// Get order details
	order, err := h.orderUseCase.GetOrderByID(r.Context(), userID, orderID)
	if err != nil {
		http.Error(w, "Failed to retrieve order details", http.StatusInternalServerError)
		return
	}

	// Create Razorpay order
	razorpayOrder, err := h.razorpayService.CreateOrder(int64(order.TotalAmount*100), "INR")
	if err != nil {
		http.Error(w, "Failed to create Razorpay order", http.StatusInternalServerError)
		return
	}

	// Prepare data for the template
	data := struct {
		OrderID         string
		FinalPrice      float64
		RazorpayKeyID   string
		RazorpayOrderID string
	}{
		OrderID:         strconv.FormatInt(order.ID, 10), // Use the actual order ID
		FinalPrice:      order.TotalAmount,
		RazorpayKeyID:   h.razorpayKeyID,
		RazorpayOrderID: razorpayOrder.ID,
	}

	// Log the Razorpay Key ID for debugging
	log.Printf("Using Razorpay Key ID: %s", h.razorpayKeyID)

	// Render the template
	if err := h.templates.ExecuteTemplate(w, "payment.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *PaymentHandler) ProcessRazorpayPayment(w http.ResponseWriter, r *http.Request) {
	var input domain.RazorpayPaymentInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		api.SendResponse(w, http.StatusBadRequest, "Invalid request body", nil, err.Error())
		return
	}

	err := h.orderUseCase.VerifyAndUpdateRazorpayPayment(r.Context(), input)
	if err != nil {
		api.SendResponse(w, http.StatusInternalServerError, "Payment verification failed", nil, err.Error())
		return
	}

	api.SendResponse(w, http.StatusOK, "Payment processed successfully", nil, "")
}
