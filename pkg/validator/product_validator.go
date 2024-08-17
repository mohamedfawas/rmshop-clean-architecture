package validator

import (
	"regexp"
	"strings"

	"github.com/mohamedfawas/rmshop-clean-architecture/internal/domain"
	"github.com/mohamedfawas/rmshop-clean-architecture/pkg/utils"
)

const (
	MaxProductNameLength        = 255
	MaxProductDescriptionLength = 5000
	MaxImagesPerProduct         = 10
)

func ValidateProduct(product *domain.Product, images []domain.ProductImage) error {
	// Validate product name
	if strings.TrimSpace(product.Name) == "" {
		return utils.ErrInvalidProductName
	}
	if len(product.Name) > MaxProductNameLength {
		return utils.ErrProductNameTooLong
	}

	// Validate product description
	if product.Description == "" {
		return utils.ErrProductDescriptionRequired
	}

	if len(product.Description) > MaxProductDescriptionLength {
		return utils.ErrInvalidProductDescription
	}

	// Validate price
	if product.Price <= 0 {
		return utils.ErrInvalidProductPrice
	}

	// Validate stock quantity
	if product.StockQuantity == nil {
		return utils.ErrStockQuantRequired
	}
	if *product.StockQuantity < 0 {
		return utils.ErrInvalidStockQuantity
	}

	// Validate category and subcategory IDs
	if product.CategoryID <= 0 {
		return utils.ErrInvalidCategoryID
	}
	if product.SubCategoryID <= 0 {
		return utils.ErrInvalidSubCategoryID
	}

	// Validate images
	if len(images) == 0 {
		return utils.ErrNoImages
	}
	if len(images) > MaxImagesPerProduct {
		return utils.ErrTooManyImages
	}

	primaryImageCount := 0
	urlRegex := regexp.MustCompile(`^https?://`)

	for _, img := range images {
		if !urlRegex.MatchString(img.ImageURL) {
			return utils.ErrInvalidImageURL
		}
		if img.IsPrimary {
			primaryImageCount++
		}
	}

	if primaryImageCount == 0 {
		return utils.ErrNoPrimaryImage
	}
	if primaryImageCount > 1 {
		return utils.ErrMultiplePrimaryImages
	}

	return nil
}
