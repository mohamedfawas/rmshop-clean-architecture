package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

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

/*
GetByUserID:
- Get wallet details using user id
*/
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
		log.Printf("failed to retrieve wallet details using user id : %v", err)
		return nil, err
	}
	return wallet, nil
}

/*
AddBalance:
- Insert refund details in wallets table
- Mainly the balance is updated after adding the refund amount
*/
func (r *walletRepository) AddBalance(ctx context.Context, tx *sql.Tx, userID int64, amount float64) error {
	/*
		query explanation:
		- If a row with same user id exists then update with the given details
	*/
	query := `
        INSERT INTO wallets (user_id, balance, created_at, updated_at)
        VALUES ($1, $2, NOW(), NOW())
        ON CONFLICT (user_id) 
        DO UPDATE SET 
            balance = wallets.balance + $2,
            updated_at = NOW()
        RETURNING balance
    `
	var newBalance float64
	err := tx.QueryRowContext(ctx, query, userID, amount).Scan(&newBalance)
	if err != nil {
		log.Printf("Error updating wallet balance: %v", err)
		return err
	}
	log.Printf("Updated wallet balance for user %d: %f", userID, newBalance)
	return nil
}

/*
CreateTransaction:
- add transaction entry in wallet_transactions table
*/
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
		time.Now().UTC(),
	).Scan(&transaction.ID)
	if err != nil {
		log.Printf("Error creating wallet transaction: %v", err)
		return fmt.Errorf("failed to create wallet transaction: %w", err)
	}
	log.Printf("Created wallet transaction with ID %d for user %d", transaction.ID, transaction.UserID)
	return nil
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

func (r *walletRepository) CreateTransactionTx(ctx context.Context, tx *sql.Tx, transaction *domain.WalletTransaction) error {
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

	if err != nil {
		return fmt.Errorf("failed to create wallet transaction: %w", err)
	}
	return nil
}

func (r *walletRepository) UpdateBalanceTx(ctx context.Context, tx *sql.Tx, userID int64, newBalance float64) error {
	query := `
        INSERT INTO wallets (user_id, balance, updated_at)
        VALUES ($1, $2, NOW())
        ON CONFLICT (user_id) DO UPDATE
        SET balance = $2, updated_at = NOW()
    `
	_, err := tx.ExecContext(ctx, query, userID, newBalance)
	if err != nil {
		return fmt.Errorf("failed to update wallet balance: %w", err)
	}
	return nil
}

/*
GetWalletTx:
- get wallet details for the given user id from the wallets table
*/
func (r *walletRepository) GetWalletTx(ctx context.Context, tx *sql.Tx, userID int64) (*domain.Wallet, error) {
	query := `SELECT id, user_id, balance, created_at, updated_at FROM wallets WHERE user_id = $1`
	var wallet domain.Wallet
	err := tx.QueryRowContext(ctx, query, userID).
		Scan(&wallet.ID,
			&wallet.UserID,
			&wallet.Balance,
			&wallet.CreatedAt,
			&wallet.UpdatedAt)
	if err != nil {
		log.Printf("error while fetching the wallet details for the given user id : %v", err)
		return nil, err
	}
	return &wallet, nil
}

/*
CreateWalletTx:
- Create a wallet entry in wallets table
*/
func (r *walletRepository) CreateWalletTx(ctx context.Context, tx *sql.Tx, wallet *domain.Wallet) error {
	query := `
        INSERT INTO wallets (user_id, balance, created_at, updated_at)
        VALUES ($1, $2, $3, $4)
        RETURNING id
    `
	err := tx.QueryRowContext(ctx, query,
		wallet.UserID,
		wallet.Balance,
		wallet.CreatedAt,
		wallet.UpdatedAt).Scan(&wallet.ID)
	if err != nil {
		log.Printf("error while creating entry in wallets table : %v", err)
		return err
	}
	return nil
}

/*
UpdateWalletBalanceTx:
- Used to update wallet balance in wallets table for the given user id
*/
func (r *walletRepository) UpdateWalletBalanceTx(ctx context.Context, tx *sql.Tx, userID int64, newBalance float64) error {
	query := `UPDATE wallets SET balance = $1, updated_at = $2 WHERE user_id = $3`
	_, err := tx.ExecContext(ctx,
		query,
		newBalance,
		time.Now().UTC(),
		userID)
	if err != nil {
		log.Printf("error while updating wallet balance in wallets table : %v", err)
	}
	return err
}

/*
CreateWalletTransactionTx:
- create wallet transaction entry in  wallet_transactions table
*/
func (r *walletRepository) CreateWalletTransactionTx(ctx context.Context, tx *sql.Tx, transaction *domain.WalletTransaction) error {
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
	if err != nil {
		log.Printf("error while creating wallet transaction entry in wallet_transactions table : %v", err)
	}
	return err
}
