package usecase

import (
	"context"
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
}

type returnUseCase struct {
	returnRepo repository.ReturnRepository
	orderRepo  repository.OrderRepository
}

func NewReturnUseCase(returnRepo repository.ReturnRepository, orderRepo repository.OrderRepository) ReturnUseCase {
	return &returnUseCase{
		returnRepo: returnRepo,
		orderRepo:  orderRepo,
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
