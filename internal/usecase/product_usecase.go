package usecase

import (
	"context"
	"fmt"
	"log"
	"mime/multipart"
	"strings"
	"time"

	"github.com/mohamedfawas/rmshop-clean-architecture/internal/domain"
	"github.com/mohamedfawas/rmshop-clean-architecture/internal/repository"
	"github.com/mohamedfawas/rmshop-clean-architecture/pkg/cloudinary"
	"github.com/mohamedfawas/rmshop-clean-architecture/pkg/utils"
	"github.com/mohamedfawas/rmshop-clean-architecture/pkg/validator"
)

type ProductUseCase interface {
	CreateProduct(ctx context.Context, product *domain.Product) error
	UpdateProduct(ctx context.Context, product *domain.Product) error
	GetProductByID(ctx context.Context, id int64) (*domain.Product, error)
	SoftDeleteProduct(ctx context.Context, id int64) error
	AddImage(ctx context.Context, productID int64, files []multipart.File, fileKeys []string, fileHeaders []multipart.FileHeader, isPrimary bool) error
	DeleteImage(ctx context.Context, productID int64, imageURL string) error
	AddImages(ctx context.Context, productID int64, files []multipart.File, headers []*multipart.FileHeader, isPrimaryFlags []bool) error
	DeleteProductImage(ctx context.Context, productID, imageID int64) error
	GetAllProducts(ctx context.Context) ([]*domain.Product, error)
	GetProducts(ctx context.Context, params domain.ProductQueryParams) ([]*domain.Product, int64, error)
	GetPublicProductByID(ctx context.Context, id int64) (*domain.PublicProduct, error)
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
	product, err := u.productRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// if the product is soft deleted
	if product.DeletedAt != nil {
		return nil, utils.ErrProductNotFound
	}

	return product, nil
}

func (u *productUseCase) SoftDeleteProduct(ctx context.Context, id int64) error {
	// check if the product exists
	_, err := u.productRepo.GetByID(ctx, id)
	if err != nil {
		if err == utils.ErrProductNotFound {
			return utils.ErrProductNotFound
		}
		log.Printf("Error retrieving product : %v", err)
		return err
	}

	// perform soft delete
	err = u.productRepo.SoftDelete(ctx, id)
	if err != nil {
		return err
	}

	return nil
}

