package postgres

import (
	"context"
	"database/sql"
	"log"
	"time"

	"github.com/mohamedfawas/rmshop-clean-architecture/internal/domain"
	"github.com/mohamedfawas/rmshop-clean-architecture/pkg/utils"
)

type cartRepository struct {
	db *sql.DB
}

func NewCartRepository(db *sql.DB) *cartRepository {
	return &cartRepository{db: db}
}

/*
AddCartItem:
- Add cart item details to cart_items table
- user_id, product_id, quantity, created_at, updated_at
*/
func (r *cartRepository) AddCartItem(ctx context.Context, item *domain.CartItem) error {
	query := `
		INSERT INTO cart_items (user_id, product_id, quantity, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`

	err := r.db.QueryRowContext(ctx, query,
		item.UserID,
		item.ProductID,
		item.Quantity,
		item.CreatedAt,
		item.UpdatedAt).Scan(&item.ID)
	if err != nil {
		log.Printf("error while adding cart item entry to cart_items table : %v", err)
	}
	return err
}

/*
GetCartItemByProductID:
- Get cart item details from cart_items table using user id and product id
*/
func (r *cartRepository) GetCartItemByProductID(ctx context.Context, userID, productID int64) (*domain.CartItem, error) {
	query := `
		SELECT id, user_id, product_id, quantity, created_at, updated_at
		FROM cart_items
		WHERE user_id = $1 AND product_id = $2
	`
	var item domain.CartItem
	err := r.db.QueryRowContext(ctx, query, userID, productID).Scan(
		&item.ID,
		&item.UserID,
		&item.ProductID,
		&item.Quantity,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, utils.ErrCartItemNotFound
		}
		log.Printf("eror while retrieving cart item details using user id and product id : %v", err)
		return nil, err
	}

	// return the retrieved cart item details
	return &item, nil
}

/*
UpdateCartItem:
- Update quantity of the given cart item, update in cart_items table
*/
func (r *cartRepository) UpdateCartItem(ctx context.Context, item *domain.CartItem) error {
	query := `
		UPDATE cart_items
		SET quantity = $1, updated_at = $2
		WHERE id = $3
	`
	_, err := r.db.ExecContext(ctx, query, item.Quantity, time.Now(), item.ID)
	if err != nil {
		log.Printf("error while updating quantity of the given cart item : %v", err)
	}
	return err
}

