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
		log.Printf("error while retrieving cart item details : %v", err)
		return nil, err
	}
	defer rows.Close()

	var items []*domain.CartItemWithProduct
	for rows.Next() {
		var item domain.CartItemWithProduct
		err := rows.Scan(&item.ID, &item.ProductID, &item.Quantity, &item.ProductName, &item.ProductPrice)
		if err != nil {
			log.Printf("error : %v", err)
			return nil, err
		}
		items = append(items, &item)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
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

func (r *checkoutRepository) GetCheckoutWithItems(ctx context.Context, checkoutID int64) (*domain.CheckoutSummary, error) {
	query := `
        SELECT cs.id, cs.user_id, cs.total_amount, cs.discount_amount, cs.final_amount, 
               cs.item_count, cs.status, cs.coupon_code, cs.coupon_applied, cs.shipping_address_id,
               ci.id, ci.product_id, p.name, ci.quantity, ci.price, ci.subtotal,
               sa.address_line1, sa.address_line2, sa.city, sa.state, sa.pincode, sa.phone_number
        FROM checkout_sessions cs
        LEFT JOIN checkout_items ci ON cs.id = ci.session_id
        LEFT JOIN products p ON ci.product_id = p.id
        LEFT JOIN shipping_addresses sa ON cs.shipping_address_id = sa.id
        WHERE cs.id = $1
    `

	rows, err := r.db.QueryContext(ctx, query, checkoutID)
	if err != nil {
		log.Printf("error while getting checkout with items : %v", err)
		return nil, err
	}
	defer rows.Close()

	var summary domain.CheckoutSummary
	var items []*domain.CheckoutItemDetail
	var address domain.UserAddress
	addressSet := false

	for rows.Next() {
		var item domain.CheckoutItemDetail
		var couponCode, addressLine1, addressLine2, city, state, pincode, phoneNumber sql.NullString
		var shippingAddressID sql.NullInt64

		err := rows.Scan(
			&summary.ID, &summary.UserID, &summary.TotalAmount, &summary.DiscountAmount, &summary.FinalAmount,
			&summary.ItemCount, &summary.Status, &couponCode, &summary.CouponApplied, &shippingAddressID,
			&item.ID, &item.ProductID, &item.Name, &item.Quantity, &item.Price, &item.Subtotal,
			&addressLine1, &addressLine2, &city, &state, &pincode, &phoneNumber,
		)
		if err != nil {
			log.Printf("error while scanning row : %v", err)
			return nil, err
		}

		if couponCode.Valid {
			summary.CouponCode = couponCode.String
		}

		items = append(items, &item)

		if !addressSet && shippingAddressID.Valid {
			address = domain.UserAddress{
				ID:           shippingAddressID.Int64,
				AddressLine1: addressLine1.String,
				AddressLine2: addressLine2.String,
				City:         city.String,
				State:        state.String,
				PinCode:      pincode.String,
				PhoneNumber:  phoneNumber.String,
			}
			addressSet = true
		}
	}

	if err = rows.Err(); err != nil {
		log.Printf("error after iterating rows : %v", err)
		return nil, err
	}

	summary.Items = items
	if addressSet {
		summary.Address = &address
	}

	// Handle empty checkout
	if len(items) == 0 {
		summary.Status = "empty"
		summary.TotalAmount = 0
		summary.DiscountAmount = 0
		summary.FinalAmount = 0
		summary.ItemCount = 0
	} else {
		// Recalculate item count
		var itemCount int
		for _, item := range items {
			itemCount += item.Quantity
		}
		summary.ItemCount = itemCount
	}

	// Update the checkout session with the correct item count
	updateQuery := `UPDATE checkout_sessions SET item_count = $1 WHERE id = $2`
	_, err = r.db.ExecContext(ctx, updateQuery, summary.ItemCount, checkoutID)
	if err != nil {
		log.Printf("error updating item count: %v", err)
		return nil, err
	}

	return &summary, nil
}

func (r *checkoutRepository) UpdateCheckoutDetails(ctx context.Context, checkout *domain.CheckoutSession) error {
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
		log.Printf("error while updating checkout session details: %v", err)
	}
	return err
}

func (r *checkoutRepository) UpdateCheckoutStatus(ctx context.Context, tx *sql.Tx, checkout *domain.CheckoutSession) error {
	query := `
		UPDATE checkout_sessions
		SET status = $1, updated_at = NOW()
		WHERE id = $2
	`
	_, err := tx.ExecContext(ctx, query, checkout.Status, checkout.ID)
	if err != nil {
		log.Printf("error while updating checkout session status: %v", err)
	}
	return err
}

func (r *checkoutRepository) BeginTx(ctx context.Context) (*sql.Tx, error) {
	return r.db.BeginTx(ctx, nil)
}

