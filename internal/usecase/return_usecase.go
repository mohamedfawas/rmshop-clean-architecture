package usecase

import (
	"context"
	"database/sql"
	"log"
	"time"

	"github.com/mohamedfawas/rmshop-clean-architecture/internal/domain"
	"github.com/mohamedfawas/rmshop-clean-architecture/internal/repository"
	"github.com/mohamedfawas/rmshop-clean-architecture/pkg/utils"
)

type ReturnUseCase interface {
	InitiateReturn(ctx context.Context, userID, orderID int64, reason string) (*domain.ReturnRequest, error)
	GetReturnRequestByOrderID(ctx context.Context, userID, orderID int64) (*domain.ReturnRequest, error)
	GetUserReturnRequests(ctx context.Context, userID int64) ([]*domain.ReturnRequest, error)
	ApproveReturnRequest(ctx context.Context, returnID int64) error
	RejectReturnRequest(ctx context.Context, returnID int64) error
	UpdateReturnRequest(ctx context.Context, returnID int64, isApproved bool) (*domain.ReturnRequest, error)
	InitiateRefund(ctx context.Context, returnID int64) (*domain.RefundDetails, error)
}

type returnUseCase struct {
	returnRepo repository.ReturnRepository
	orderRepo  repository.OrderRepository
	walletRepo repository.WalletRepository
}

func NewReturnUseCase(returnRepo repository.ReturnRepository,
	orderRepo repository.OrderRepository,
	walletRepo repository.WalletRepository) ReturnUseCase {
	return &returnUseCase{
		returnRepo: returnRepo,
		orderRepo:  orderRepo,
		walletRepo: walletRepo,
	}
}

func (u *returnUseCase) InitiateReturn(ctx context.Context, userID, orderID int64, reason string) (*domain.ReturnRequest, error) {
	order, err := u.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		return nil, err
	}

	if order.UserID != userID {
		return nil, utils.ErrUnauthorized
	}

	if order.OrderStatus != utils.OrderStatusCompleted { // can't return orders which are not completed
		return nil, utils.ErrOrderNotEligibleForReturn
	}

	if order.DeliveredAt == nil {
		return nil, utils.ErrOrderNotEligibleForReturn
	}
	if time.Since(*order.DeliveredAt) > 14*24*time.Hour {
		return nil, utils.ErrReturnWindowExpired
	}

	if order.HasReturnRequest {
		return nil, utils.ErrReturnAlreadyRequested
	}

	if reason == "" {
		return nil, utils.ErrInvalidReturnReason
	}

	returnRequest := &domain.ReturnRequest{
		OrderID:       orderID,
		UserID:        userID,
		ReturnReason:  reason,
		IsApproved:    false,
		RequestedDate: time.Now().UTC(),
	}

	err = u.returnRepo.CreateReturnRequest(ctx, returnRequest)
	if err != nil {
		return nil, err
	}

	err = u.orderRepo.UpdateOrderHasReturnRequest(ctx, orderID, true)
	if err != nil {
		return nil, err
	}

	return returnRequest, nil
}

func (u *returnUseCase) GetReturnRequestByOrderID(ctx context.Context, userID, orderID int64) (*domain.ReturnRequest, error) {
	order, err := u.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		return nil, err
	}

	if order.UserID != userID {
		return nil, utils.ErrUnauthorized
	}

	return u.returnRepo.GetReturnRequestByOrderID(ctx, orderID)
}

func (u *returnUseCase) GetUserReturnRequests(ctx context.Context, userID int64) ([]*domain.ReturnRequest, error) {
	return u.returnRepo.GetUserReturnRequests(ctx, userID)
}

func (u *returnUseCase) ApproveReturnRequest(ctx context.Context, returnID int64) error {
	return u.returnRepo.UpdateReturnRequestStatus(ctx, returnID, true)
}

func (u *returnUseCase) RejectReturnRequest(ctx context.Context, returnID int64) error {
	return u.returnRepo.UpdateReturnRequestStatus(ctx, returnID, false)
}

func (u *returnUseCase) UpdateReturnRequest(ctx context.Context, returnID int64, isApproved bool) (*domain.ReturnRequest, error) {
	returnRequest, err := u.returnRepo.GetByID(ctx, returnID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, utils.ErrReturnRequestNotFound
		}
		return nil, err
	}

	if returnRequest.ApprovedAt != nil || returnRequest.RejectedAt != nil {
		return nil, utils.ErrReturnRequestAlreadyProcessed
	}

	now := time.Now().UTC()
	returnRequest.IsApproved = isApproved
	if isApproved {
		returnRequest.ApprovedAt = &now
	} else {
		returnRequest.RejectedAt = &now
	}

	err = u.returnRepo.UpdateApprovedOrRejected(ctx, returnRequest)
	if err != nil {
		return nil, err
	}

	// Update the order's status if the return is approved
	if isApproved {
		err = u.orderRepo.UpdateOrderStatus(ctx, returnRequest.OrderID, "return_approved")
		if err != nil {
			// Log this error, but don't fail the return request update
			log.Printf("Failed to update order status: %v", err)
		}
	}

	return returnRequest, nil
}

func (u *returnUseCase) InitiateRefund(ctx context.Context, returnID int64) (*domain.RefundDetails, error) {
	// Start a database transaction
	tx, err := u.orderRepo.BeginTx(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// Get the return request
	returnRequest, err := u.returnRepo.GetByID(ctx, returnID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, utils.ErrReturnRequestNotFound
		}
		return nil, err
	}

	// Check if the return is approved
	if !returnRequest.IsApproved {
		return nil, utils.ErrReturnRequestNotApproved
	}

	// Check if refund is already initiated
	if returnRequest.RefundInitiated {
		return nil, utils.ErrRefundAlreadyInitiated
	}

	// Get the order details
	order, err := u.orderRepo.GetByID(ctx, returnRequest.OrderID)
	if err != nil {
		return nil, err
	}

	// Check if the order was cancelled
	if order.OrderStatus == "cancelled" {
		return nil, utils.ErrOrderCancelled
	}

	// Calculate refund amount (you might want to implement a more sophisticated calculation)
	refundAmount := order.FinalAmount

	// Update return request
	returnRequest.RefundInitiated = true
	returnRequest.RefundAmount = &refundAmount
	err = u.returnRepo.UpdateRefundDetails(ctx, returnRequest)
	if err != nil {
		return nil, err
	}

	// Add amount to user's wallet
	err = u.walletRepo.AddBalance(ctx, tx, order.UserID, refundAmount)
	if err != nil {
		if err == utils.ErrInsufficientBalance {
			return nil, utils.ErrInsufficientBalance
		}
		return nil, err
	}

	// Update order status
	err = u.orderRepo.UpdateOrderStatusTx(ctx, tx, order.ID, "refunded")
	if err != nil {
		return nil, err
	}

	// Commit the transaction
	if err = tx.Commit(); err != nil {
		return nil, err
	}

	refundDetails := &domain.RefundDetails{
		ReturnID:     returnID,
		OrderID:      order.ID,
		RefundAmount: refundAmount,
		RefundStatus: "completed",
		RefundedAt:   time.Now().UTC(),
	}

	return refundDetails, nil
}