func (r *cartRepository) GetCartByUserID(ctx context.Context, userID int64) ([]*domain.CartItemWithProduct, error) {
	query := `
		SELECT ci.id, ci.user_id, ci.product_id, ci.quantity, ci.created_at, ci.updated_at,
			   p.name, p.price
		FROM cart_items ci
		JOIN products p ON ci.product_id = p.id
		WHERE ci.user_id = $1 AND p.deleted_at IS NULL
		ORDER BY ci.created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		log.Printf("error while retrieving cart using userID : %v", err)
		return nil, err
	}
	defer rows.Close()

	var cartItems []*domain.CartItemWithProduct
	for rows.Next() {
		var ci domain.CartItemWithProduct
		err := rows.Scan(
			&ci.ID, &ci.UserID, &ci.ProductID, &ci.Quantity, &ci.CreatedAt, &ci.UpdatedAt,
			&ci.ProductName, &ci.ProductPrice,
		)
		if err != nil {
			log.Printf("db error : %v", err)
			return nil, err
		}
		cartItems = append(cartItems, &ci)
	}

	if err = rows.Err(); err != nil {
		log.Printf("db error : %v", err)
		return nil, err
	}

	return cartItems, nil
}

func (r *cartRepository) UpdateCartItemQuantity(ctx context.Context, userID, itemID int64, quantity int) error {
	query := `UPDATE cart_items
				SET quantity=$1, updated_at = NOW()
				WHERE id=$2 AND user_id=$3`
	result, err := r.db.ExecContext(ctx, query, quantity, itemID, userID)
	if err != nil {
		log.Printf("error while updating the cart_items : %v", err)
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("erorr while checking the rows affected : %v", err)
		return err
	}

	if rowsAffected == 0 {
		return utils.ErrCartItemNotFound
	}

	return nil
}

func (r *cartRepository) DeleteCartItem(ctx context.Context, itemID int64) error {
	query := `
        DELETE FROM cart_items
        WHERE id = $1
    `
	result, err := r.db.ExecContext(ctx, query, itemID)
	if err != nil {
		log.Printf("error while deleting the cart item : %v", err)
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("error while checking the rows affected : %v", err)
		return err
	}

	if rowsAffected == 0 {
		return utils.ErrCartItemNotFound
	}

	return nil
}

// GetCartItemByID
// - Retrieve cart item details from cart_items table
func (r *cartRepository) GetCartItemByID(ctx context.Context, itemID int64) (*domain.CartItem, error) {
	query := `
        SELECT id, user_id, product_id, quantity, created_at, updated_at
        FROM cart_items
        WHERE id = $1
    `
	var item domain.CartItem
	err := r.db.QueryRowContext(ctx, query, itemID).Scan(
		&item.ID,
		&item.UserID,
		&item.ProductID,
		&item.Quantity,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, utils.ErrCartItemNotFound
	}
	if err != nil {
		log.Printf("error while retrieiving cart item details using cart item using id: %v", err)
		return nil, err
	}
	return &item, nil
}

func (r *cartRepository) GetCartTotal(ctx context.Context, userID int64) (float64, error) {
	// COALESCE ensures that the result is 0 instead of NULL
	query := `SELECT COALESCE(SUM(ci.quantity * p.price), 0) 
              FROM cart_items ci 
              JOIN products p ON ci.product_id = p.id 
              WHERE ci.user_id = $1`

	var total float64
	err := r.db.QueryRowContext(ctx, query, userID).Scan(&total)
	if err != nil {
		log.Printf("error while getting cart total : %v", err)
		return 0, err
	}
	return total, nil
}

func (r *cartRepository) ApplyCoupon(ctx context.Context, userID int64, coupon *domain.Coupon) error {
	query := `INSERT INTO applied_coupons (user_id, coupon_id) VALUES ($1, $2)
              ON CONFLICT (user_id) DO UPDATE SET coupon_id = $2`

	_, err := r.db.ExecContext(ctx, query, userID, coupon.ID)
	if err != nil {
		log.Printf("error while adding to applied coupons : %v", err)
	}
	return err
}

func (r *cartRepository) RemoveCoupon(ctx context.Context, userID int64) error {
	query := `DELETE FROM applied_coupons WHERE user_id = $1`

	_, err := r.db.ExecContext(ctx, query, userID)
	if err != nil {
		log.Printf("error while removing applied coupon : %v", err)
	}
	return err
}

func (r *cartRepository) GetAppliedCoupon(ctx context.Context, userID int64) (*domain.Coupon, error) {
	query := `SELECT c.id, c.code, c.discount_percentage, c.min_order_amount, c.is_active, c.created_at, c.updated_at, c.expires_at
              FROM applied_coupons ac
              JOIN coupons c ON ac.coupon_id = c.id
              WHERE ac.user_id = $1`

	var coupon domain.Coupon
	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&coupon.ID, &coupon.Code, &coupon.DiscountPercentage, &coupon.MinOrderAmount,
		&coupon.IsActive, &coupon.CreatedAt, &coupon.UpdatedAt, &coupon.ExpiresAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		log.Printf("error while retrieving applied coupon details : %v", err)
		return nil, err
	}

	return &coupon, nil
}

/*
ClearCart:
- Delete the cart item entry from cart_items table
*/
func (r *cartRepository) ClearCart(ctx context.Context, userID int64) error {
	query := `DELETE FROM cart_items WHERE user_id = $1`
	_, err := r.db.ExecContext(ctx, query, userID)
	if err != nil {
		log.Printf("error while clearing cart item from the cart_items table : %v", err)
	}
	return err
}
