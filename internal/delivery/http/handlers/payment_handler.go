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
	"github.com/mohamedfawas/rmshop-clean-architecture/pkg/payment/razorpay"
	"github.com/mohamedfawas/rmshop-clean-architecture/pkg/utils"
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

	var userID int64
	if id, ok := r.Context().Value(middleware.UserIDKey).(int64); ok {
		userID = id
	}

	order, err := h.orderUseCase.GetOrderByID(r.Context(), userID, orderID)
	if err != nil {
		http.Error(w, "Failed to retrieve order details", http.StatusInternalServerError)
		return
	}

	// Check if payment is already completed
	if order.OrderStatus == utils.OrderStatusConfirmed || order.OrderStatus == utils.OrderStatusCompleted {
		data := struct {
			OrderID string
			Message string
		}{
			OrderID: orderIDStr,
			Message: "Payment for this order has already been completed.",
		}
		h.templates.ExecuteTemplate(w, "payment_completed.html", data)
		return
	}

	// Create Razorpay order
	razorpayOrder, err := h.razorpayService.CreateOrder(int64(order.FinalAmount*100), "INR")
	if err != nil {
		http.Error(w, "Failed to create Razorpay order", http.StatusInternalServerError)
		return
	}

	// Update the order with Razorpay order ID
	err = h.orderUseCase.UpdateOrderRazorpayID(r.Context(), orderID, razorpayOrder.ID)
	if err != nil {
		http.Error(w, "Failed to update order with Razorpay ID", http.StatusInternalServerError)
		return
	}

	data := struct {
		OrderID         string
		FinalPrice      float64
		RazorpayKeyID   string
		RazorpayOrderID string
	}{
		OrderID:         orderIDStr,
		FinalPrice:      order.FinalAmount,
		RazorpayKeyID:   h.razorpayKeyID,
		RazorpayOrderID: razorpayOrder.ID,
	}

	if err := h.templates.ExecuteTemplate(w, "payment.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *PaymentHandler) ProcessRazorpayPayment(w http.ResponseWriter, r *http.Request) {
	var input domain.RazorpayPaymentInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		log.Printf("Error decoding request body: %v", err)
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	log.Printf("Received Razorpay payment input: %+v", input)

	err := h.orderUseCase.VerifyAndUpdateRazorpayPayment(r.Context(), input)
	if err != nil {
		log.Printf("Error verifying and updating Razorpay payment: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("Payment processed successfully for order ID: %s", input.OrderID)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

func (h *PaymentHandler) RenderPaymentFailurePage(w http.ResponseWriter, r *http.Request) {
	orderID := r.URL.Query().Get("order_id")
	errorMsg := r.URL.Query().Get("error")
	if orderID == "" {
		http.Error(w, "Order ID is required", http.StatusBadRequest)
		return
	}

	data := struct {
		OrderID string
		Error   string
	}{
		OrderID: orderID,
		Error:   errorMsg,
	}

	err := h.templates.ExecuteTemplate(w, "payment_failure.html", data)
	if err != nil {
		log.Printf("Error executing template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func (h *PaymentHandler) RenderPaymentSuccessPage(w http.ResponseWriter, r *http.Request) {
	razorpayOrderID := r.URL.Query().Get("order_id")
	if razorpayOrderID == "" {
		http.Error(w, "Razorpay Order ID is required", http.StatusBadRequest)
		return
	}

	// First, get the payment details using the Razorpay order ID
	payment, err := h.orderUseCase.GetPaymentByRazorpayOrderID(r.Context(), razorpayOrderID)
	if err != nil {
		log.Printf("Error retrieving payment: %v", err)
		http.Error(w, "Failed to retrieve payment details", http.StatusInternalServerError)
		return
	}

	// Now use the order ID from the payment to get the order details
	order, err := h.orderUseCase.GetOrderByID(r.Context(), 0, payment.OrderID) // 0 for userID as it's not required here
	if err != nil {
		log.Printf("Error retrieving order: %v", err)
		http.Error(w, "Failed to retrieve order details", http.StatusInternalServerError)
		return
	}

	data := struct {
		OrderID         string
		Amount          float64
		RazorpayOrderID string
	}{
		OrderID:         strconv.FormatInt(order.ID, 10),
		Amount:          order.FinalAmount,
		RazorpayOrderID: razorpayOrderID,
	}

	err = h.templates.ExecuteTemplate(w, "payment_success.html", data)
	if err != nil {
		log.Printf("Error executing template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
