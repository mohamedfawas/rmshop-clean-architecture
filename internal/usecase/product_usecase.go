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
	// retrieve subcategory details
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
	return u.productRepo.Create(ctx, product)
}

func (u *productUseCase) ensureUniqueSlug(ctx context.Context, slug string) (string, error) {
	uniqueSlug := slug
	suffix := 1
	for {
		exists, err := u.productRepo.SlugExists(ctx, uniqueSlug)
		if err != nil {
			log.Printf("error while checking if the product slug exists: %v", err)
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

// package usecase

// import (
// 	"context"
// 	"errors"
// 	"fmt"
// 	"time"

// 	"github.com/mohamedfawas/rmshop-clean-architecture/internal/domain"
// 	"github.com/mohamedfawas/rmshop-clean-architecture/internal/repository"
// 	"github.com/mohamedfawas/rmshop-clean-architecture/pkg/utils"
// )

// type ProductUseCase interface {
// 	CreateProduct(ctx context.Context, product *domain.Product) error
// 	// GetAllProducts(ctx context.Context) ([]*domain.Product, error)
// 	// GetProductByID(ctx context.Context, id int64) (*domain.Product, error)
// 	// UpdateProduct(ctx context.Context, product *domain.Product) error
// 	// SoftDeleteProduct(ctx context.Context, id int64) error
// 	// GetActiveProducts(ctx context.Context, page, pageSize int) ([]*domain.Product, int, error)
// 	// CreateProductWithImages(ctx context.Context, product *domain.Product, images []domain.ProductImage) error
// 	// UpdatePrimaryImage(ctx context.Context, productID int64, imageID int64) error
// 	// AddProductImages(ctx context.Context, productID int64, images []domain.ProductImage) error
// 	// RemoveProductImage(ctx context.Context, productID, imageID int64) error
// }

// type productUseCase struct {
// 	productRepo     repository.ProductRepository
// 	categoryRepo    repository.CategoryRepository
// 	subCategoryRepo repository.SubCategoryRepository
// }

// func NewProductUseCase(productRepo repository.ProductRepository, categoryRepo repository.CategoryRepository, subCategoryRepo repository.SubCategoryRepository) ProductUseCase {
// 	return &productUseCase{
// 		productRepo:     productRepo,
// 		categoryRepo:    categoryRepo,
// 		subCategoryRepo: subCategoryRepo,
// 	}
// }

// func (u *productUseCase) CreateProduct(ctx context.Context, product *domain.Product) error {
// 	// Validate subcategory
// 	subCategory, err := u.subCategoryRepo.GetByID(ctx, product.SubCategoryID)
// 	if err != nil {
// 		return utils.ErrInvalidSubCategory
// 	}

// 	// Ensure the subcategory is not deleted
// 	if subCategory.DeletedAt != nil {
// 		return utils.ErrInvalidSubCategory
// 	}

// 	// Generate slug for product
// 	slug := fmt.Sprintf("%s/%s", subCategory.Slug, utils.GenerateSlug(product.Name))
// 	// Ensure slug uniqueness
// 	uniqueSlug, err := u.ensureUniqueSlug(ctx, slug)
// 	if err != nil {
// 		return err
// 	}
// 	product.Slug = uniqueSlug

// 	// Set creation and update times
// 	now := time.Now()
// 	product.CreatedAt = now
// 	product.UpdatedAt = now

// 	// Create the product
// 	return u.productRepo.Create(ctx, product)
// }

// func (u *productUseCase) ensureUniqueSlug(ctx context.Context, slug string) (string, error) {
// 	uniqueSlug := slug
// 	suffix := 1
// 	for {
// 		exists, err := u.productRepo.SlugExists(ctx, uniqueSlug)
// 		if err != nil {
// 			return "", err
// 		}
// 		if !exists {
// 			return uniqueSlug, nil
// 		}
// 		uniqueSlug = fmt.Sprintf("%s-%d", slug, suffix)
// 		suffix++
// 	}
// }

// func (u *productUseCase) GetAllProducts(ctx context.Context) ([]*domain.Product, error) {
// 	return u.productRepo.GetAll(ctx)
// }

// func (u *productUseCase) GetProductByID(ctx context.Context, id int64) (*domain.Product, error) {
// 	return u.productRepo.GetByID(ctx, id)
// }

// func (u *productUseCase) UpdateProduct(ctx context.Context, product *domain.Product) error {
// 	// Perform any necessary validations here
// 	// For example, check if the category and subcategory exist, validate price, etc.

// 	// Update the product
// 	product.UpdatedAt = time.Now()
// 	return u.productRepo.Update(ctx, product)
// }

// func (u *productUseCase) SoftDeleteProduct(ctx context.Context, id int64) error {
// 	return u.productRepo.SoftDelete(ctx, id)
// }

// func (u *productUseCase) GetActiveProducts(ctx context.Context, page, pageSize int) ([]*domain.Product, int, error) {
// 	if page < 1 {
// 		page = 1
// 	}
// 	if pageSize < 1 || pageSize > 100 {
// 		pageSize = 20 // Default page size
// 	}
// 	return u.productRepo.GetActiveProducts(ctx, page, pageSize)
// }

// // func (u *productUseCase) CreateProductWithImages(ctx context.Context, product *domain.Product, images []domain.ProductImage) error {
// // 	// Start a transaction
// // 	tx, err := u.productRepo.BeginTx(ctx)
// // 	if err != nil {
// // 		return err
// // 	}
// // 	defer tx.Rollback()

// // 	// Create the product
// // 	err = u.productRepo.CreateWithTx(ctx, tx, product)
// // 	if err != nil {
// // 		return err
// // 	}

// // 	// Add images and set primary image
// // 	imageIDs, err := u.productRepo.AddProductImagesWithTx(ctx, tx, product.ID, images)
// // 	if err != nil {
// // 		return err
// // 	}

// // 	// Find the primary image
// // 	var primaryImageID int64
// // 	for i, img := range images {
// // 		if img.IsPrimary {
// // 			primaryImageID = imageIDs[i]
// // 			break
// // 		}
// // 	}

// // 	// If no primary image was specified, use the first image
// // 	if primaryImageID == 0 && len(imageIDs) > 0 {
// // 		primaryImageID = imageIDs[0]
// // 	}

// // 	if primaryImageID != 0 {
// // 		err = u.productRepo.UpdatePrimaryImageWithTx(ctx, tx, product.ID, primaryImageID)
// // 		if err != nil {
// // 			return err
// // 		}
// // 	}

// // 	// Commit the transaction
// // 	return tx.Commit()
// // }

// func (u *productUseCase) UpdatePrimaryImage(ctx context.Context, productID int64, imageID int64) error {
// 	// First, check if the product exists
// 	product, err := u.productRepo.GetByID(ctx, productID)
// 	if err != nil {
// 		return err
// 	}

// 	// Check if the image belongs to the product
// 	imageExists := false
// 	for _, img := range product.Images {
// 		if img.ID == imageID {
// 			imageExists = true
// 			break
// 		}
// 	}
// 	if !imageExists {
// 		return errors.New("image does not belong to the product")
// 	}

// 	// Update the primary image
// 	return u.productRepo.UpdatePrimaryImage(ctx, productID, imageID)
// }

// func (u *productUseCase) AddProductImages(ctx context.Context, productID int64, images []domain.ProductImage) error {
// 	// Check if the product exists
// 	product, err := u.productRepo.GetByID(ctx, productID)
// 	if err != nil {
// 		return err
// 	}
// 	if product == nil {
// 		return ErrProductNotFound
// 	}

// 	// Check for duplicate URLs
// 	existingURLs := make(map[string]bool)
// 	for _, img := range product.Images {
// 		existingURLs[img.ImageURL] = true
// 	}
// 	for _, newImg := range images {
// 		if existingURLs[newImg.ImageURL] {
// 			return ErrDuplicateImageURL
// 		}
// 		existingURLs[newImg.ImageURL] = true
// 	}

// 	// Add the new images
// 	_, err = u.productRepo.AddProductImages(ctx, productID, images)
// 	if err != nil {
// 		return err
// 	}

// 	return nil
// }

// func (u *productUseCase) RemoveProductImage(ctx context.Context, productID, imageID int64) error {
// 	// Check if the product exists
// 	product, err := u.productRepo.GetByID(ctx, productID)
// 	if err != nil {
// 		return err
// 	}
// 	if product == nil {
// 		return errors.New("product not found")
// 	}

// 	// Check if the image exists and belongs to the product
// 	image, err := u.productRepo.GetProductImageByID(ctx, imageID)
// 	if err != nil {
// 		return err
// 	}
// 	if image == nil || image.ProductID != productID {
// 		return errors.New("image not found")
// 	}

// 	// Check if it's the primary image
// 	if product.PrimaryImageID != nil && *product.PrimaryImageID == imageID {
// 		return errors.New("cannot delete primary image")
// 	}

// 	// Remove the image
// 	err = u.productRepo.RemoveProductImage(ctx, imageID)
// 	if err != nil {
// 		return err
// 	}

// 	return nil
// }

// func (u *productUseCase) CreateProductWithImages(ctx context.Context, product *domain.Product, images []domain.ProductImage) error {
// 	// Validate category and subcategory
// 	_, err := u.categoryRepo.GetByID(ctx, product.CategoryID)
// 	if err != nil {
// 		return ErrInvalidCategory
// 	}

// 	_, err = u.subCategoryRepo.GetByID(ctx, product.SubCategoryID)
// 	if err != nil {
// 		return ErrInvalidSubCategory
// 	}

// 	// Start a transaction
// 	tx, err := u.productRepo.BeginTx(ctx)
// 	if err != nil {
// 		return err
// 	}
// 	defer tx.Rollback()

// 	// Set creation and update times
// 	now := time.Now()
// 	product.CreatedAt = now
// 	product.UpdatedAt = now

// 	// Create the product
// 	err = u.productRepo.CreateWithTx(ctx, tx, product)
// 	if err != nil {
// 		return err
// 	}

// 	// Add images
// 	imageIDs, err := u.productRepo.AddProductImagesWithTx(ctx, tx, product.ID, images)
// 	if err != nil {
// 		return err
// 	}

// 	// Set primary image
// 	var primaryImageID int64
// 	for i, img := range images {
// 		if img.IsPrimary {
// 			primaryImageID = imageIDs[i]
// 			break
// 		}
// 	}

// 	if primaryImageID != 0 {
// 		err = u.productRepo.UpdatePrimaryImageWithTx(ctx, tx, product.ID, primaryImageID)
// 		if err != nil {
// 			return err
// 		}
// 		product.PrimaryImageID = &primaryImageID
// 	}

// 	// Commit the transaction
// 	if err = tx.Commit(); err != nil {
// 		return err
// 	}

// 	// Fetch the images for the product
// 	productImages, err := u.productRepo.GetProductImages(ctx, product.ID)
// 	if err != nil {
// 		return err
// 	}

// 	product.Images = productImages

// 	return nil
// }
