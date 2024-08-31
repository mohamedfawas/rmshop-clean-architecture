package usecase

import (
	"context"

	"github.com/mohamedfawas/rmshop-clean-architecture/internal/domain"
	"github.com/mohamedfawas/rmshop-clean-architecture/internal/repository"
	"github.com/mohamedfawas/rmshop-clean-architecture/pkg/utils"
)

type CartUseCase interface {
	AddToCart(ctx context.Context, userID, productID int64, quantity int) (*domain.CartItem, error)
}

type cartUseCase struct {
	cartRepo    repository.CartRepository
	productRepo repository.ProductRepository
}

func NewCartUseCase(cartRepo repository.CartRepository, productRepo repository.ProductRepository) CartUseCase {
	return &cartUseCase{
		cartRepo:    cartRepo,
		productRepo: productRepo,
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
