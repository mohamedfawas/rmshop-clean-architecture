package postgres

import (
	"context"
	"database/sql"
	"log"
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

/*
GetByOrderID:
- Get payment details from payments table using order id
*/
func (r *paymentRepository) GetByOrderID(ctx context.Context, orderID int64) (*domain.Payment, error) {
	query := `
        SELECT id, order_id, amount, payment_method, payment_status, created_at, updated_at, 
               razorpay_order_id, razorpay_payment_id, razorpay_signature
        FROM payments
        WHERE order_id = $1
    `
	var payment domain.Payment
	var rzpPaymentID, rzpSignature sql.NullString
	err := r.db.QueryRowContext(ctx, query, orderID).Scan(
		&payment.ID,
		&payment.OrderID,
		&payment.Amount,
		&payment.PaymentMethod,
		&payment.Status,
		&payment.CreatedAt,
		&payment.UpdatedAt,
		&payment.RazorpayOrderID,
		&rzpPaymentID,
		&rzpSignature,
	)

	if rzpPaymentID.Valid {
		payment.RazorpayPaymentID = rzpPaymentID.String
	}
	if rzpSignature.Valid {
		payment.RazorpaySignature = rzpSignature.String
	}

	if err == sql.ErrNoRows {
		return nil, utils.ErrPaymentNotFound
	}
	if err != nil {
		return nil, err
	}
	return &payment, nil
}

/*
UpdatePayment:
- update payment details in payments table
- payment_status, razorpay_payment_id, razorpay_signature, updated_at
-
*/
func (r *paymentRepository) UpdatePayment(ctx context.Context, payment *domain.Payment) error {
	query := `
        UPDATE payments
        SET payment_status = $1, razorpay_payment_id = $2, razorpay_signature = $3, updated_at = $4
        WHERE id = $5
    `
	_, err := r.db.ExecContext(ctx, query,
		payment.Status,
		payment.RazorpayPaymentID,
		payment.RazorpaySignature,
		payment.UpdatedAt,
		payment.ID)

	return err
}

func (r *paymentRepository) GetByRazorpayOrderID(ctx context.Context, razorpayOrderID string) (*domain.Payment, error) {
	query := `SELECT id, order_id, amount, payment_method, payment_status, created_at, updated_at, 
              razorpay_order_id, razorpay_payment_id, razorpay_signature
              FROM payments WHERE razorpay_order_id = $1`

	var payment domain.Payment
	var rzpPaymentID, rzpSignature sql.NullString

	err := r.db.QueryRowContext(ctx, query, razorpayOrderID).Scan(
		&payment.ID,
		&payment.OrderID,
		&payment.Amount,
		&payment.PaymentMethod,
		&payment.Status,
		&payment.CreatedAt,
		&payment.UpdatedAt,
		&payment.RazorpayOrderID,
		&rzpPaymentID,
		&rzpSignature,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, utils.ErrPaymentNotFound
		}
		return nil, err
	}

	if rzpPaymentID.Valid {
		payment.RazorpayPaymentID = rzpPaymentID.String
	}
	if rzpSignature.Valid {
		payment.RazorpaySignature = rzpSignature.String
	}

	return &payment, nil
}

/*
GetByOrderIDTx:
- Get payment details from payments table using order id
*/
func (r *paymentRepository) GetByOrderIDTx(ctx context.Context, tx *sql.Tx, orderID int64) (*domain.Payment, error) {
	query := `
        SELECT id, order_id, amount, payment_method, payment_status, created_at, updated_at, 
               razorpay_order_id, razorpay_payment_id, razorpay_signature
        FROM payments
        WHERE order_id = $1
    `
	var payment domain.Payment
	var razorpayPaymentID, razorpaySignature sql.NullString

	err := tx.QueryRowContext(ctx, query, orderID).Scan(
		&payment.ID, &payment.OrderID, &payment.Amount, &payment.PaymentMethod, &payment.Status,
		&payment.CreatedAt, &payment.UpdatedAt, &payment.RazorpayOrderID,
		&razorpayPaymentID, &razorpaySignature,
	)

	if err == sql.ErrNoRows {
		return nil, utils.ErrPaymentNotFound
	}
	if err != nil {
		log.Printf("failed to get payment: %v", err)
		return nil, err
	}

	if razorpayPaymentID.Valid {
		payment.RazorpayPaymentID = razorpayPaymentID.String
	}
	if razorpaySignature.Valid {
		payment.RazorpaySignature = razorpaySignature.String
	}

	return &payment, nil
}

/*
UpdateStatusTx:
- Update payment_status for the given payment id
*/
func (r *paymentRepository) UpdateStatusTx(ctx context.Context, tx *sql.Tx, paymentID int64, status string) error {
	query := `
        UPDATE payments
        SET payment_status = $1, updated_at = NOW()
        WHERE id = $2
    `
	_, err := tx.ExecContext(ctx, query, status, paymentID)
	if err != nil {
		log.Printf("failed to update payment status: %v", err)
		return err
	}
	return nil
}
