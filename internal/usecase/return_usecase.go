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
	UpdateReturnRequest(ctx context.Context, returnID int64, isApproved bool) (*domain.ReturnRequest, error)
	InitiateRefund(ctx context.Context, returnID int64) (*domain.RefundDetails, error)
	MarkOrderReturnedToSeller(ctx context.Context, returnID int64) error
}

type returnUseCase struct {
	returnRepo  repository.ReturnRepository
	orderRepo   repository.OrderRepository
	walletRepo  repository.WalletRepository
	productRepo repository.ProductRepository
}

func NewReturnUseCase(returnRepo repository.ReturnRepository,
	orderRepo repository.OrderRepository,
	walletRepo repository.WalletRepository,
	productRepo repository.ProductRepository) ReturnUseCase {
	return &returnUseCase{
		returnRepo:  returnRepo,
		orderRepo:   orderRepo,
		walletRepo:  walletRepo,
		productRepo: productRepo,
	}
}

/*
InitiateReturn:
- Get order using order id
- Validates whether the given order is eligible for return
- if eligible, create a return request
*/
func (u *returnUseCase) InitiateReturn(ctx context.Context, userID, orderID int64, reason string) (*domain.ReturnRequest, error) {
	// Get the order details from the orders table using order id
	order, err := u.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		log.Printf("failed to fetch orders details using order id : %v", err)
		return nil, err
	}

	// make sure the order selected belongs to the authenticated user
	if order.UserID != userID {
		return nil, utils.ErrUnauthorized
	}

	// Can't return the orders which don't have order status 'completed' (indicates order is delivered successfully to the user)
	if order.OrderStatus != utils.OrderStatusCompleted { // can't return orders which are not completed
		return nil, utils.ErrOrderNotEligibleForReturn
	}

	// if order is not already delivered, return not allowed
	if order.DeliveredAt == nil {
		return nil, utils.ErrOrderNotEligibleForReturn
	}
	// return window allowed is 14 days
	if time.Since(*order.DeliveredAt) > 14*24*time.Hour {
		return nil, utils.ErrReturnWindowExpired
	}

	// If already a return request is made for the selected order
	if order.HasReturnRequest {
		return nil, utils.ErrReturnAlreadyRequested
	}

	// no return reason provided
	if reason == "" {
		return nil, utils.ErrInvalidReturnReason
	}

	// Create a return request entry
	returnRequest := &domain.ReturnRequest{
		OrderID:                 orderID,
		UserID:                  userID,
		ReturnReason:            reason,
		IsApproved:              false,
		RequestedDate:           time.Now().UTC(),
		IsOrderReachedTheSeller: false,
		IsStockUpdated:          false,
	}

	// Create respective return request entry in the return_requests table
	err = u.returnRepo.CreateReturnRequest(ctx, returnRequest)
	if err != nil {
		log.Printf("error while adding return request in database : %v", err)
		return nil, err
	}

	// Update has_return_request column in orders table
	err = u.orderRepo.UpdateOrderHasReturnRequest(ctx, orderID, true)
	if err != nil {
		log.Printf("failed to update has_return_request of orders table in database : %v", err)
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

/*
UpdateReturnRequest:
- Get return request using return id
- Based on admin's approval
  - Update return requests table
  - If approved,
  - update order status in orders table
*/
func (u *returnUseCase) UpdateReturnRequest(ctx context.Context, returnID int64, isApproved bool) (*domain.ReturnRequest, error) {
	// Get return request by id
	returnRequest, err := u.returnRepo.GetByID(ctx, returnID)
	if err != nil {
		log.Printf("failed to fetch the return request using return id : %v", err)
		return nil, err
	}

	// If admin already reviwed this return request before, no need to review again
	if returnRequest.ApprovedAt != nil || returnRequest.RejectedAt != nil {
		return nil, utils.ErrReturnRequestAlreadyProcessed
	}

	now := time.Now().UTC()
	// Update based on admin's decision
	returnRequest.IsApproved = isApproved
	if isApproved {
		returnRequest.ApprovedAt = &now
	} else {
		returnRequest.RejectedAt = &now
	}

	err = u.returnRepo.UpdateApprovedOrRejected(ctx, returnRequest)
	if err != nil {
		log.Printf("failed to update admin's decision in return_requests table : %v", err)
		return nil, err
	}

	// Update the order's status if the return is approved
	if isApproved {
		err = u.orderRepo.UpdateOrderStatus(ctx, returnRequest.OrderID, utils.OrderStatusReturnApproved)
		if err != nil {
			log.Printf("Failed to update order status: %v", err)
			return nil, err
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

	// Calculate refund amount
	refundAmount := order.FinalAmount

	// Get current wallet balance
	wallet, err := u.walletRepo.GetByUserID(ctx, order.UserID)
	if err != nil {
		return nil, err
	}

	// Calculate new balance
	newBalance := wallet.Balance + refundAmount

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
		return nil, err
	}

	// Create wallet transaction
	walletTransaction := &domain.WalletTransaction{
		UserID:          order.UserID,
		Amount:          refundAmount,
		TransactionType: "REFUND",
		ReferenceID:     &returnID,
		ReferenceType:   utils.Ptr("RETURN"),
		BalanceAfter:    newBalance,
		CreatedAt:       time.Now().UTC(),
	}
	err = u.walletRepo.CreateTransaction(ctx, tx, walletTransaction)
	if err != nil {
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
		ReturnID:      returnID,
		OrderID:       order.ID,
		RefundAmount:  refundAmount,
		RefundStatus:  "completed",
		RefundedAt:    time.Now().UTC(),
		TransactionID: walletTransaction.ID,
	}

	return refundDetails, nil
}

func (u *returnUseCase) MarkOrderReturnedToSeller(ctx context.Context, returnID int64) error {
	// Start a transaction
	tx, err := u.returnRepo.BeginTx(ctx)
	if err != nil {
		log.Printf("failed to start transaction : %v", err)
		return err
	}
	defer tx.Rollback()

	// Get the return request using return id
	returnRequest, err := u.returnRepo.GetByID(ctx, returnID)
	if err != nil {
		log.Printf("failed to fetch the return request : %v", err)
		return err
	}

	// Check if the return is approved and not already marked as returned
	if !returnRequest.IsApproved {
		return utils.ErrReturnRequestNotApproved
	}
	if returnRequest.OrderReturnedToSellerAt != nil {
		return utils.ErrAlreadyMarkedAsReturned
	}

	// Update stock for returned products
	err = u.updateStockForReturnedOrder(ctx, tx, returnRequest.OrderID)
	if err != nil {
		log.Printf("failed to update stock quantity for the returned order items : %v", err)
		return err
	}

	// Update return request
	now := time.Now()
	returnRequest.OrderReturnedToSellerAt = &now
	returnRequest.IsOrderReachedTheSeller = true
	returnRequest.IsStockUpdated = true

	err = u.returnRepo.UpdateReturnRequest(ctx, returnRequest)
	if err != nil {
		log.Printf("failed to update the return request in the database : %v", err)
		return err
	}

	// commit the transaction
	err = tx.Commit()
	if err != nil {
		log.Printf("failed to commit the transaction : %v", err)
		return err
	}

	return nil
}

func (u *returnUseCase) updateStockForReturnedOrder(ctx context.Context, tx *sql.Tx, orderID int64) error {
	// Get order items
	orderItems, err := u.orderRepo.GetOrderItems(ctx, orderID)
	if err != nil {
		log.Printf("failed to get order items: %v", err)
		return err
	}

	// If there are no order items
	if len(orderItems) == 0 {
		return utils.ErrNoStockUpdated
	}

	// iterate through each order item
	for _, item := range orderItems {
		err = u.productRepo.UpdateStockTx(ctx, tx, item.ProductID, item.Quantity)
		if err != nil {
			log.Printf("failed to update stock for product %d: %v", item.ProductID, err)
			return err
		}
	}

	return nil
}
