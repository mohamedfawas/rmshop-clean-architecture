package usecase

import (
	"context"

	"github.com/mohamedfawas/rmshop-clean-architecture/internal/domain"
	"github.com/mohamedfawas/rmshop-clean-architecture/internal/repository"
	"github.com/mohamedfawas/rmshop-clean-architecture/pkg/utils"
)

type WishlistUseCase interface {
	AddToWishlist(ctx context.Context, userID, productID int64) (*domain.WishlistItem, error)
	RemoveFromWishlist(ctx context.Context, userID, productID int64) (bool, error)
	GetUserWishlist(ctx context.Context, userID int64, page, limit int, sortBy, order string) ([]*domain.WishlistItem, int64, error)
}

type wishlistUseCase struct {
	wishlistRepo repository.WishlistRepository
	productRepo  repository.ProductRepository
	userRepo     repository.UserRepository
}

func NewWishlistUseCase(wishlistRepo repository.WishlistRepository, productRepo repository.ProductRepository, userRepo repository.UserRepository) WishlistUseCase {
	return &wishlistUseCase{
		wishlistRepo: wishlistRepo,
		productRepo:  productRepo,
		userRepo:     userRepo,
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

func (u *wishlistUseCase) RemoveFromWishlist(ctx context.Context, userID, productID int64) (bool, error) {
	// Check if the item exists in the user's wishlist
	exists, err := u.wishlistRepo.ItemExists(ctx, userID, productID)
	if err != nil {
		return false, err
	}
	if !exists {
		return false, utils.ErrProductNotInWishlist
	}

	// Remove the item
	err = u.wishlistRepo.RemoveItem(ctx, userID, productID)
	if err != nil {
		return false, err
	}

	// Check if the wishlist is now empty
	count, err := u.wishlistRepo.GetWishlistItemCount(ctx, userID)
	if err != nil {
		return false, err
	}

	return count == 0, nil
}

func (u *wishlistUseCase) GetUserWishlist(ctx context.Context, userID int64, page, limit int, sortBy, order string) ([]*domain.WishlistItem, int64, error) {
	// Check if user exists
	_, err := u.userRepo.GetByID(ctx, userID)
	if err != nil {
		if err == utils.ErrUserNotFound {
			return nil, 0, utils.ErrUserNotFound
		}
		return nil, 0, err
	}

	// Validate and set default sorting options
	if sortBy == "" {
		sortBy = "created_at"
	}
	if order != "asc" && order != "desc" {
		order = "desc"
	}

	// Get wishlist items
	items, totalCount, err := u.wishlistRepo.GetUserWishlistItems(ctx, userID, page, limit, sortBy, order)
	if err != nil {
		return nil, 0, err
	}

	return items, totalCount, nil
}
