package usecase

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"time"

	"github.com/mohamedfawas/rmshop-clean-architecture/internal/domain"
	"github.com/mohamedfawas/rmshop-clean-architecture/internal/repository"
	"github.com/mohamedfawas/rmshop-clean-architecture/pkg/payment/razorpay"
	"github.com/mohamedfawas/rmshop-clean-architecture/pkg/utils"
)

type OrderUseCase interface {
	GetOrderByID(ctx context.Context, userID, orderID int64) (*domain.Order, error)
	GetUserOrders(ctx context.Context, userID int64, page, limit int, sortBy, order, status string) ([]*domain.Order, int64, error)
	CancelOrder(ctx context.Context, userID, orderID int64) (*domain.OrderCancellationResult, error)
	GetOrders(ctx context.Context, params domain.OrderQueryParams) ([]*domain.Order, int64, error)
	UpdateOrderStatus(ctx context.Context, orderID int64, newStatus string) (*domain.OrderStatusUpdateResult, error)
	GetPaymentByOrderID(ctx context.Context, orderID int64) (*domain.Payment, error)
	CreatePayment(ctx context.Context, payment *domain.Payment) error
	UpdatePayment(ctx context.Context, payment *domain.Payment) error
	ProcessPayment(ctx context.Context, orderID int64, paymentMethod string, amount float64) (*domain.Payment, error)
	VerifyAndUpdateRazorpayPayment(ctx context.Context, input domain.RazorpayPaymentInput) error
}

type orderUseCase struct {
	orderRepo       repository.OrderRepository
	razorpayService *razorpay.Service
}

func NewOrderUseCase(orderRepo repository.OrderRepository, razorpayKeyID, razorpaySecret string) OrderUseCase {
	return &orderUseCase{
		orderRepo:       orderRepo,
		razorpayService: razorpay.NewService(razorpayKeyID, razorpaySecret),
	}
}

func (u *orderUseCase) GetOrderByID(ctx context.Context, userID, orderID int64) (*domain.Order, error) {
	order, err := u.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		if err == utils.ErrOrderNotFound {
			return nil, utils.ErrOrderNotFound
		}
		return nil, err
	}

	// If userID is provided (not 0), check if the order belongs to the user
	if userID != 0 && order.UserID != userID {
		return nil, utils.ErrUnauthorized
	}

	return order, nil
}

func (u *orderUseCase) GetUserOrders(ctx context.Context, userID int64, page, limit int, sortBy, order, status string) ([]*domain.Order, int64, error) {
	// Validate pagination parameters
	if page < 1 || limit < 1 {
		return nil, 0, utils.ErrInvalidPaginationParams
	}

	// Validate and set default values for sorting
	if sortBy == "" {
		sortBy = "created_at"
	}
	if order == "" {
		order = "desc"
	}
	if order != "asc" && order != "desc" {
		order = "desc"
	}

	// Call repository method to get orders
	orders, totalCount, err := u.orderRepo.GetUserOrders(ctx, userID, page, limit, sortBy, order, status)
	if err != nil {
		return nil, 0, err
	}

	return orders, totalCount, nil
}

func (u *orderUseCase) CancelOrder(ctx context.Context, userID, orderID int64) (*domain.OrderCancellationResult, error) {
	// Get the order
	order, err := u.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		if err == utils.ErrOrderNotFound {
			return nil, utils.ErrOrderNotFound
		}
		return nil, err
	}

	// Check if the order belongs to the user
	if order.UserID != userID {
		return nil, utils.ErrUnauthorized
	}

	// Check if the order is already cancelled
	if order.OrderStatus == "cancelled" {
		return nil, utils.ErrOrderAlreadyCancelled
	}

	// Check if the order is in a cancellable state
	if !isCancellable(order.OrderStatus) {
		return nil, utils.ErrOrderNotCancellable
	}

	// Check if the cancellation period has expired
	if time.Since(order.CreatedAt) > 24*time.Hour {
		return nil, utils.ErrCancellationPeriodExpired
	}

	// Perform the cancellation
	err = u.orderRepo.UpdateOrderStatus(ctx, orderID, "cancelled")
	if err != nil {
		return nil, err
	}

	// Initiate refund if necessary
	refundStatus := "not_applicable"
	payment, err := u.orderRepo.GetPaymentByOrderID(ctx, orderID)
	if err != nil {
		// Log the error but continue, as we don't want to fail the cancellation due to payment retrieval issues
		log.Printf("Error retrieving payment for order %d: %v", orderID, err)
	} else if payment != nil && payment.Status == "paid" {
		refundStatus = "initiated"
		// Here you would typically call a payment service to initiate the refund
		// For now, we'll just update the status
		err = u.orderRepo.UpdateRefundStatus(ctx, orderID, sql.NullString{String: refundStatus, Valid: true})
		if err != nil {
			// Log the error, but don't fail the cancellation
			log.Printf("Error updating refund status for order %d: %v", orderID, err)
		}
	}

	return &domain.OrderCancellationResult{
		OrderID:      orderID,
		RefundStatus: refundStatus,
	}, nil
}

