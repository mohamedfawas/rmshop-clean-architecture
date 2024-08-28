package validator

import (
	"mime/multipart"
	"path/filepath"
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
	if product.StockQuantity == 0 || product.StockQuantity < 0 {
		return utils.ErrInvalidStockQuantity
	}

	// Validate subcategory IDs
	if product.SubCategoryID <= 0 {
		return utils.ErrInvalidSubCategoryID
	}

	return nil
}

func ValidateProductName(name string) error {
	if len(name) < 2 {
		return utils.ErrProductNameTooShort
	}
	if len(name) > 255 {
		return utils.ErrProductNameTooLong
	}
	return nil
}

func ValidateProductDescription(description string) error {
	if len(description) < 10 || len(description) > 5000 {
		return utils.ErrProductDescriptionRequired
	}
	return nil
}

func ValidateProductPrice(price float64) error {
	if price <= 0 {
		return utils.ErrInvalidProductPrice
	}
	return nil
}

func ValidateProductStockQuantity(quantity int) error {
	if quantity < 0 {
		return utils.ErrInvalidStockQuantity
	}
	return nil
}

func ValidateProductSubCategoryID(subCategoryID int) error {
	if subCategoryID <= 0 {
		return utils.ErrInvalidSubCategoryID
	}
	return nil
}

func ValidateFile(header *multipart.FileHeader) error {
	// Check file size
	if header.Size > utils.MaxFileSize {
		return utils.ErrFileTooLarge
	}

	// Check file type
	if !IsValidImageType(header.Filename) {
		return utils.ErrInvalidFileType
	}

	// Check for empty file
	if header.Size == 0 {
		return utils.ErrEmptyFile
	}

	return nil
}

func IsValidImageType(filename string) bool {
	ext := filepath.Ext(filename)
	switch ext {
	case ".jpg", ".jpeg", ".png", ".gif":
		return true
	}
	return false
}
