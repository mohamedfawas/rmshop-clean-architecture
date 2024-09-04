package usecase

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/mohamedfawas/rmshop-clean-architecture/internal/domain"
	"github.com/mohamedfawas/rmshop-clean-architecture/internal/repository"
	"github.com/mohamedfawas/rmshop-clean-architecture/pkg/utils"
)

type OrderUseCase interface {
	GetOrderByID(ctx context.Context, userID, orderID int64) (*domain.Order, error)
	GetUserOrders(ctx context.Context, userID int64, page, limit int, sortBy, order, status string) ([]*domain.Order, int64, error)
	CancelOrder(ctx context.Context, userID, orderID int64) (*domain.OrderCancellationResult, error)
	GetOrders(ctx context.Context, params domain.OrderQueryParams) ([]*domain.Order, int64, error)
	UpdateOrderStatus(ctx context.Context, orderID int64, newStatus string) (*domain.OrderStatusUpdateResult, error)
}

type orderUseCase struct {
	orderRepo repository.OrderRepository
}

func NewOrderUseCase(orderRepo repository.OrderRepository) OrderUseCase {
	return &orderUseCase{orderRepo: orderRepo}
}

func (u *orderUseCase) GetOrderByID(ctx context.Context, userID, orderID int64) (*domain.Order, error) {
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
	if order.PaymentStatus == "paid" {
		refundStatus = "initiated"
		// Here you would typically call a payment service to initiate the refund
		// For now, we'll just update the status
		err = u.orderRepo.UpdateRefundStatus(ctx, orderID, sql.NullString{String: refundStatus, Valid: true})
		if err != nil {
			// Log the error, but don't fail the cancellation
			// You might want to handle this differently based on your requirements
		}
	}

	return &domain.OrderCancellationResult{
		OrderID:      orderID,
		RefundStatus: refundStatus,
	}, nil
}

func isCancellable(status string) bool {
	cancellableStatuses := map[string]bool{
		"pending":    true,
		"processing": true,
	}
	return cancellableStatuses[status]
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

	// Get the current order
	order, err := u.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		if err == utils.ErrOrderNotFound {
			return nil, utils.ErrOrderNotFound
		}
		return nil, err
	}

	// Validate the new status
	if !isValidOrderStatus(newStatus) {
		return nil, utils.ErrInvalidOrderStatus
	}

	// If status is being set to cancelled, perform cancellation logic
	var refundStatus string
	if newStatus == "cancelled" {
		if order.OrderStatus == "cancelled" {
			return nil, utils.ErrOrderAlreadyCancelled
		}
		if !isCancellable(order.OrderStatus) {
			return nil, utils.ErrOrderNotCancellable
		}
		// Perform cancellation-specific logic (e.g., initiate refund)
		refundStatus = "not_applicable"
		if order.PaymentStatus == "paid" {
			refundStatus = "initiated"
			err = u.orderRepo.UpdateRefundStatus(ctx, orderID, sql.NullString{String: refundStatus, Valid: true})
			if err != nil {
				// Handle refund initiation error
				return nil, err
			}
		}
	}

	// Update the order status
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
