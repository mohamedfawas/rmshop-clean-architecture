package usecase

import (
	"context"

	"github.com/mohamedfawas/rmshop-clean-architecture/internal/domain"
	"github.com/mohamedfawas/rmshop-clean-architecture/internal/repository"
	"github.com/mohamedfawas/rmshop-clean-architecture/pkg/utils"
)

type WalletUseCase interface {
	GetBalance(ctx context.Context, userID int64) (float64, error)
	GetWalletTransactions(ctx context.Context, userID int64, page, limit int, sort, order, transactionType string) ([]*domain.WalletTransaction, int64, error)
}

type walletUseCase struct {
	walletRepo repository.WalletRepository
	userRepo   repository.UserRepository
}

func NewWalletUseCase(walletRepo repository.WalletRepository, userRepo repository.UserRepository) WalletUseCase {
	return &walletUseCase{walletRepo: walletRepo, userRepo: userRepo}
}

func (u *walletUseCase) GetBalance(ctx context.Context, userID int64) (float64, error) {
	// Check if user exists
	_, err := u.userRepo.GetByID(ctx, userID)
	if err != nil {
		if err == utils.ErrUserNotFound {
			return 0, utils.ErrUserNotFound
		}
		return 0, err
	}

	// Get wallet balance
	wallet, err := u.walletRepo.GetByUserID(ctx, userID)
	if err != nil {
		if err == utils.ErrWalletNotFound {
			return 0, utils.ErrWalletNotInitialized
		}
		return 0, err
	}

	return wallet.Balance, nil
}

func (u *walletUseCase) GetWalletTransactions(ctx context.Context, userID int64, page, limit int, sort, order, transactionType string) ([]*domain.WalletTransaction, int64, error) {
	// Validate sort and order
	validSortFields := map[string]string{
		"date":   "created_at",
		"amount": "amount",
	}
	sortField, valid := validSortFields[sort]
	if !valid {
		sortField = "created_at"
	}
	if order != "asc" && order != "desc" {
		order = "desc"
	}

	// Call repository
	transactions, totalCount, err := u.walletRepo.GetTransactions(ctx, userID, page, limit, sortField, order, transactionType)
	if err != nil {
		return nil, 0, err
	}

	return transactions, totalCount, nil
}
