package postgres

import (
	"context"
	"database/sql"
	"log"

	"github.com/mohamedfawas/rmshop-clean-architecture/internal/domain"
	"github.com/mohamedfawas/rmshop-clean-architecture/pkg/utils"
)

type returnRepository struct {
	db *sql.DB
}

func NewReturnRepository(db *sql.DB) *returnRepository {
	return &returnRepository{db: db}
}

/*
CreateReturnRequest:
- Create an entry in return_requests table
*/
func (r *returnRepository) CreateReturnRequest(ctx context.Context, returnRequest *domain.ReturnRequest) error {
	query := `
		INSERT INTO return_requests (order_id, user_id, return_reason, is_approved, requested_date, is_order_reached_the_seller, is_stock_updated )
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`
	err := r.db.QueryRowContext(ctx, query,
		returnRequest.OrderID,
		returnRequest.UserID,
		returnRequest.ReturnReason,
		returnRequest.IsApproved,
		returnRequest.RequestedDate,
		returnRequest.IsOrderReachedTheSeller,
		returnRequest.IsStockUpdated,
	).Scan(&returnRequest.ID)
	if err != nil {
		log.Printf("error while adding return request enty : %v", err)
	}
	return err
}

func (r *returnRepository) GetReturnRequestByOrderID(ctx context.Context, orderID int64) (*domain.ReturnRequest, error) {
	query := `
		SELECT id, order_id, user_id, return_reason, is_approved, requested_date, 
		approved_at, rejected_at, refund_initiated, refund_amount, 
		order_returned_to_seller_at, is_order_reached_the_seller, is_stock_updated
		FROM return_requests
		WHERE order_id = $1
	`
	var returnRequest domain.ReturnRequest
	var approvedAt, rejectedAt, orderReturnedToSellerAt sql.NullTime
	var refund_amount sql.NullFloat64
	err := r.db.QueryRowContext(ctx, query, orderID).Scan(
		&returnRequest.ID,
		&returnRequest.OrderID,
		&returnRequest.UserID,
		&returnRequest.ReturnReason,
		&returnRequest.IsApproved,
		&returnRequest.RequestedDate,
		&approvedAt,
		&rejectedAt,
		&returnRequest.RefundInitiated,
		&refund_amount,
		&orderReturnedToSellerAt,
		&returnRequest.IsOrderReachedTheSeller,
		&returnRequest.IsStockUpdated,
	)
	if err == sql.ErrNoRows {
		return nil, utils.ErrReturnRequestNotFound
	}
	if err != nil {
		log.Printf("error while fetching return requests using order id : %v", err)
		return nil, err
	}
	if approvedAt.Valid {
		returnRequest.ApprovedAt = &approvedAt.Time
	}
	if rejectedAt.Valid {
		returnRequest.RejectedAt = &rejectedAt.Time
	}
	if refund_amount.Valid {
		returnRequest.RefundAmount = &refund_amount.Float64
	}

	if orderReturnedToSellerAt.Valid {
		returnRequest.OrderReturnedToSellerAt = &orderReturnedToSellerAt.Time
	}
	return &returnRequest, nil
}

func (r *returnRepository) GetUserReturnRequests(ctx context.Context, userID int64) ([]*domain.ReturnRequest, error) {
	query := `
		SELECT id, order_id, user_id, return_reason, is_approved, requested_date, approved_at, rejected_at,
		refund_initiated, refund_amount, 
		order_returned_to_seller_at, is_order_reached_the_seller, is_stock_updated
		FROM return_requests
		WHERE user_id = $1
		ORDER BY requested_date DESC
	`
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		log.Printf("error while executing query in GetUserReturnRequests : %v", err)
		return nil, err
	}
	defer rows.Close()

	var returnRequests []*domain.ReturnRequest
	for rows.Next() {
		var returnRequest domain.ReturnRequest
		var approvedAt, rejectedAt, orderReturnedToSellerAt sql.NullTime
		var refund_amount sql.NullFloat64
		err := rows.Scan(
			&returnRequest.ID,
			&returnRequest.OrderID,
			&returnRequest.UserID,
			&returnRequest.ReturnReason,
			&returnRequest.IsApproved,
			&returnRequest.RequestedDate,
			&approvedAt,
			&rejectedAt,
			&returnRequest.RefundInitiated,
			&refund_amount,
			&orderReturnedToSellerAt,
			&returnRequest.IsOrderReachedTheSeller,
			&returnRequest.IsStockUpdated,
		)
		if err != nil {
			log.Printf("error while iterating over rows of return requests : %v", err)
			return nil, err
		}
		if approvedAt.Valid {
			returnRequest.ApprovedAt = &approvedAt.Time
		}
		if rejectedAt.Valid {
			returnRequest.RejectedAt = &rejectedAt.Time
		}
		if refund_amount.Valid {
			returnRequest.RefundAmount = &refund_amount.Float64
		}

		if orderReturnedToSellerAt.Valid {
			returnRequest.OrderReturnedToSellerAt = &orderReturnedToSellerAt.Time
		}
		returnRequests = append(returnRequests, &returnRequest)
	}
	if err = rows.Err(); err != nil {
		log.Printf("more error captured while iterating over rows : %v", err)
		return nil, err
	}
	return returnRequests, nil
}

