package usecase

import (
	"context"
	"time"

	"github.com/mohamedfawas/rmshop-clean-architecture/internal/domain"
	"github.com/mohamedfawas/rmshop-clean-architecture/internal/repository"
)

type ProductUseCase interface {
	CreateProduct(ctx context.Context, product *domain.Product) error
}

type productUseCase struct {
	productRepo repository.ProductRepository
}

func NewProductUseCase(productRepo repository.ProductRepository) ProductUseCase {
	return &productUseCase{productRepo: productRepo}
}

func (u *productUseCase) CreateProduct(ctx context.Context, product *domain.Product) error {
	// Set creation and update times
	now := time.Now()
	product.CreatedAt = now
	product.UpdatedAt = now

	// Perform any necessary validations here
	// For example, check if the category and subcategory exist, validate price, etc.

	// Create the product
	return u.productRepo.Create(ctx, product)
}
