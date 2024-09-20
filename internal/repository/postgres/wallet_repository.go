package postgres

import (
	"context"
	"database/sql"

	"github.com/mohamedfawas/rmshop-clean-architecture/internal/domain"
	"github.com/mohamedfawas/rmshop-clean-architecture/internal/repository"
	"github.com/mohamedfawas/rmshop-clean-architecture/pkg/utils"
)

type walletRepository struct {
	db *sql.DB
}

func NewWalletRepository(db *sql.DB) repository.WalletRepository {
	return &walletRepository{db: db}
}

func (r *walletRepository) GetByUserID(ctx context.Context, userID int64) (*domain.Wallet, error) {
	query := `
		SELECT id, user_id, balance, created_at, updated_at
		FROM wallets
		WHERE user_id = $1
	`
	wallet := &domain.Wallet{}
	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&wallet.ID,
		&wallet.UserID,
		&wallet.Balance,
		&wallet.CreatedAt,
		&wallet.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, utils.ErrWalletNotFound
		}
		return nil, err
	}
	return wallet, nil
}

func (r *walletRepository) AddBalance(ctx context.Context, tx *sql.Tx, userID int64, amount float64) error {
	query := `
        INSERT INTO wallets (user_id, balance)
        VALUES ($1, $2)
        ON CONFLICT (user_id)
        DO UPDATE SET balance = wallets.balance + $2
    `
	_, err := tx.ExecContext(ctx, query, userID, amount)
	return err
}
