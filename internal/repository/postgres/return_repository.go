package postgres

import (
	"context"
	"database/sql"
	"time"

	"github.com/mohamedfawas/rmshop-clean-architecture/internal/domain"
	"github.com/mohamedfawas/rmshop-clean-architecture/pkg/utils"
)

type returnRepository struct {
	db *sql.DB
}

func NewReturnRepository(db *sql.DB) *returnRepository {
	return &returnRepository{db: db}
}

func (r *returnRepository) CreateReturnRequest(ctx context.Context, returnRequest *domain.ReturnRequest) error {
	query := `
		INSERT INTO return_requests (order_id, user_id, return_reason, is_approved, requested_date)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`
	err := r.db.QueryRowContext(ctx, query,
		returnRequest.OrderID,
		returnRequest.UserID,
		returnRequest.ReturnReason,
		returnRequest.IsApproved,
		returnRequest.RequestedDate,
	).Scan(&returnRequest.ID)
	return err
}

func (r *returnRepository) GetReturnRequestByOrderID(ctx context.Context, orderID int64) (*domain.ReturnRequest, error) {
	query := `
		SELECT id, order_id, user_id, return_reason, is_approved, requested_date, approved_at, rejected_at
		FROM return_requests
		WHERE order_id = $1
	`
	var returnRequest domain.ReturnRequest
	var approvedAt, rejectedAt sql.NullTime
	err := r.db.QueryRowContext(ctx, query, orderID).Scan(
		&returnRequest.ID,
		&returnRequest.OrderID,
		&returnRequest.UserID,
		&returnRequest.ReturnReason,
		&returnRequest.IsApproved,
		&returnRequest.RequestedDate,
		&approvedAt,
		&rejectedAt,
	)
	if err == sql.ErrNoRows {
		return nil, utils.ErrReturnRequestNotFound
	}
	if err != nil {
		return nil, err
	}
	if approvedAt.Valid {
		returnRequest.ApprovedAt = &approvedAt.Time
	}
	if rejectedAt.Valid {
		returnRequest.RejectedAt = &rejectedAt.Time
	}
	return &returnRequest, nil
}

func (r *returnRepository) UpdateReturnRequestStatus(ctx context.Context, returnID int64, isApproved bool) error {
	var query string
	var args []interface{}
	if isApproved {
		query = `UPDATE return_requests SET is_approved = $1, approved_at = $2 WHERE id = $3`
		args = []interface{}{true, time.Now(), returnID}
	} else {
		query = `UPDATE return_requests SET is_approved = $1, rejected_at = $2 WHERE id = $3`
		args = []interface{}{false, time.Now(), returnID}
	}
	_, err := r.db.ExecContext(ctx, query, args...)
	return err
}

func (r *returnRepository) GetUserReturnRequests(ctx context.Context, userID int64) ([]*domain.ReturnRequest, error) {
	query := `
		SELECT id, order_id, user_id, return_reason, is_approved, requested_date, approved_at, rejected_at
		FROM return_requests
		WHERE user_id = $1
		ORDER BY requested_date DESC
	`
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var returnRequests []*domain.ReturnRequest
	for rows.Next() {
		var returnRequest domain.ReturnRequest
		var approvedAt, rejectedAt sql.NullTime
		err := rows.Scan(
			&returnRequest.ID,
			&returnRequest.OrderID,
			&returnRequest.UserID,
			&returnRequest.ReturnReason,
			&returnRequest.IsApproved,
			&returnRequest.RequestedDate,
			&approvedAt,
			&rejectedAt,
		)
		if err != nil {
			return nil, err
		}
		if approvedAt.Valid {
			returnRequest.ApprovedAt = &approvedAt.Time
		}
		if rejectedAt.Valid {
			returnRequest.RejectedAt = &rejectedAt.Time
		}
		returnRequests = append(returnRequests, &returnRequest)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return returnRequests, nil
}

func (r *returnRepository) UpdateApprovedOrRejected(ctx context.Context, returnRequest *domain.ReturnRequest) error {
	query := `
        UPDATE return_requests
        SET is_approved = $1, approved_at = $2, rejected_at = $3
        WHERE id = $4
    `
	_, err := r.db.ExecContext(ctx, query,
		returnRequest.IsApproved, returnRequest.ApprovedAt, returnRequest.RejectedAt,
		returnRequest.ID)
	return err
}

func (r *returnRepository) GetByID(ctx context.Context, id int64) (*domain.ReturnRequest, error) {
	query := `
        SELECT id, order_id, user_id, return_reason, is_approved, requested_date, approved_at, rejected_at
        FROM return_requests
        WHERE id = $1
    `
	var returnRequest domain.ReturnRequest
	var isApproved sql.NullBool
	var approvedAt, rejectedAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&returnRequest.ID, &returnRequest.OrderID, &returnRequest.UserID, &returnRequest.ReturnReason,
		&isApproved, &returnRequest.RequestedDate, &approvedAt, &rejectedAt,
	)
	if err != nil {
		return nil, err
	}

	if isApproved.Valid {
		returnRequest.IsApproved = isApproved.Bool
	}
	if approvedAt.Valid {
		returnRequest.ApprovedAt = &approvedAt.Time
	}
	if rejectedAt.Valid {
		returnRequest.RejectedAt = &rejectedAt.Time
	}

	return &returnRequest, nil
}

func (r *returnRepository) UpdateRefundDetails(ctx context.Context, returnRequest *domain.ReturnRequest) error {
	query := `
        UPDATE return_requests
        SET is_approved = $1, refund_initiated = $2, refund_amount = $3
        WHERE id = $4
    `
	_, err := r.db.ExecContext(ctx, query,
		returnRequest.IsApproved,
		returnRequest.RefundInitiated,
		returnRequest.RefundAmount,
		returnRequest.ID)
	return err
}
