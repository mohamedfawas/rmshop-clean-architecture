package usecase

import (
	"context"
	"fmt"
	"log"
	"mime/multipart"
	"path/filepath"
	"time"

	"github.com/mohamedfawas/rmshop-clean-architecture/internal/domain"
	"github.com/mohamedfawas/rmshop-clean-architecture/internal/repository"
	"github.com/mohamedfawas/rmshop-clean-architecture/pkg/cloudinary"
	"github.com/mohamedfawas/rmshop-clean-architecture/pkg/utils"
)

type ProductUseCase interface {
	CreateProduct(ctx context.Context, product *domain.Product) error
	UpdateProduct(ctx context.Context, product *domain.Product) error
	GetProductByID(ctx context.Context, id int64) (*domain.Product, error)
	SoftDeleteProduct(ctx context.Context, id int64) error
	AddImage(ctx context.Context, productID int64, file multipart.File, fileHeader *multipart.FileHeader, isPrimary bool) error
	DeleteImage(ctx context.Context, productID int64, imageURL string) error
}

type productUseCase struct {
	productRepo     repository.ProductRepository
	subCategoryRepo repository.SubCategoryRepository
	cloudinary      *cloudinary.CloudinaryService
}

func NewProductUseCase(productRepo repository.ProductRepository, subCategoryRepo repository.SubCategoryRepository, cloudinary *cloudinary.CloudinaryService) ProductUseCase {
	return &productUseCase{
		productRepo:     productRepo,
		subCategoryRepo: subCategoryRepo,
		cloudinary:      cloudinary,
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
func (u *productUseCase) GetProductByID(ctx context.Context, id int64) (*domain.Product, error) {
	return u.productRepo.GetByID(ctx, id)
}

func (u *productUseCase) SoftDeleteProduct(ctx context.Context, id int64) error {
	return u.productRepo.SoftDelete(ctx, id)
}

func (u *productUseCase) UpdateProduct(ctx context.Context, product *domain.Product) error {
	existingProduct, err := u.productRepo.GetByID(ctx, product.ID)
	if err != nil {
		log.Printf("Error while retrieving product details using product ID, %v: ", err)
		return err
	}

	// Check if the subcategory exists and is not deleted
	if product.SubCategoryID != existingProduct.SubCategoryID {
		subCategory, err := u.subCategoryRepo.GetByID(ctx, product.SubCategoryID)
		if err != nil {
			return utils.ErrInvalidSubCategory
		}
		if subCategory.DeletedAt != nil {
			return utils.ErrInvalidSubCategory
		}
	}

	// Generate new slug if name has changed
	if product.Name != existingProduct.Name {
		product.Slug = utils.GenerateSlug(product.Name)
	} else {
		product.Slug = existingProduct.Slug
	}

	// Check for duplicate name only if the name has changed
	if product.Name != existingProduct.Name {
		exists, err := u.productRepo.NameExistsBeforeUpdate(ctx, product.Name, product.ID)
		if err != nil {
			return err
		}
		if exists {
			return utils.ErrDuplicateProductName
		}
	}

	// Update the product
	err = u.productRepo.Update(ctx, product)
	if err != nil {
		log.Printf("Error while updating product details, %v: ", err)
		return err
	}

	return nil
}

func (u *productUseCase) AddImage(ctx context.Context, productID int64, file multipart.File, fileHeader *multipart.FileHeader, isPrimary bool) error {
	// Check if the product exists
	product, err := u.productRepo.GetByID(ctx, productID)
	if err != nil {
		if err == utils.ErrProductNotFound {
			return utils.ErrProductNotFound
		}
		return err
	}

	// Check file size
	if fileHeader.Size > utils.MaxFileSize {
		return utils.ErrFileTooLarge
	}

	// Check file type
	if !isValidImageType(fileHeader.Filename) {
		return utils.ErrInvalidFileType
	}

	// Check the number of existing images
	currentCount, err := u.productRepo.GetImageCount(ctx, productID)
	if err != nil {
		return err
	}

	if currentCount >= utils.MaxImagesPerProduct {
		return utils.ErrTooManyImages
	}

	// Upload image to Cloudinary
	imageURL, err := u.cloudinary.UploadImage(ctx, file, product.Slug)
	if err != nil {
		return err
	}

	// Add image to database
	err = u.productRepo.AddImage(ctx, productID, imageURL, isPrimary)
	if err != nil {
		// If there's an error adding to the database, we should delete the image from Cloudinary
		_ = u.cloudinary.DeleteImage(ctx, imageURL)
		return err
	}

	return nil
}

func (u *productUseCase) DeleteImage(ctx context.Context, productID int64, imageURL string) error {
	// Check if the image exists and if it's primary
	image, err := u.productRepo.GetImageByURL(ctx, productID, imageURL)
	if err != nil {
		if err == utils.ErrImageNotFound {
			return utils.ErrImageNotFound
		}
		return err
	}

	// Delete the image from Cloudinary
	err = u.cloudinary.DeleteImage(ctx, imageURL)
	if err != nil {
		log.Printf("Failed to delete image from Cloudinary: %v", err)
		return err
	}

	// Get the current image count before deletion
	currentCount, err := u.productRepo.GetImageCount(ctx, productID)
	if err != nil {
		return err
	}

	// Delete the image from the database
	err = u.productRepo.DeleteImage(ctx, productID, imageURL)
	if err != nil {
		return err
	}

	// If the deleted image was primary, or it was the last image, update the product
	if image.IsPrimary || currentCount == 1 {
		err = u.updateProductAfterImageDeletion(ctx, productID)
		if err != nil {
			return err
		}
	}

	return nil
}

func (u *productUseCase) updateProductAfterImageDeletion(ctx context.Context, productID int64) error {
	// Get remaining images
	images, err := u.productRepo.GetProductImages(ctx, productID)
	if err != nil {
		return err
	}

	if len(images) == 0 {
		// No images left, set primary_image_id to NULL
		return u.productRepo.UpdateProductPrimaryImage(ctx, productID, nil)
	} else {
		// Set the first remaining image as primary
		newPrimaryImage := images[0]
		err = u.productRepo.SetImageAsPrimary(ctx, productID, newPrimaryImage.ID)
		if err != nil {
			return err
		}
		return u.productRepo.UpdateProductPrimaryImage(ctx, productID, &newPrimaryImage.ID)
	}
}

func isValidImageType(filename string) bool {
	ext := filepath.Ext(filename)
	switch ext {
	case ".jpg", ".jpeg", ".png", ".gif":
		return true
	}
	return false
}
