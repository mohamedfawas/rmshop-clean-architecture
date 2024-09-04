package usecase

import (
	"context"

	"github.com/mohamedfawas/rmshop-clean-architecture/internal/domain"
	"github.com/mohamedfawas/rmshop-clean-architecture/internal/repository"
	"github.com/mohamedfawas/rmshop-clean-architecture/pkg/utils"
)

type InventoryUseCase interface {
	GetInventory(ctx context.Context, params domain.InventoryQueryParams) ([]*domain.InventoryItem, int64, error)
	UpdateProductStock(ctx context.Context, productID int64, quantity int) error
}

type inventoryUseCase struct {
	inventoryRepo repository.InventoryRepository
	productRepo   repository.ProductRepository
}

func NewInventoryUseCase(inventoryRepo repository.InventoryRepository, productRepo repository.ProductRepository) InventoryUseCase {
	return &inventoryUseCase{
		inventoryRepo: inventoryRepo,
		productRepo:   productRepo}
}

func (u *inventoryUseCase) GetInventory(ctx context.Context, params domain.InventoryQueryParams) ([]*domain.InventoryItem, int64, error) {
	// Validate and set default values
	if params.Page < 1 {
		params.Page = 1
	}
	if params.Limit < 1 {
		params.Limit = 10
	} else if params.Limit > 100 {
		params.Limit = 100
	}

	// Validate sorting parameters
	validSortFields := map[string]bool{"product_name": true, "category_name": true, "stock_quantity": true, "price": true}
	if params.SortBy != "" && !validSortFields[params.SortBy] {
		params.SortBy = "product_name"
	}
	if params.SortOrder != "asc" && params.SortOrder != "desc" {
		params.SortOrder = "asc"
	}

	return u.inventoryRepo.GetInventory(ctx, params)
}

func (u *inventoryUseCase) UpdateProductStock(ctx context.Context, productID int64, quantity int) error {
	if quantity < 0 {
		return utils.ErrInvalidStockQuantity
	}

	// You might want to add an upper limit check here if needed
	if quantity > 1000000 { // Example limit
		return utils.ErrStockQuantityTooLarge
	}

	return u.productRepo.UpdateStockQuantity(ctx, productID, quantity)
}
