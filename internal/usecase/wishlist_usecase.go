package usecase

import (
	"context"

	"github.com/mohamedfawas/rmshop-clean-architecture/internal/domain"
	"github.com/mohamedfawas/rmshop-clean-architecture/internal/repository"
	"github.com/mohamedfawas/rmshop-clean-architecture/pkg/utils"
)

type WishlistUseCase interface {
	AddToWishlist(ctx context.Context, userID, productID int64) (*domain.WishlistItem, error)
}

type wishlistUseCase struct {
	wishlistRepo repository.WishlistRepository
	productRepo  repository.ProductRepository
}

func NewWishlistUseCase(wishlistRepo repository.WishlistRepository, productRepo repository.ProductRepository) WishlistUseCase {
	return &wishlistUseCase{
		wishlistRepo: wishlistRepo,
		productRepo:  productRepo,
	}
}

func (u *wishlistUseCase) AddToWishlist(ctx context.Context, userID, productID int64) (*domain.WishlistItem, error) {
	// Check if the product exists
	product, err := u.productRepo.GetByID(ctx, productID)
	if err != nil {
		if err == utils.ErrProductNotFound {
			return nil, utils.ErrProductNotFound
		}
		return nil, err
	}

	// Check if the item is already in the wishlist
	exists, err := u.wishlistRepo.ItemExists(ctx, userID, productID)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, utils.ErrDuplicateWishlistItem
	}

	// Check if the wishlist is full
	count, err := u.wishlistRepo.GetWishlistItemCount(ctx, userID)
	if err != nil {
		return nil, err
	}
	if count >= 50 { // Assuming a maximum of 50 items in the wishlist
		return nil, utils.ErrWishlistFull
	}

	// Add the item to the wishlist
	wishlistItem := &domain.WishlistItem{
		UserID:      userID,
		ProductID:   productID,
		IsAvailable: product.StockQuantity > 0,
		Price:       product.Price,
		ProductName: product.Name,
	}
	err = u.wishlistRepo.AddItem(ctx, wishlistItem)
	if err != nil {
		return nil, err
	}

	return wishlistItem, nil
}