func (r *checkoutRepository) CreateOrGetShippingAddress(ctx context.Context, userID, addressID int64) (int64, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	var shippingAddressID int64
	err = tx.QueryRowContext(ctx, `
        INSERT INTO shipping_addresses (user_id, address_id, address_line1, address_line2, city, state, landmark, pincode, phone_number)
        SELECT $1, ua.id, ua.address_line1, ua.address_line2, ua.city, ua.state, ua.landmark, ua.pincode, ua.phone_number
        FROM user_address ua
        WHERE ua.id = $2 AND ua.user_id = $1
        ON CONFLICT (user_id, address_id) 
        DO UPDATE SET 
            address_line1 = EXCLUDED.address_line1,
            address_line2 = EXCLUDED.address_line2,
            city = EXCLUDED.city,
            state = EXCLUDED.state,
            landmark = EXCLUDED.landmark,
            pincode = EXCLUDED.pincode,
            phone_number = EXCLUDED.phone_number,
            created_at = CURRENT_TIMESTAMP
        RETURNING id
    `, userID, addressID).Scan(&shippingAddressID)
	if err != nil {
		return 0, err
	}

	if err = tx.Commit(); err != nil {
		return 0, err
	}

	return shippingAddressID, nil
}

func (r *checkoutRepository) GetCheckoutWithAddressByID(ctx context.Context, checkoutID int64) (*domain.CheckoutSession, error) {
	query := `
        SELECT cs.id, cs.user_id, cs.total_amount, cs.discount_amount, cs.final_amount, 
               cs.item_count, cs.status, cs.coupon_code, cs.coupon_applied, cs.shipping_address_id,
               sa.id, sa.address_line1, sa.address_line2, sa.city, sa.state, sa.landmark, sa.pincode, sa.phone_number
        FROM checkout_sessions cs
        LEFT JOIN shipping_addresses sa ON cs.shipping_address_id = sa.id
        WHERE cs.id = $1
    `
	var checkout domain.CheckoutSession
	var shippingAddressID sql.NullInt64
	var shippingAddress domain.ShippingAddress
	var couponCode, addressLine1, addressLine2, city, state, landmark, pincode, phoneNumber sql.NullString

	err := r.db.QueryRowContext(ctx, query, checkoutID).Scan(
		&checkout.ID, &checkout.UserID, &checkout.TotalAmount, &checkout.DiscountAmount, &checkout.FinalAmount,
		&checkout.ItemCount, &checkout.Status, &couponCode, &checkout.CouponApplied, &shippingAddressID,
		&shippingAddress.ID, &addressLine1, &addressLine2, &city, &state, &landmark, &pincode, &phoneNumber,
	)

	if err == sql.ErrNoRows {
		return nil, utils.ErrCheckoutNotFound
	}
	if err != nil {
		return nil, err
	}

	if couponCode.Valid {
		checkout.CouponCode = couponCode.String
	}

	if shippingAddressID.Valid {
		shippingAddress.AddressLine1 = addressLine1.String
		shippingAddress.AddressLine2 = addressLine2.String
		shippingAddress.City = city.String
		shippingAddress.State = state.String
		shippingAddress.Landmark = landmark.String
		shippingAddress.PinCode = pincode.String
		shippingAddress.PhoneNumber = phoneNumber.String
		checkout.ShippingAddress = &shippingAddress
	}

	return &checkout, nil
}

func (r *checkoutRepository) getShippingAddressById(ctx context.Context, id int64) (*domain.ShippingAddress, error) {
	query := `
        SELECT id, address_line1, address_line2, city, state, landmark, pincode, phone_number
        FROM shipping_addresses
        WHERE id = $1
    `
	var address domain.ShippingAddress
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&address.ID, &address.AddressLine1, &address.AddressLine2, &address.City,
		&address.State, &address.Landmark, &address.PinCode, &address.PhoneNumber,
	)
	if err != nil {
		return nil, err
	}
	return &address, nil
}

func (r *checkoutRepository) GetCheckoutByID(ctx context.Context, checkoutID int64) (*domain.CheckoutSession, error) {
	query := `
        SELECT id, user_id, total_amount, discount_amount, final_amount, 
               item_count, created_at, updated_at, status, coupon_code, 
               coupon_applied, shipping_address_id
        FROM checkout_sessions
        WHERE id = $1
    `
	var checkout domain.CheckoutSession
	var couponCode sql.NullString
	err := r.db.QueryRowContext(ctx, query, checkoutID).Scan(
		&checkout.ID, &checkout.UserID, &checkout.TotalAmount, &checkout.DiscountAmount,
		&checkout.FinalAmount, &checkout.ItemCount, &checkout.CreatedAt, &checkout.UpdatedAt,
		&checkout.Status, &couponCode, &checkout.CouponApplied, &checkout.ShippingAddressID,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, utils.ErrCheckoutNotFound
		}
		return nil, err
	}

	if couponCode.Valid {
		checkout.CouponCode = couponCode.String
	}

	return &checkout, nil
}

func (r *checkoutRepository) UpdateCheckoutShippingAddress(ctx context.Context, checkoutID, shippingAddressID int64) error {
	query := `
        UPDATE checkout_sessions 
        SET shipping_address_id = $1, updated_at = NOW() 
        WHERE id = $2
    `
	result, err := r.db.ExecContext(ctx, query, shippingAddressID, checkoutID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return utils.ErrCheckoutNotFound
	}

	return nil
}