func (u *orderUseCase) GetOrders(ctx context.Context, params domain.OrderQueryParams) ([]*domain.Order, int64, error) {
	// Validate and set default values
	if params.Page < 1 {
		params.Page = 1
	}
	if params.Limit < 1 {
		params.Limit = 10
	} else if params.Limit > 100 {
		params.Limit = 100
	}

	validSortFields := map[string]bool{"created_at": true, "total_amount": true, "order_status": true}
	if params.SortBy != "" && !validSortFields[params.SortBy] {
		return nil, 0, errors.New("invalid sort field")
	}

	if params.SortOrder != "" && params.SortOrder != "asc" && params.SortOrder != "desc" {
		params.SortOrder = "desc"
	}

	// Call repository method
	return u.orderRepo.GetOrders(ctx, params)
}

func (u *orderUseCase) UpdateOrderStatus(ctx context.Context, orderID int64, newStatus string) (*domain.OrderStatusUpdateResult, error) {
	order, err := u.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		return nil, err
	}

	if !isValidOrderStatus(newStatus) {
		return nil, utils.ErrInvalidOrderStatus
	}

	var refundStatus string
	if newStatus == "cancelled" {
		if order.OrderStatus == "cancelled" {
			return nil, utils.ErrOrderAlreadyCancelled
		}
		if !isCancellable(order.OrderStatus) {
			return nil, utils.ErrOrderNotCancellable
		}
		refundStatus = "not_applicable"
		payment, err := u.orderRepo.GetPaymentByOrderID(ctx, orderID)
		if err != nil {
			log.Printf("Error retrieving payment for order %d: %v", orderID, err)
		} else if payment != nil && payment.Status == "paid" {
			refundStatus = "initiated"
			err = u.orderRepo.UpdateRefundStatus(ctx, orderID, sql.NullString{String: refundStatus, Valid: true})
			if err != nil {
				return nil, err
			}
		}
	}

	err = u.orderRepo.UpdateOrderStatus(ctx, orderID, newStatus)
	if err != nil {
		return nil, err
	}

	return &domain.OrderStatusUpdateResult{
		OrderID:      orderID,
		NewStatus:    newStatus,
		RefundStatus: refundStatus,
	}, nil
}

func isValidOrderStatus(status string) bool {
	validStatuses := []string{"pending", "processing", "shipped", "delivered", "cancelled"}
	for _, s := range validStatuses {
		if status == s {
			return true
		}
	}
	return false
}

func isCancellable(status string) bool {
	cancellableStatuses := map[string]bool{
		"pending":    true,
		"processing": true,
	}
	return cancellableStatuses[status]
}

func (u *orderUseCase) GetPaymentByOrderID(ctx context.Context, orderID int64) (*domain.Payment, error) {
	return u.orderRepo.GetPaymentByOrderID(ctx, orderID)
}

func (u *orderUseCase) CreatePayment(ctx context.Context, payment *domain.Payment) error {
	return u.orderRepo.CreatePayment(ctx, payment)
}

func (u *orderUseCase) UpdatePayment(ctx context.Context, payment *domain.Payment) error {
	return u.orderRepo.UpdatePayment(ctx, payment)
}

func (u *orderUseCase) ProcessPayment(ctx context.Context, orderID int64, paymentMethod string, amount float64) (*domain.Payment, error) {
	payment := &domain.Payment{
		OrderID:       orderID,
		Amount:        amount,
		PaymentMethod: paymentMethod,
		Status:        "pending",
	}

	if paymentMethod == "razorpay" {
		razorpayOrder, err := u.razorpayService.CreateOrder(int64(amount*100), "INR")
		if err != nil {
			return nil, err
		}
		payment.RazorpayOrderID = razorpayOrder.ID
	}

	err := u.orderRepo.CreatePayment(ctx, payment)
	if err != nil {
		return nil, err
	}

	return payment, nil
}

func (u *orderUseCase) VerifyAndUpdateRazorpayPayment(ctx context.Context, input domain.RazorpayPaymentInput) error {
	payment, err := u.orderRepo.GetPaymentByRazorpayOrderID(ctx, input.OrderID)
	if err != nil {
		return err
	}

	attributes := map[string]interface{}{
		"razorpay_order_id":   input.OrderID,
		"razorpay_payment_id": input.PaymentID,
		"razorpay_signature":  input.Signature,
	}

	if err := u.razorpayService.VerifyPaymentSignature(attributes); err != nil {
		return errors.New("invalid signature")
	}

	payment.Status = "paid"
	payment.RazorpayPaymentID = input.PaymentID
	payment.RazorpaySignature = input.Signature

	err = u.orderRepo.UpdatePayment(ctx, payment)
	if err != nil {
		return err
	}

	// Update order status
	err = u.orderRepo.UpdateOrderStatus(ctx, payment.OrderID, "paid")
	if err != nil {
		// Log the error but don't return it, as the payment was successful
		log.Printf("Failed to update order status: %v", err)
	}

	return nil
}
