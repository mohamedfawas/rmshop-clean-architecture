package usecase

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/mohamedfawas/rmshop-clean-architecture/internal/domain"
	"github.com/mohamedfawas/rmshop-clean-architecture/internal/repository"
	"github.com/mohamedfawas/rmshop-clean-architecture/pkg/utils"
)

type ProductUseCase interface {
	CreateProduct(ctx context.Context, product *domain.Product) error
}

type productUseCase struct {
	productRepo     repository.ProductRepository
	subCategoryRepo repository.SubCategoryRepository
}

func NewProductUseCase(productRepo repository.ProductRepository, subCategoryRepo repository.SubCategoryRepository) ProductUseCase {
	return &productUseCase{
		productRepo:     productRepo,
		subCategoryRepo: subCategoryRepo,
	}
}

func (u *productUseCase) CreateProduct(ctx context.Context, product *domain.Product) error {
	// Check if product name already exists
	exists, err := u.productRepo.NameExists(ctx, product.Name)
	if err != nil {
		log.Printf("Failed to check if product name exists: %v", err)
		return err
	}
	if exists {
		return utils.ErrDuplicateProductName
	}

	// retrieve subcategory details : sub categories that are not soft deleted are retrieved
	subCategory, err := u.subCategoryRepo.GetByID(ctx, product.SubCategoryID)
	if err != nil {
		if err == utils.ErrSubCategoryNotFound {
			return utils.ErrInvalidSubCategory
		}
		log.Printf("Failed to retrieve sub category details : %v", err)
		return err
	}

	// Ensure the subcategory is not deleted
	if subCategory.DeletedAt != nil {
		return utils.ErrInvalidSubCategory
	}

	// Generate slug
	slug := fmt.Sprintf("%s/%s", subCategory.Slug,
		utils.GenerateSlug(product.Name))

	// Ensure slug uniqueness
	uniqueSlug, err := u.ensureUniqueSlug(ctx, slug)
	if err != nil {
		log.Printf("Failed to ensure slug uniqueness: %v", err)
		return err
	}
	product.Slug = uniqueSlug

	// Set creation and update times
	now := time.Now()
	product.CreatedAt = now
	product.UpdatedAt = now

	// Create the product
	err = u.productRepo.Create(ctx, product)
	return nil
}

func (u *productUseCase) ensureUniqueSlug(ctx context.Context, slug string) (string, error) {
	uniqueSlug := slug
	suffix := 1
	for {
		exists, err := u.productRepo.SlugExists(ctx, uniqueSlug)
		if err != nil {
			log.Printf("error while checking if the product's 'slug' exists: %v", err)
			return "", err
		}
		if !exists {
			return uniqueSlug, nil
		}
		//If the slug does exist, it adds a number to the end of the slug and tries again.
		uniqueSlug = fmt.Sprintf("%s-%d", slug, suffix)
		suffix++
	}
}