func (u *productUseCase) UpdateProduct(ctx context.Context, product *domain.Product) error {
	existingProduct, err := u.productRepo.GetByID(ctx, product.ID)
	if err != nil {
		log.Printf("Error while retrieving product details using product ID, %v: ", err)
		return err
	}

	// Check if the subcategory exists and is not soft deleted
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
		exists, err := u.productRepo.NameExistsBeforeUpdate(ctx, product.Name, product.ID) //check whether any active product names exist with the updated name
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

func (u *productUseCase) AddImage(ctx context.Context, productID int64, files []multipart.File, fileKeys []string, fileHeaders []multipart.FileHeader, isPrimary bool) error {
	// Check if the product exists
	product, err := u.productRepo.GetByID(ctx, productID)
	if err != nil {
		if err == utils.ErrProductNotFound {
			return utils.ErrProductNotFound
		}
		return err
	}

	// Check the number of existing images
	currentCount, err := u.productRepo.GetImageCount(ctx, productID)
	if err != nil {
		return err
	}

	if currentCount+len(files) > utils.MaxImagesPerProduct {
		return utils.ErrTooManyImages
	}

	if isPrimary && len(files) > 1 {
		return utils.ErrMultiplePrimaryImages
	}

	for i, file := range files {
		// Check file size
		if fileHeaders[i].Size > utils.MaxFileSize {
			return utils.ErrFileTooLarge
		}

		// Check file type
		if !validator.IsValidImageType(fileHeaders[i].Filename) {
			return utils.ErrInvalidFileType
		}

		// Check for empty file
		if fileHeaders[i].Size == 0 {
			return utils.ErrEmptyFile
		}

		// Upload image to Cloudinary
		imageURL, err := u.cloudinary.UploadImage(ctx, file, product.Slug)
		if err != nil {
			return err
		}

		// If this is a primary image, check if there's an existing primary image
		if isPrimary && fileKeys[i] == "image_primary" {
			existingPrimary, err := u.productRepo.GetPrimaryImage(ctx, productID)
			if err != nil {
				return err
			}
			if existingPrimary != nil {
				// Update the existing primary image to non-primary
				err = u.productRepo.UpdateImagePrimary(ctx, existingPrimary.ID, false)
				if err != nil {
					return err
				}
			}
		}

		// Add image to database
		err = u.productRepo.AddImage(ctx, productID, imageURL, isPrimary && fileKeys[i] == "image_primary")
		if err != nil {
			// If there's an error adding to the database, we should delete the image from Cloudinary
			_ = u.cloudinary.DeleteImage(ctx, imageURL)
			return err
		}
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

func (u *productUseCase) AddImages(ctx context.Context, productID int64, files []multipart.File, headers []*multipart.FileHeader, isPrimaryFlags []bool) error {
	// Check if the product exists
	product, err := u.productRepo.GetByID(ctx, productID)
	if err != nil {
		return err
	}

	// Check the number of existing images
	currentCount, err := u.productRepo.GetImageCount(ctx, productID)
	if err != nil {
		return err
	}

	if currentCount+len(files) > utils.MaxImagesPerProduct {
		return utils.ErrTooManyImages
	}

	for i, file := range files {
		// Perform file validations (size, type, etc.)
		if err := validator.ValidateFile(headers[i]); err != nil {
			log.Printf("error while validating file : %v", err)
			return err
		}

		// Upload image to Cloudinary
		imageURL, err := u.cloudinary.UploadImage(ctx, file, product.Slug)
		if err != nil {
			log.Printf("error while uploading file to cloudinary: %v", err)
			return err
		}

		// Add image to database
		err = u.productRepo.AddImage(ctx, productID, imageURL, isPrimaryFlags[i])
		if err != nil {
			// If there's an error adding to the database, we should delete the image from Cloudinary
			_ = u.cloudinary.DeleteImage(ctx, imageURL)
			return err
		}
	}

	return nil
}

func (u *productUseCase) DeleteProductImage(ctx context.Context, productID, imageID int64) error {
	// check if the product exists
	_, err := u.productRepo.GetByID(ctx, productID)
	if err != nil {
		if err == utils.ErrProductNotFound {
			return utils.ErrProductNotFound
		}
		return err
	}

	// check if the image exists and belongs to the product
	image, err := u.productRepo.GetImageByID(ctx, imageID)
	if err != nil {
		if err == utils.ErrImageNotFound {
			return utils.ErrImageNotFound
		}
		return err
	}

	if image.ProductID != productID {
		return utils.ErrImageNotFound
	}

	// Check if it's the last image
	imageCount, err := u.productRepo.GetImageCount(ctx, productID)
	if err != nil {
		return err
	}
	if imageCount == 1 {
		return utils.ErrLastImage
	}

	// Delete the image from cloudinary
	err = u.cloudinary.DeleteImage(ctx, image.ImageURL)
	if err != nil {
		log.Printf("Failed to delete image from Cloudinary : %v", err)
	}
	// Delete the image from the database
	err = u.productRepo.DeleteImageByID(ctx, imageID)
	if err != nil {
		return err
	}

	// If the deleted image was the primary image, set a new primary image
	if image.IsPrimary {
		err = u.setNewPrimaryImage(ctx, productID)
		if err != nil {
			return err
		}
	}

	return nil
}

func (u *productUseCase) setNewPrimaryImage(ctx context.Context, productID int64) error {
	images, err := u.productRepo.GetProductImages(ctx, productID)
	if err != nil {
		return err
	}
	if len(images) > 0 {
		return u.productRepo.SetImageAsPrimary(ctx, productID, images[0].ID)
	}
	return nil
}

func (u *productUseCase) GetAllProducts(ctx context.Context) ([]*domain.Product, error) {
	return u.productRepo.GetAll(ctx)
}

func (u *productUseCase) GetProducts(ctx context.Context, params domain.ProductQueryParams) ([]*domain.Product, int64, error) {
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
	validSortFields := map[string]bool{"price": true, "name": true, "created_at": true, "updated_at": true}
	if params.Sort != "" && !validSortFields[params.Sort] {
		params.Sort = "created_at"
	}
	if params.Order != "asc" && params.Order != "desc" {
		params.Order = "desc"
	}

	// Convert all string parameters to lowercase for case-insensitive search
	params.Category = strings.ToLower(params.Category)
	params.Subcategory = strings.ToLower(params.Subcategory)
	params.Search = strings.ToLower(params.Search)

	for i, category := range params.Categories {
		params.Categories[i] = strings.ToLower(category)
	}

	// Call repository method
	return u.productRepo.GetProducts(ctx, params)
}

func (u *productUseCase) GetPublicProductByID(ctx context.Context, id int64) (*domain.PublicProduct, error) {
	product, err := u.productRepo.GetPublicProductByID(ctx, id)
	if err != nil {
		if err == utils.ErrProductNotFound {
			return nil, utils.ErrProductNotFound
		}
		return nil, fmt.Errorf("failed to retrieve product: %w", err)
	}

	return product, nil
}
