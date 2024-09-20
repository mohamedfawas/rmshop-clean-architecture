package postgres

import (
	"context"
	"database/sql"
	"fmt"

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

func (r *walletRepository) CreateTransaction(ctx context.Context, tx *sql.Tx, transaction *domain.WalletTransaction) error {
	query := `
        INSERT INTO wallet_transactions (user_id, amount, transaction_type, reference_id, reference_type, balance_after, created_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7)
        RETURNING id
    `
	err := tx.QueryRowContext(ctx, query,
		transaction.UserID,
		transaction.Amount,
		transaction.TransactionType,
		transaction.ReferenceID,
		transaction.ReferenceType,
		transaction.BalanceAfter,
		transaction.CreatedAt,
	).Scan(&transaction.ID)
	return err
}

func (r *walletRepository) UpdateBalance(ctx context.Context, tx *sql.Tx, userID int64, newBalance float64) error {
	query := `
		UPDATE wallets
		SET balance = $1, updated_at = NOW()
		WHERE user_id = $2
	`
	_, err := tx.ExecContext(ctx, query, newBalance, userID)
	return err
}

func (r *walletRepository) GetTransactions(ctx context.Context, userID int64, page, limit int, sort, order, transactionType string) ([]*domain.WalletTransaction, int64, error) {
	query := `
        SELECT id, user_id, amount, transaction_type, reference_id, reference_type, balance_after, created_at
        FROM wallet_transactions
        WHERE user_id = $1
    `
	countQuery := `SELECT COUNT(*) FROM wallet_transactions WHERE user_id = $1`
	args := []interface{}{userID}

	// Add type filter if provided
	if transactionType != "" {
		query += ` AND transaction_type = $2`
		countQuery += ` AND transaction_type = $2`
		args = append(args, transactionType)
	}

	// Add sorting
	query += fmt.Sprintf(` ORDER BY %s %s`, sort, order)

	// Add pagination
	query += fmt.Sprintf(` LIMIT $%d OFFSET $%d`, len(args)+1, len(args)+2)
	args = append(args, limit, (page-1)*limit)

	// Get total count
	var totalCount int64
	err := r.db.QueryRowContext(ctx, countQuery, args[:len(args)-2]...).Scan(&totalCount)
	if err != nil {
		return nil, 0, err
	}

	// Execute main query
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var transactions []*domain.WalletTransaction
	for rows.Next() {
		var t domain.WalletTransaction
		err := rows.Scan(&t.ID, &t.UserID, &t.Amount, &t.TransactionType, &t.ReferenceID, &t.ReferenceType, &t.BalanceAfter, &t.CreatedAt)
		if err != nil {
			return nil, 0, err
		}
		transactions = append(transactions, &t)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, err
	}

	return transactions, totalCount, nil
}
