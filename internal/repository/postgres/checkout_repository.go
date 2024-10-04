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

/*
	GetCartItems :

Get the current cart of the user.
Take id,product_id, quantity from cart items
Take product name and product price from products table
*/
func (r *checkoutRepository) GetCartItems(ctx context.Context, userID int64) ([]*domain.CartItemWithProduct, error) {
	// Inner join cart items and products table
	// Retrieve cart items, product name, product price
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

	// here we use pointer for memory efficiency :  creating a slice of pointers to CartItemWithProduct structs.
	var items []*domain.CartItemWithProduct

	// start iterating over each row
	for rows.Next() { // move the cursor to the next row
		var item domain.CartItemWithProduct
		// read the column values in the current row
		err := rows.Scan(&item.ID, &item.ProductID, &item.Quantity, &item.ProductName, &item.ProductPrice)
		if err != nil {
			log.Printf("error while scanning rows of cart items and products : %v", err)
			return nil, err
		}
		items = append(items, &item)
	}

	// check whether any error occured that weren't caught in the loop
	if err = rows.Err(); err != nil {
		log.Printf("error while iterating over cart items and products : %v", err)
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

/*
UpdateCheckoutStatus:
- Update checkout status in checkout_sessions table
*/
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

/*
CreateOrGetShippingAddress:
- Select values from user_address table and then insert those values into shipping _addresses table.
- On conflict, update the existing values with new values.
*/
func (r *checkoutRepository) CreateOrGetShippingAddress(ctx context.Context, userID, addressID int64) (int64, error) {
	// start transaction
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		log.Printf("error while starting transaction in CreateOrGetShippingAddress method : %v", err)
		return 0, err
	}
	defer tx.Rollback()

	var shippingAddressID int64
	err = tx.QueryRowContext(ctx, `
        INSERT INTO shipping_addresses (user_id, 
									address_id, 
									address_line1, 
									address_line2, 
									city, 
									state, 
									landmark, 
									pincode, 
									phone_number)
        SELECT $1, 
				ua.id, 
				ua.address_line1, 
				ua.address_line2, 
				ua.city, ua.state, 
				ua.landmark, 
				ua.pincode, 
				ua.phone_number
				FROM user_address ua
        WHERE 
			ua.id = $2 
			AND ua.user_id = $1  		-- Ensure address belongs to the same user by matching user_id
        ON CONFLICT (user_id, address_id)  -- Manage conflicts if a record with the same (user_id, address_id) already exists
        DO UPDATE SET 
			-- On conflict, update with new values
			-- EXCLUDED keyword provides access to the values that you were trying to insert.
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
		log.Printf("error while adding values to shipping_addresses table : %v", err)
		return 0, err
	}

	if err = tx.Commit(); err != nil {
		log.Printf("error while commiting transaction in Update shipping address process : %v", err)
		return 0, err
	}

	return shippingAddressID, nil
}

/*
UpdateCheckoutShippingAddress:
- update checkout_sessions table with the shipping_address id
*/
func (r *checkoutRepository) UpdateCheckoutShippingAddress(ctx context.Context, checkoutID, shippingAddressID int64) error {
	query := `
        UPDATE checkout_sessions 
        SET shipping_address_id = $1, updated_at = NOW() 
        WHERE id = $2
    `
	result, err := r.db.ExecContext(ctx, query, shippingAddressID, checkoutID)
	if err != nil {
		log.Printf("error while updating checkout_sessions table with the new shipping_address_id : %v", err)
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("error while checking rows affected in UpdateCheckoutShippingAddress method in checkout_repository : %v", err)
		return err
	}

	if rowsAffected == 0 {
		return utils.ErrCheckoutNotFound
	}

	return nil
}

/*
GetShippingAddress:
- Get values from shipping_addresses table
*/
func (r *checkoutRepository) GetShippingAddress(ctx context.Context, addressID int64) (*domain.ShippingAddress, error) {
	query := `
		SELECT id, user_id,address_id, address_line1, address_line2, city, state, pincode, landmark, phone_number
		FROM shipping_addresses
		WHERE id = $1
	`
	var address domain.ShippingAddress
	err := r.db.QueryRowContext(ctx, query, addressID).Scan(
		&address.ID,
		&address.UserID,
		&address.AddressID,
		&address.AddressLine1,
		&address.AddressLine2,
		&address.City,
		&address.State,
		&address.PinCode,
		&address.Landmark,
		&address.PhoneNumber,
	)
	if err != nil {
		log.Printf("error while retrieving values from shipping address table : %v", err)
		return nil, err
	}
	return &address, nil
}

// ///////////////////////////////////////////////////////////////////////////////////////
func (r *checkoutRepository) GetOrCreateCheckoutSession(ctx context.Context, userID int64) (*domain.CheckoutSession, error) {
	// First, try to get an existing session
	query := `
        SELECT id, user_id, total_amount, item_count, status, coupon_applied, coupon_code, discount_amount, final_amount, shipping_address_id, created_at, updated_at
        FROM checkout_sessions
        WHERE user_id = $1 AND status = 'pending' AND is_deleted= false
        ORDER BY created_at DESC
        LIMIT 1
    `
	var session domain.CheckoutSession
	var shippingAddressId sql.NullInt64
	var couponCode sql.NullString
	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&session.ID,
		&session.UserID,
		&session.TotalAmount,
		&session.ItemCount,
		&session.Status,
		&session.CouponApplied,
		&couponCode,
		&session.DiscountAmount,
		&session.FinalAmount,
		&shippingAddressId,
		&session.CreatedAt,
		&session.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		// If no session exists, create a new one
		return r.CreateCheckoutSession(ctx, userID)
	} else if err != nil {
		log.Printf("error while getting checkout session: %v", err)
		return nil, err
	}

	if shippingAddressId.Valid {
		session.ShippingAddressID = shippingAddressId.Int64
	}

	if couponCode.Valid {
		session.CouponCode = couponCode.String
	}

	return &session, nil
}

func (r *checkoutRepository) CreateCheckoutSession(ctx context.Context, userID int64) (*domain.CheckoutSession, error) {
	query := `
        INSERT INTO checkout_sessions (user_id, status, created_at, updated_at)
        VALUES ($1, 'pending', NOW(), NOW())
        RETURNING id, created_at, updated_at
    `
	var session domain.CheckoutSession
	err := r.db.QueryRowContext(ctx, query,
		userID).Scan(&session.ID,
		&session.CreatedAt,
		&session.UpdatedAt)
	if err != nil {
		log.Printf("error while creating checkout session entry: %v", err)
		return nil, err
	}

	session.UserID = userID
	session.Status = utils.CheckoutStatusPending
	return &session, nil
}

/*
UpdateCheckoutDetails:
- Update values in checkout_sessions table
- total_amount, discount_amount, final_amount, coupon_code, coupon_applied and item_count are updated
*/
func (r *checkoutRepository) UpdateCheckoutDetails(ctx context.Context, checkout *domain.CheckoutSession) error {
	query := `
        UPDATE checkout_sessions
        SET total_amount = $1, discount_amount = $2, final_amount = $3, updated_at = $4, 
            coupon_code = $5, coupon_applied = $6, item_count = $7
        WHERE id = $8
    `
	result, err := r.db.ExecContext(ctx, query,
		checkout.TotalAmount,
		checkout.DiscountAmount,
		checkout.FinalAmount,
		time.Now().UTC(),
		checkout.CouponCode,
		checkout.CouponApplied,
		checkout.ItemCount,
		checkout.ID,
	)
	if err != nil {
		log.Printf("error while updating checkout_sessions table: %v", err)
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("error while checking rows affected in checkout_sessions table")
		return err
	}

	if rowsAffected == 0 {
		return utils.ErrCheckoutNotFound
	}

	return nil
}

/*
GetCheckoutByID:
- Get rows from checkout_sessions table
*/
func (r *checkoutRepository) GetCheckoutByID(ctx context.Context, checkoutID int64) (*domain.CheckoutSession, error) {
	query := `
        SELECT id, user_id, total_amount, discount_amount, final_amount, 
               item_count, created_at, updated_at, status, coupon_code, 
               coupon_applied, shipping_address_id
        FROM checkout_sessions
        WHERE id = $1
    `
	var shippingAddrID sql.NullInt64
	var checkout domain.CheckoutSession
	var couponCode sql.NullString
	err := r.db.QueryRowContext(ctx, query, checkoutID).Scan(
		&checkout.ID,
		&checkout.UserID,
		&checkout.TotalAmount,
		&checkout.DiscountAmount,
		&checkout.FinalAmount,
		&checkout.ItemCount,
		&checkout.CreatedAt,
		&checkout.UpdatedAt,
		&checkout.Status,
		&couponCode,
		&checkout.CouponApplied,
		&shippingAddrID,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, utils.ErrCheckoutNotFound
		}
		log.Printf("error while retrieiving rows from checkout_sessions table : %v", err)
		return nil, err
	}

	// If not null, get the coupon code
	if couponCode.Valid {
		checkout.CouponCode = couponCode.String
	}

	// If not null, get the shipping address id
	if shippingAddrID.Valid {
		checkout.ShippingAddressID = shippingAddrID.Int64
	}

	return &checkout, nil
}

func (r *checkoutRepository) GetCheckoutSession(ctx context.Context, userID int64) (*domain.CheckoutSession, error) {
	// First, try to get an existing session
	query := `
        SELECT id, user_id, total_amount, item_count, status, coupon_applied, coupon_code, discount_amount, final_amount, shipping_address_id, created_at, updated_at
        FROM checkout_sessions
        WHERE user_id = $1 AND status = 'pending' AND is_deleted=false
        ORDER BY created_at DESC
        LIMIT 1
    `
	var session domain.CheckoutSession
	var shippingAddressId sql.NullInt64
	var couponCode sql.NullString
	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&session.ID,
		&session.UserID,
		&session.TotalAmount,
		&session.ItemCount,
		&session.Status,
		&session.CouponApplied,
		&couponCode,
		&session.DiscountAmount,
		&session.FinalAmount,
		&shippingAddressId,
		&session.CreatedAt,
		&session.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		// If no session exists, create a new one
		return nil, utils.ErrCheckoutNotFound
	} else if err != nil {
		log.Printf("error while getting checkout session: %v", err)
		return nil, err
	}

	if shippingAddressId.Valid {
		session.ShippingAddressID = shippingAddressId.Int64
	}

	if couponCode.Valid {
		session.CouponCode = couponCode.String
	}

	return &session, nil
}

func (r *checkoutRepository) MarkCheckoutAsDeleted(ctx context.Context, tx *sql.Tx, checkoutID int64) error {
	query := `UPDATE checkout_sessions SET is_deleted = true, updated_at = NOW() WHERE id = $1`
	_, err := tx.ExecContext(ctx, query, checkoutID)
	if err != nil {
		log.Printf("error while marking checkout as deleted: %v", err)
	}
	return err
}
