package postgres

import (
	"context"
	"database/sql"
	"time"

	"github.com/mohamedfawas/rmshop-clean-architecture/internal/domain"
	"github.com/mohamedfawas/rmshop-clean-architecture/pkg/utils"
)

type paymentRepository struct {
	db *sql.DB
}

func NewPaymentRepository(db *sql.DB) *paymentRepository {
	return &paymentRepository{db: db}
}

func (r *paymentRepository) GetByOrderID(ctx context.Context, orderID int64) (*domain.Payment, error) {
	query := `
        SELECT id, order_id, amount, payment_method, status, created_at, updated_at, 
               razorpay_order_id, razorpay_payment_id, razorpay_signature
        FROM payments
        WHERE order_id = $1
    `
	var payment domain.Payment
	err := r.db.QueryRowContext(ctx, query, orderID).Scan(
		&payment.ID, &payment.OrderID, &payment.Amount, &payment.PaymentMethod, &payment.Status,
		&payment.CreatedAt, &payment.UpdatedAt, &payment.RazorpayOrderID, &payment.RazorpayPaymentID,
		&payment.RazorpaySignature,
	)

	if err == sql.ErrNoRows {
		return nil, utils.ErrPaymentNotFound
	}
	if err != nil {
		return nil, err
	}

	return &payment, nil
}

func (r *paymentRepository) InitiateRefund(ctx context.Context, paymentID int64) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Get payment details
	var payment domain.Payment
	err = tx.QueryRowContext(ctx, "SELECT order_id, amount, status FROM payments WHERE id = $1", paymentID).Scan(
		&payment.OrderID, &payment.Amount, &payment.Status,
	)
	if err != nil {
		return err
	}

	if payment.Status != "paid" {
		return utils.ErrPaymentNotRefundable
	}

	// Get user ID from order
	var userID int64
	err = tx.QueryRowContext(ctx, "SELECT user_id FROM orders WHERE id = $1", payment.OrderID).Scan(&userID)
	if err != nil {
		return err
	}

	// Add refund amount to user's wallet
	_, err = tx.ExecContext(ctx, `
        INSERT INTO wallets (user_id, balance)
        VALUES ($1, $2)
        ON CONFLICT (user_id) DO UPDATE
        SET balance = wallets.balance + $2
    `, userID, payment.Amount)
	if err != nil {
		return err
	}

	// Create wallet transaction record
	_, err = tx.ExecContext(ctx, `
        INSERT INTO wallet_transactions (user_id, amount, transaction_type, reference_id, created_at)
        VALUES ($1, $2, 'REFUND', $3, $4)
    `, userID, payment.Amount, payment.OrderID, time.Now())
	if err != nil {
		return err
	}

	// Update payment status
	_, err = tx.ExecContext(ctx, "UPDATE payments SET status = 'refunded', updated_at = $1 WHERE id = $2", time.Now(), paymentID)
	if err != nil {
		return err
	}

	return tx.Commit()
}