/*
UpdateApprovedOrRejected:
- record admin's approval or rejection for the given order return request
- Updates return_requests table
*/
func (r *returnRepository) UpdateApprovedOrRejected(ctx context.Context, returnRequest *domain.ReturnRequest) error {
	query := `
        UPDATE return_requests
        SET is_approved = $1, approved_at = $2, rejected_at = $3
        WHERE id = $4
    `
	_, err := r.db.ExecContext(ctx, query,
		returnRequest.IsApproved,
		returnRequest.ApprovedAt,
		returnRequest.RejectedAt,
		returnRequest.ID)
	if err != nil {
		log.Printf("failed to update approved or rejected : %v", err)
	}
	return err
}

/*
UpdateRefundDetails:
- Update refund related details in return_requests table
*/
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
	if err != nil {
		log.Printf("failed to update refund details in return_requests table : %v", err)
	}
	return err
}

func (r *returnRepository) BeginTx(ctx context.Context) (*sql.Tx, error) {
	return r.db.BeginTx(ctx, nil)
}

/*
GetByID:
- get return request details using return id
*/
func (r *returnRepository) GetByID(ctx context.Context, id int64) (*domain.ReturnRequest, error) {
	query := `
		SELECT id, order_id, user_id, return_reason, is_approved, requested_date, 
			   approved_at, rejected_at, refund_initiated,  refund_amount, 
			   order_returned_to_seller_at, is_order_reached_the_seller, is_stock_updated
		FROM return_requests
		WHERE id = $1
	`
	var returnRequest domain.ReturnRequest
	var refundAmount sql.NullFloat64
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&returnRequest.ID,
		&returnRequest.OrderID,
		&returnRequest.UserID,
		&returnRequest.ReturnReason,
		&returnRequest.IsApproved,
		&returnRequest.RequestedDate,
		&returnRequest.ApprovedAt,
		&returnRequest.RejectedAt,
		&returnRequest.RefundInitiated,
		&refundAmount,
		&returnRequest.OrderReturnedToSellerAt,
		&returnRequest.IsOrderReachedTheSeller,
		&returnRequest.IsStockUpdated,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, utils.ErrReturnRequestNotFound
		}
		log.Printf("failed to fetch the return request : %v", err)
		return nil, err
	}
	if refundAmount.Valid {
		returnRequest.RefundAmount = &refundAmount.Float64
	}
	return &returnRequest, nil
}

func (r *returnRepository) MarkOrderReturnedToSeller(ctx context.Context, returnID int64) error {
	query := `
        UPDATE return_requests
        SET order_returned_to_seller_at = NOW(),
            is_order_reached_the_seller = true
        WHERE id = $1 AND is_approved = true 
        RETURNING id
    `
	var id int64
	err := r.db.QueryRowContext(ctx, query, returnID).Scan(&id)
	if err == sql.ErrNoRows {
		return utils.ErrReturnRequestNotFound
	}
	return err
}

/*
UpdateReturnRequest:
- Update return_requests table
- is_stock_updated, order_returned_to_seller_at, is_order_reached_the_seller
*/
func (r *returnRepository) UpdateReturnRequest(ctx context.Context, returnRequest *domain.ReturnRequest) error {
	query := `UPDATE return_requests 
              SET is_stock_updated = $1, order_returned_to_seller_at = $2, is_order_reached_the_seller = $3
              WHERE id = $4`
	_, err := r.db.ExecContext(ctx, query,
		returnRequest.IsStockUpdated,
		returnRequest.OrderReturnedToSellerAt,
		returnRequest.IsOrderReachedTheSeller,
		returnRequest.ID)
	return err
}

func (r *returnRepository) GetPendingReturnRequests(ctx context.Context, params domain.ReturnRequestParams) ([]*domain.ReturnRequest, int64, error) {
	query := `
        SELECT id, order_id, user_id, return_reason, is_approved, requested_date, approved_at, rejected_at
        FROM return_requests
        WHERE is_approved = false AND rejected_at IS NULL
        ORDER BY requested_date DESC
        LIMIT $1 OFFSET $2
    `
	countQuery := `
        SELECT COUNT(*)
        FROM return_requests
        WHERE is_approved = false AND rejected_at IS NULL
    `

	// Get total count
	var totalCount int64
	err := r.db.QueryRowContext(ctx, countQuery).Scan(&totalCount)
	if err != nil {
		return nil, 0, err
	}

	// Execute main query
	rows, err := r.db.QueryContext(ctx, query, params.Limit, (params.Page-1)*params.Limit)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var requests []*domain.ReturnRequest
	for rows.Next() {
		var req domain.ReturnRequest
		err := rows.Scan(
			&req.ID, &req.OrderID, &req.UserID, &req.ReturnReason,
			&req.IsApproved, &req.RequestedDate, &req.ApprovedAt, &req.RejectedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		requests = append(requests, &req)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, err
	}

	return requests, totalCount, nil
}
