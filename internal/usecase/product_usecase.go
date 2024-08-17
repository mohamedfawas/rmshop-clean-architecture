package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/mohamedfawas/rmshop-clean-architecture/internal/domain"
	"github.com/mohamedfawas/rmshop-clean-architecture/internal/repository"
)

type ProductUseCase interface {
	CreateProduct(ctx context.Context, product *domain.Product) error
	GetAllProducts(ctx context.Context) ([]*domain.Product, error)
	GetProductByID(ctx context.Context, id int64) (*domain.Product, error)
	UpdateProduct(ctx context.Context, product *domain.Product) error
	SoftDeleteProduct(ctx context.Context, id int64) error
	GetActiveProducts(ctx context.Context, page, pageSize int) ([]*domain.Product, int, error)
	CreateProductWithImages(ctx context.Context, product *domain.Product, images []domain.ProductImage) error
	UpdatePrimaryImage(ctx context.Context, productID int64, imageID int64) error
	AddProductImages(ctx context.Context, productID int64, images []domain.ProductImage) error
	RemoveProductImage(ctx context.Context, productID, imageID int64) error
}

type productUseCase struct {
	productRepo     repository.ProductRepository
	categoryRepo    repository.CategoryRepository
	subCategoryRepo repository.SubCategoryRepository
}

func NewProductUseCase(productRepo repository.ProductRepository, categoryRepo repository.CategoryRepository, subCategoryRepo repository.SubCategoryRepository) ProductUseCase {
	return &productUseCase{
		productRepo:     productRepo,
		categoryRepo:    categoryRepo,
		subCategoryRepo: subCategoryRepo,
	}
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

func (u *productUseCase) GetAllProducts(ctx context.Context) ([]*domain.Product, error) {
	return u.productRepo.GetAll(ctx)
}

func (u *productUseCase) GetProductByID(ctx context.Context, id int64) (*domain.Product, error) {
	return u.productRepo.GetByID(ctx, id)
}

func (u *productUseCase) UpdateProduct(ctx context.Context, product *domain.Product) error {
	// Perform any necessary validations here
	// For example, check if the category and subcategory exist, validate price, etc.

	// Update the product
	product.UpdatedAt = time.Now()
	return u.productRepo.Update(ctx, product)
}

func (u *productUseCase) SoftDeleteProduct(ctx context.Context, id int64) error {
	return u.productRepo.SoftDelete(ctx, id)
}

func (u *productUseCase) GetActiveProducts(ctx context.Context, page, pageSize int) ([]*domain.Product, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20 // Default page size
	}
	return u.productRepo.GetActiveProducts(ctx, page, pageSize)
}

// func (u *productUseCase) CreateProductWithImages(ctx context.Context, product *domain.Product, images []domain.ProductImage) error {
// 	// Start a transaction
// 	tx, err := u.productRepo.BeginTx(ctx)
// 	if err != nil {
// 		return err
// 	}
// 	defer tx.Rollback()

// 	// Create the product
// 	err = u.productRepo.CreateWithTx(ctx, tx, product)
// 	if err != nil {
// 		return err
// 	}

// 	// Add images and set primary image
// 	imageIDs, err := u.productRepo.AddProductImagesWithTx(ctx, tx, product.ID, images)
// 	if err != nil {
// 		return err
// 	}

// 	// Find the primary image
// 	var primaryImageID int64
// 	for i, img := range images {
// 		if img.IsPrimary {
// 			primaryImageID = imageIDs[i]
// 			break
// 		}
// 	}

// 	// If no primary image was specified, use the first image
// 	if primaryImageID == 0 && len(imageIDs) > 0 {
// 		primaryImageID = imageIDs[0]
// 	}

// 	if primaryImageID != 0 {
// 		err = u.productRepo.UpdatePrimaryImageWithTx(ctx, tx, product.ID, primaryImageID)
// 		if err != nil {
// 			return err
// 		}
// 	}

// 	// Commit the transaction
// 	return tx.Commit()
// }

func (u *productUseCase) UpdatePrimaryImage(ctx context.Context, productID int64, imageID int64) error {
	// First, check if the product exists
	product, err := u.productRepo.GetByID(ctx, productID)
	if err != nil {
		return err
	}

	// Check if the image belongs to the product
	imageExists := false
	for _, img := range product.Images {
		if img.ID == imageID {
			imageExists = true
			break
		}
	}
	if !imageExists {
		return errors.New("image does not belong to the product")
	}

	// Update the primary image
	return u.productRepo.UpdatePrimaryImage(ctx, productID, imageID)
}

func (u *productUseCase) AddProductImages(ctx context.Context, productID int64, images []domain.ProductImage) error {
	// Check if the product exists
	product, err := u.productRepo.GetByID(ctx, productID)
	if err != nil {
		return err
	}
	if product == nil {
		return errors.New("product not found")
	}

	// Add the new images
	_, err = u.productRepo.AddProductImages(ctx, productID, images)
	if err != nil {
		return err
	}

	return nil
}

func (u *productUseCase) RemoveProductImage(ctx context.Context, productID, imageID int64) error {
	// Check if the product exists
	product, err := u.productRepo.GetByID(ctx, productID)
	if err != nil {
		return err
	}
	if product == nil {
		return errors.New("product not found")
	}

	// Check if the image exists and belongs to the product
	image, err := u.productRepo.GetProductImageByID(ctx, imageID)
	if err != nil {
		return err
	}
	if image == nil || image.ProductID != productID {
		return errors.New("image not found")
	}

	// Check if it's the primary image
	if product.PrimaryImageID != nil && *product.PrimaryImageID == imageID {
		return errors.New("cannot delete primary image")
	}

	// Remove the image
	err = u.productRepo.RemoveProductImage(ctx, imageID)
	if err != nil {
		return err
	}

	return nil
}

func (u *productUseCase) CreateProductWithImages(ctx context.Context, product *domain.Product, images []domain.ProductImage) error {
	// Validate category and subcategory
	_, err := u.categoryRepo.GetByID(ctx, product.CategoryID)
	if err != nil {
		return ErrInvalidCategory
	}

	_, err = u.subCategoryRepo.GetByID(ctx, product.SubCategoryID)
	if err != nil {
		return ErrInvalidSubCategory
	}

	// Start a transaction
	tx, err := u.productRepo.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Set creation and update times
	now := time.Now()
	product.CreatedAt = now
	product.UpdatedAt = now

	// Create the product
	err = u.productRepo.CreateWithTx(ctx, tx, product)
	if err != nil {
		return err
	}

	// Add images
	imageIDs, err := u.productRepo.AddProductImagesWithTx(ctx, tx, product.ID, images)
	if err != nil {
		return err
	}

	// Set primary image
	var primaryImageID int64
	for i, img := range images {
		if img.IsPrimary {
			primaryImageID = imageIDs[i]
			break
		}
	}

	if primaryImageID != 0 {
		err = u.productRepo.UpdatePrimaryImageWithTx(ctx, tx, product.ID, primaryImageID)
		if err != nil {
			return err
		}
	}

	// Commit the transaction
	return tx.Commit()
}
