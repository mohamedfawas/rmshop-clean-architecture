package postgres

import (
	"context"
	"database/sql"
	"log"
	"time"

	"github.com/mohamedfawas/rmshop-clean-architecture/internal/domain"
	"github.com/mohamedfawas/rmshop-clean-architecture/pkg/utils"
)

type checkoutRepository struct {
	db *sql.DB
}

func NewCheckoutRepository(db *sql.DB) *checkoutRepository {
	return &checkoutRepository{db: db}
}

func (r *checkoutRepository) CreateCheckoutSession(ctx context.Context, userID int64) (*domain.CheckoutSession, error) {
	query := `
        INSERT INTO checkout_sessions (user_id, status, created_at, updated_at)
        VALUES ($1, 'pending', NOW(), NOW())
        RETURNING id, created_at
    `
	var session domain.CheckoutSession
	err := r.db.QueryRowContext(ctx, query, userID).Scan(&session.ID, &session.CreatedAt)
	if err != nil {
		log.Printf("error while creating checkout session entry : %v", err)
		return nil, err
	}
	session.UserID = userID
	session.Status = "pending"
	return &session, nil
}

func (r *checkoutRepository) AddCheckoutItems(ctx context.Context, sessionID int64, items []*domain.CheckoutItem) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `
        INSERT INTO checkout_items (session_id, product_id, quantity, price, subtotal)
        VALUES ($1, $2, $3, $4, $5)
    `
	for _, item := range items {
		_, err := tx.ExecContext(ctx, query, sessionID, item.ProductID, item.Quantity, item.Price, item.Subtotal)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *checkoutRepository) GetCartItems(ctx context.Context, userID int64) ([]*domain.CartItemWithProduct, error) {
	query := `
        SELECT ci.id, ci.product_id, ci.quantity, p.name, p.price
        FROM cart_items ci
        JOIN products p ON ci.product_id = p.id
        WHERE ci.user_id = $1
    `
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*domain.CartItemWithProduct
	for rows.Next() {
		var item domain.CartItemWithProduct
		err := rows.Scan(&item.ID, &item.ProductID, &item.Quantity, &item.ProductName, &item.ProductPrice)
		if err != nil {
			return nil, err
		}
		items = append(items, &item)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

func (r *checkoutRepository) GetCheckoutByID(ctx context.Context, checkoutID int64) (*domain.CheckoutSession, error) {
	query := `
        SELECT id, user_id, total_amount, discount_amount, final_amount, item_count, created_at, updated_at, status, coupon_code, coupon_applied, address_id
        FROM checkout_sessions
        WHERE id = $1
    `
	var checkout domain.CheckoutSession
	var couponCode sql.NullString
	var addressID sql.NullInt64
	err := r.db.QueryRowContext(ctx, query, checkoutID).Scan(
		&checkout.ID, &checkout.UserID, &checkout.TotalAmount, &checkout.DiscountAmount, &checkout.FinalAmount,
		&checkout.ItemCount, &checkout.CreatedAt, &checkout.UpdatedAt, &checkout.Status,
		&couponCode, &checkout.CouponApplied, &addressID,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, utils.ErrCheckoutNotFound
		}
		log.Printf("error while retrieving checkout sessions : %v", err)
		return nil, err
	}

	// Assign the coupon code only if it's not NULL
	if couponCode.Valid {
		checkout.CouponCode = couponCode.String
	} else {
		checkout.CouponCode = "" // or however you want to represent a missing coupon code
	}
	if addressID.Valid {
		checkout.AddressID = addressID.Int64
	}

	return &checkout, nil
}

func (r *checkoutRepository) UpdateCheckout(ctx context.Context, checkout *domain.CheckoutSession) error {
	query := `
        UPDATE checkout_sessions
        SET total_amount = $1, discount_amount = $2, final_amount = $3, updated_at = $4, 
            coupon_code = $5, coupon_applied = $6
        WHERE id = $7
    `
	_, err := r.db.ExecContext(ctx, query,
		checkout.TotalAmount, checkout.DiscountAmount, checkout.FinalAmount, time.Now(),
		checkout.CouponCode, checkout.CouponApplied, checkout.ID,
	)
	if err != nil {
		log.Printf("error while updating checkout session : %v", err)
	}
	return err
}

func (r *couponRepository) IsApplied(ctx context.Context, checkoutID int64) (bool, error) {
	query := `SELECT coupon_applied FROM checkout_sessions WHERE id = $1`
	var isApplied bool
	err := r.db.QueryRowContext(ctx, query, checkoutID).Scan(&isApplied)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, utils.ErrCheckoutNotFound
		}
		log.Printf("error while retrieving applied coupon details : %v", err)
		return false, err
	}
	return isApplied, nil
}

func (r *checkoutRepository) GetCheckoutItems(ctx context.Context, checkoutID int64) ([]*domain.CheckoutItem, error) {
	query := `
        SELECT id, product_id, quantity, price, subtotal
        FROM checkout_items
        WHERE session_id = $1
    `
	rows, err := r.db.QueryContext(ctx, query, checkoutID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*domain.CheckoutItem
	for rows.Next() {
		var item domain.CheckoutItem
		err := rows.Scan(&item.ID, &item.ProductID, &item.Quantity, &item.Price, &item.Subtotal)
		if err != nil {
			return nil, err
		}
		items = append(items, &item)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

func (r *checkoutRepository) UpdateCheckoutAddress(ctx context.Context, checkoutID int64, addressID int64) error {
	query := `UPDATE checkout_sessions SET address_id = $1, updated_at = NOW() WHERE id = $2`
	_, err := r.db.ExecContext(ctx, query, addressID, checkoutID)
	if err != nil {
		log.Printf("error while updating checkout address : %v", err)
	}
	return err
}

func (r *checkoutRepository) AddNewAddressToCheckout(ctx context.Context, checkoutID int64, address *domain.UserAddress) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		log.Printf("error : %v", err)
		return err
	}
	defer tx.Rollback()

	// Insert new address
	addressQuery := `INSERT INTO user_address (user_id, address_line1, address_line2, city, state, pincode, phone_number) 
                     VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id`
	var newAddressID int64
	err = tx.QueryRowContext(ctx, addressQuery, address.UserID, address.AddressLine1, address.AddressLine2,
		address.City, address.State, address.PinCode, address.PhoneNumber).Scan(&newAddressID)
	if err != nil {
		log.Printf("error while adding new address : %v", err)
		return err
	}

	// Update checkout with new address
	checkoutQuery := `UPDATE checkout_sessions SET address_id = $1 WHERE id = $2`
	_, err = tx.ExecContext(ctx, checkoutQuery, newAddressID, checkoutID)
	if err != nil {
		log.Printf("error while updating the address id in checkout_sessions table : %v", err)
		return err
	}

	return tx.Commit()
}
