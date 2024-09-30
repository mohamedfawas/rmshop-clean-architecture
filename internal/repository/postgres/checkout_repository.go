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
	CreateCheckoutSession

- Create the checkout session entry (user_id, status, created_at, updated_at)
*/
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

	// Update session values after creating checkout sessions entry
	session.UserID = userID
	session.Status = "pending"
	return &session, nil
}

/*
	AddCheckoutItems :

- Adds Checkout item values in checkout_items table, done with transaction
*/
func (r *checkoutRepository) AddCheckoutItems(ctx context.Context, sessionID int64, items []*domain.CheckoutItem) error {
	// start transaction
	tx, err := r.db.BeginTx(ctx, nil) // nil means default behaviour will be used
	if err != nil {
		log.Printf("error while beginnint transaction in AddCheckoutItems method : %v", err)
		return err
	}
	defer tx.Rollback()

	// Define the query
	query := `
        INSERT INTO checkout_items (session_id, product_id, quantity, price, subtotal)
        VALUES ($1, $2, $3, $4, $5)
    `

	// Iterate through checkout items
	for _, item := range items {
		_, err := tx.ExecContext(ctx, query,
			sessionID,
			item.ProductID,
			item.Quantity,
			item.Price,
			item.Subtotal)
		if err != nil {
			log.Printf("error while adding each checkout item entry in checkout_items table : %v", err)
			return err
		}
	}

	// commit the transaction
	return tx.Commit()
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

/*
GetCheckoutItems :
- Retrieves values from checkout items table
- product_id, price, quantity and subtotal are retrieved
*/
func (r *checkoutRepository) GetCheckoutItems(ctx context.Context, checkoutID int64) ([]*domain.CheckoutItem, error) {
	query := `
        SELECT id, product_id, quantity, price, subtotal
        FROM checkout_items
        WHERE session_id = $1
    `
	rows, err := r.db.QueryContext(ctx, query, checkoutID)
	if err != nil {
		log.Printf("error while retrieving rows from checkout_items table : %v", err)
		return nil, err
	}
	defer rows.Close()

	// Define the slice to store each checkout items
	var items []*domain.CheckoutItem
	for rows.Next() {
		var item domain.CheckoutItem
		err := rows.Scan(&item.ID, &item.ProductID, &item.Quantity, &item.Price, &item.Subtotal)
		if err != nil {
			log.Printf("error while iterating over rows of checkout items : %v", err)
			return nil, err
		}
		items = append(items, &item)
	}

	if err = rows.Err(); err != nil {
		log.Printf("error while fetching rows of checkout items : %v", err)
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
GetCheckoutItemsWithProductDetails:
- Join checkout_items with products based on product id
*/
func (r *checkoutRepository) GetCheckoutItemsWithProductDetails(ctx context.Context, checkoutID int64) ([]*domain.CheckoutItemDetail, error) {
	query := `
		SELECT ci.id, ci.product_id, p.name, ci.quantity, ci.price, ci.subtotal
		FROM checkout_items ci
		JOIN products p ON ci.product_id = p.id
		WHERE ci.session_id = $1
	`
	rows, err := r.db.QueryContext(ctx, query, checkoutID)
	if err != nil {
		log.Printf("error while retrieving details from checkout_items table and products table : %v)", err)
		return nil, err
	}
	defer rows.Close()

	var items []*domain.CheckoutItemDetail
	for rows.Next() {
		var item domain.CheckoutItemDetail
		if err := rows.Scan(&item.ID,
			&item.ProductID,
			&item.Name,
			&item.Quantity,
			&item.Price,
			&item.Subtotal); err != nil {
			log.Printf("error while fetching values from each row in GetCheckoutItemsWithProductDetails method : %v", err)
			return nil, err
		}
		// Append each checkout item row to checkout items slice
		items = append(items, &item)
	}

	// IF there are which are not captured while iterating over the rows
	err = rows.Err()
	if err != nil {
		log.Printf("error captured while iterating over the rows : %v", err)
		return nil, err
	}

	return items, nil
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

/*
UpdateCheckoutItemCount:
- Update item_count value in checkout_sessions table
*/
func (r *checkoutRepository) UpdateCheckoutItemCount(ctx context.Context, checkoutID int64, itemCount int) error {
	query := `UPDATE checkout_sessions SET item_count = $1 WHERE id = $2`
	_, err := r.db.ExecContext(ctx, query, itemCount, checkoutID)
	if err != nil {
		log.Printf("error while updating item_count in checkout_sesions table : %v", err)
	}
	return err
}
