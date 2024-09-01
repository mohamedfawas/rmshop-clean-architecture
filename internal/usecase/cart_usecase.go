package usecase

import (
	"context"

	"github.com/mohamedfawas/rmshop-clean-architecture/internal/domain"
	"github.com/mohamedfawas/rmshop-clean-architecture/internal/repository"
	"github.com/mohamedfawas/rmshop-clean-architecture/pkg/utils"
)

type CartUseCase interface {
	AddToCart(ctx context.Context, userID, productID int64, quantity int) (*domain.CartItem, error)
	GetUserCart(ctx context.Context, userID int64) (*domain.Cart, error)
	UpdateCartItemQuantity(ctx context.Context, userID, itemID int64, quantity int) error
	DeleteCartItem(ctx context.Context, userID, itemID int64) error
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

func (u *cartUseCase) AddToCart(ctx context.Context, userID, productID int64, quantity int) (*domain.CartItem, error) {
	// Validate quantity
	if quantity <= 0 {
		return nil, utils.ErrInvalidQuantity
	}

	if quantity > utils.MaxCartItemQuantity {
		return nil, utils.ErrExceedsMaxQuantity
	}

	// Check if product exists and is active
	product, err := u.productRepo.GetByID(ctx, productID)
	if err != nil {
		return nil, err
	}
	if product == nil || product.DeletedAt != nil {
		return nil, utils.ErrProductNotFound
	}

	// Check stock availability
	if product.StockQuantity < quantity {
		return nil, utils.ErrInsufficientStock
	}

	// Check if item already exists in cart
	existingItem, err := u.cartRepo.GetCartItemByProductID(ctx, userID, productID)
	if err != nil && err != utils.ErrCartItemNotFound {
		return nil, err
	}

	if existingItem != nil {
		// Check if quantity after updation exceeds maximum limit
		quantityAfterUpdation := existingItem.Quantity + quantity
		if quantityAfterUpdation > utils.MaxCartItemQuantity {
			return nil, utils.ErrExceedsMaxQuantity
		}

		// Update quantity of existing item
		existingItem.Quantity += quantity
		err = u.cartRepo.UpdateCartItem(ctx, existingItem)
		if err != nil {
			return nil, err
		}
		return existingItem, nil
	}

	// Add new item to cart
	newItem := &domain.CartItem{
		UserID:    userID,
		ProductID: productID,
		Quantity:  quantity,
	}

	err = u.cartRepo.AddCartItem(ctx, newItem)
	if err != nil {
		return nil, err
	}

	return newItem, nil
}

func (u *cartUseCase) GetUserCart(ctx context.Context, userID int64) (*domain.Cart, error) {
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
		totalValue += float64(item.Quantity) * item.ProductPrice
	}

	return &domain.Cart{
		Items:      cartItems,
		TotalValue: totalValue,
	}, nil
}

func (u *cartUseCase) UpdateCartItemQuantity(ctx context.Context, userID, itemID int64, quantity int) error {
	if quantity < 0 {
		return utils.ErrInvalidQuantity
	}

	if quantity > utils.MaxCartItemQuantity {
		return utils.ErrExceedsMaxQuantity
	}

	// Check if the cart item exists and belongs to the user
	existingItem, err := u.cartRepo.GetCartItemByID(ctx, itemID)
	if err != nil {
		if err == utils.ErrCartItemNotFound {
			return utils.ErrCartItemNotFound
		}
		return err
	}

	if existingItem.UserID != userID {
		return utils.ErrUnauthorized
	}

	// If quantity is 0, remove the item from the cart
	if quantity == 0 {
		return u.cartRepo.DeleteCartItem(ctx, itemID)
	}

	// Check product availability
	product, err := u.productRepo.GetByID(ctx, existingItem.ProductID)
	if err != nil {
		return err
	}

	if product.StockQuantity < quantity {
		return utils.ErrInsufficientStock
	}

	// Update the quantity
	return u.cartRepo.UpdateCartItemQuantity(ctx, userID, itemID, quantity)
}

func (u *cartUseCase) DeleteCartItem(ctx context.Context, userID, itemID int64) error {
	// Check if the item exists in the user's cart
	cartItem, err := u.cartRepo.GetCartItemByID(ctx, itemID)
	if err != nil {
		if err == utils.ErrCartItemNotFound {
			return utils.ErrCartItemNotFound
		}
		return err
	}

	// Check if the item belongs to the user
	if cartItem.UserID != userID {
		return utils.ErrUnauthorized
	}

	// Delete the item
	err = u.cartRepo.DeleteCartItem(ctx, itemID)
	if err != nil {
		return err
	}

	return nil
}
