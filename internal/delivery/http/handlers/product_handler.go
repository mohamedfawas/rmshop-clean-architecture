package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/mohamedfawas/rmshop-clean-architecture/internal/domain"
	"github.com/mohamedfawas/rmshop-clean-architecture/internal/usecase"
	"github.com/mohamedfawas/rmshop-clean-architecture/pkg/api"
	"github.com/mohamedfawas/rmshop-clean-architecture/pkg/utils"
	"github.com/mohamedfawas/rmshop-clean-architecture/pkg/validator"
)

type ProductHandler struct {
	productUseCase usecase.ProductUseCase
}

func NewProductHandler(productUseCase usecase.ProductUseCase) *ProductHandler {
	return &ProductHandler{productUseCase: productUseCase}
}

func (h *ProductHandler) CreateProduct(w http.ResponseWriter, r *http.Request) {
	// Parse the request body
	var product domain.Product
	err := json.NewDecoder(r.Body).Decode(&product)
	if err != nil {
		api.SendResponse(w, http.StatusBadRequest, "Invalid request", nil, "Failed to parse request body")
		return
	}

	product.Name = strings.ToLower(strings.TrimSpace(product.Name))
	product.Description = strings.TrimSpace(product.Description)

	err = validator.ValidateProduct(&product)
	if err != nil {
		errorMessages := map[error]string{
			utils.ErrInvalidProductName:         "Please provide a valid product name",
			utils.ErrProductNameTooLong:         "Product name should not be greater than 255 characters",
			utils.ErrProductNameTooShort:        "Product name should have at least 2 characters",
			utils.ErrProductDescriptionRequired: "Please provide a valid product description",
			utils.ErrInvalidProductDescription:  "Product description should have at least 2 characters and should be less than 5000 characters",
			utils.ErrInvalidProductPrice:        "Please provide a valid product price",
			utils.ErrInvalidStockQuantity:       "Please provide a valid stock quantity",
			utils.ErrInvalidSubCategoryID:       "Please provide a valid sub category ID",
		}
		if errMsg, exists := errorMessages[err]; exists {
			api.SendResponse(w, http.StatusBadRequest, "Validation failed", nil, errMsg)
		} else {
			api.SendResponse(w, http.StatusInternalServerError, "Validation failed", nil, "Internal server error")
		}
		return
	}

	// Create the product
	err = h.productUseCase.CreateProduct(r.Context(), &product)
	if err != nil {
		switch err {
		case utils.ErrInvalidSubCategory:
			api.SendResponse(w, http.StatusBadRequest, "Failed to create product", nil, "The specified sub-category is invalid or deleted")
		case utils.ErrDuplicateProductName:
			api.SendResponse(w, http.StatusConflict, "Failed to create product", nil, "A product with this name already exists")
		default:
			api.SendResponse(w, http.StatusInternalServerError, "Failed to create product", nil, "An unexpected error occurred")
		}
		return
	}

	// Product created successfully
	api.SendResponse(w, http.StatusCreated, "Product created successfully", product, "")

}
