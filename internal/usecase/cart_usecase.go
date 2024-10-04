package usecase

import (
	"context"
	"log"
	"time"

	"github.com/mohamedfawas/rmshop-clean-architecture/internal/domain"
	"github.com/mohamedfawas/rmshop-clean-architecture/internal/repository"
	"github.com/mohamedfawas/rmshop-clean-architecture/pkg/utils"
)

type CartUseCase interface {
	AddToCart(ctx context.Context, userID, productID int64, quantity int) (*domain.CartItem, error)
	GetUserCart(ctx context.Context, userID int64) (*domain.CartResponse, error)
	UpdateCartItemQuantity(ctx context.Context, userID, itemID int64, quantity int) (*domain.CartItem, error)
	DeleteCartItem(ctx context.Context, userID, itemID int64) error
	ClearCart(ctx context.Context, userID int64) error
}

type cartUseCase struct {
	cartRepo    repository.CartRepository
	productRepo repository.ProductRepository
	userRepo    repository.UserRepository
}

func NewCartUseCase(cartRepo repository.CartRepository, productRepo repository.ProductRepository, userRepo repository.UserRepository) CartUseCase {
	return &cartUseCase{
		cartRepo:    cartRepo,
		productRepo: productRepo,
		userRepo:    userRepo,
	}
}

/*
AddToCart:
- Validate the given quantity
- Get product details , verify it's not soft deleted, stock is available
- Check if the given product is already added in the cart
- If already exists in cart, then update quantity (make sure it's not exceeding max limit)
- If product added is new to cart, then make new cart item entry
*/
func (u *cartUseCase) AddToCart(ctx context.Context, userID, productID int64, quantity int) (*domain.CartItem, error) {
	// Validate quantity
	if quantity <= 0 {
		return nil, utils.ErrInvalidQuantity
	}

	// If quantity is more than max allowed quantity for each product
	if quantity > utils.MaxCartItemQuantity {
		return nil, utils.ErrExceedsMaxQuantity
	}

	// Check if product exists and is active (not soft deleted)
	product, err := u.productRepo.GetByID(ctx, productID)
	if err != nil {
		if err == utils.ErrProductNotFound {
			return nil, utils.ErrProductNotFound
		}
		log.Printf("error while getting product details using ID : %v", err)
		return nil, err
	}

	// Check stock availability
	if product.StockQuantity < quantity {
		return nil, utils.ErrInsufficientStock
	}

	// Calculate subtotal
	subtotal := product.Price * float64(quantity)

	// Check if item already exists in cart
	existingItem, err := u.cartRepo.GetCartItemByProductID(ctx, userID, productID)
	// If any error other than cart item not found happens
	if err != nil && err != utils.ErrCartItemNotFound {
		log.Printf("error while checking if the product already exists in the cart : %v", err)
		return nil, err
	}

	if existingItem != nil {
		quantityAfterUpdation := existingItem.Quantity + quantity
		// Check if quantity after updation exceeds maximum limit
		if quantityAfterUpdation > utils.MaxCartItemQuantity {
			return nil, utils.ErrExceedsMaxQuantity
		}

		// If not violates max limit , Update quantity of existing item
		existingItem.Quantity += quantity
		existingItem.Price = product.Price // Update price in case it has changed
		existingItem.Subtotal = float64(existingItem.Quantity) * product.Price
		existingItem.UpdatedAt = time.Now().UTC()

		// Update cart item details
		err = u.cartRepo.UpdateCartItem(ctx, existingItem)
		if err != nil {
			log.Printf("error while updating cart item quantity for existing cart item : %v", err)
			return nil, err
		}
		return existingItem, nil
	}

	// Define the new cart item details
	now := time.Now().UTC()
	newItem := &domain.CartItem{
		UserID:    userID,
		ProductID: productID,
		Quantity:  quantity,
		Price:     product.Price,
		Subtotal:  subtotal,
		CreatedAt: now,
		UpdatedAt: now,
	}

	// Create the cart item entry in cart_items table
	err = u.cartRepo.AddCartItem(ctx, newItem)
	if err != nil {
		log.Printf("error while adding cart item entry in cart_items table : %v", err)
		return nil, err
	}

	return newItem, nil
}

func (u *cartUseCase) GetUserCart(ctx context.Context, userID int64) (*domain.CartResponse, error) {
	// Check if user exists and is not blocked
	user, err := u.userRepo.GetByID(ctx, userID)
	if err != nil {
		if err == utils.ErrUserNotFound {
			return nil, utils.ErrUserNotFound
		}
		return nil, err
	}
	if user.IsBlocked {
		return nil, utils.ErrUserBlocked
	}

	// Get cart items
	cartItems, err := u.cartRepo.GetCartByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Calculate total value
	var totalValue float64
	for _, item := range cartItems {
		totalValue += item.Subtotal
	}

	return &domain.CartResponse{
		Items:      cartItems,
		TotalValue: totalValue,
	}, nil
}

func (u *cartUseCase) UpdateCartItemQuantity(ctx context.Context, userID, itemID int64, quantity int) (*domain.CartItem, error) {
	if quantity <= 0 {
		return nil, utils.ErrInvalidQuantity
	}

	if quantity > utils.MaxCartItemQuantity {
		return nil, utils.ErrExceedsMaxQuantity
	}

	// Check if the cart item exists and belongs to the user
	existingItem, err := u.cartRepo.GetCartItemByID(ctx, itemID)
	if err != nil {
		if err == utils.ErrCartItemNotFound {
			return nil, utils.ErrCartItemNotFound
		}
		return nil, err
	}

	if existingItem.UserID != userID {
		return nil, utils.ErrUnauthorized
	}

	// Check product availability
	product, err := u.productRepo.GetByID(ctx, existingItem.ProductID)
	if err != nil {
		return nil, err
	}

	if product.StockQuantity < quantity {
		return nil, utils.ErrInsufficientStock
	}

	// Update quantity
	existingItem.Quantity = quantity
	existingItem.Subtotal = float64(quantity) * existingItem.Price
	existingItem.UpdatedAt = time.Now().UTC()

	err = u.cartRepo.UpdateCartItemQuantity(ctx, existingItem)
	if err != nil {
		log.Printf("failed to update cart item quantity : %v", err)
		return nil, err
	}

	// Update the quantity
	return existingItem, err
}

func (u *cartUseCase) DeleteCartItem(ctx context.Context, userID, itemID int64) error {
	// Check if the item exists in the user's cart
	cartItem, err := u.cartRepo.GetCartItemByID(ctx, itemID)
	if err != nil {
		if err == utils.ErrCartItemNotFound {
			return utils.ErrCartItemNotFound
		}
		log.Printf("error while retrieving cart item using item id : %v", err)
		return err
	}

	// Check if the item belongs to the user
	if cartItem.UserID != userID {
		return utils.ErrUnauthorized
	}

	// Delete the item
	err = u.cartRepo.DeleteCartItem(ctx, itemID)
	if err != nil {
		if err == utils.ErrCartItemNotFound {
			return utils.ErrCartItemNotFound
		}
		return err
	}

	return nil
}

func (u *cartUseCase) ClearCart(ctx context.Context, userID int64) error {
	return u.cartRepo.ClearCart(ctx, userID)
}
