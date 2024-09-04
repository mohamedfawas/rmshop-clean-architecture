package usecase

import (
	"context"

	"github.com/mohamedfawas/rmshop-clean-architecture/internal/domain"
	"github.com/mohamedfawas/rmshop-clean-architecture/internal/repository"
)

type InventoryUseCase interface {
	GetInventory(ctx context.Context, params domain.InventoryQueryParams) ([]*domain.InventoryItem, int64, error)
}

type inventoryUseCase struct {
	inventoryRepo repository.InventoryRepository
}

func NewInventoryUseCase(inventoryRepo repository.InventoryRepository) InventoryUseCase {
	return &inventoryUseCase{inventoryRepo: inventoryRepo}
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
