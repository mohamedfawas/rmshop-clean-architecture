package validator

import (
	"strings"

	"github.com/mohamedfawas/rmshop-clean-architecture/internal/domain"
	"github.com/mohamedfawas/rmshop-clean-architecture/pkg/utils"
)

const (
	MaxProductNameLength        = 255
	MinProductNameLength        = 2
	MaxProductDescriptionLength = 5000
	MinProductDescriptionLength = 10
	MaxImagesPerProduct         = 5
)

func ValidateProduct(product *domain.Product) error {
	// Validate product name
	if strings.TrimSpace(product.Name) == "" {
		return utils.ErrInvalidProductName
	}
	if len(product.Name) > MaxProductNameLength {
		return utils.ErrProductNameTooLong
	}

	if len(product.Name) < MinProductNameLength {
		return utils.ErrProductNameTooShort
	}

	// Validate product description
	if product.Description == "" {
		return utils.ErrProductDescriptionRequired
	}

	if len(product.Description) > MaxProductDescriptionLength || len(product.Description) < MinProductDescriptionLength {
		return utils.ErrInvalidProductDescription // product desc min_length<desc<max_length
	}

	// Validate price
	if product.Price <= 0 {
		return utils.ErrInvalidProductPrice
	}

	// Validate stock quantity
	if product.StockQuantity == nil || *product.StockQuantity < 0 {
		return utils.ErrInvalidStockQuantity
	}

	// Validate subcategory IDs
	if product.SubCategoryID <= 0 {
		return utils.ErrInvalidSubCategoryID
	}

	return nil
}
