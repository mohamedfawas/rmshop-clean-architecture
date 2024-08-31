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
