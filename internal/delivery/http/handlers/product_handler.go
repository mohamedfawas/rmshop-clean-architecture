package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
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

func (h *ProductHandler) UpdateProduct(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	productID, err := strconv.ParseInt(vars["productId"], 10, 64)
	if err != nil {
		api.SendResponse(w, http.StatusBadRequest, "Invalid product ID", nil, "Product ID must be a number")
		return
	}

	// Get the existing product
	existingProduct, err := h.productUseCase.GetProductByID(r.Context(), productID)
	if err != nil {
		if err == utils.ErrProductNotFound {
			api.SendResponse(w, http.StatusNotFound, "Failed to update product", nil, "Product not found")
		} else {
			api.SendResponse(w, http.StatusInternalServerError, "Failed to update product", nil, "An unexpected error occurred")
		}
		return
	}

	// Decode the request body into a map
	var updateFields map[string]interface{} // interface can take any data type
	err = json.NewDecoder(r.Body).Decode(&updateFields)
	if err != nil {
		api.SendResponse(w, http.StatusBadRequest, "Invalid request", nil, "Failed to parse request body")
		return
	}

	// If the request body is empty, return success without making any changes
	if len(updateFields) == 0 {
		api.SendResponse(w, http.StatusOK, "No changes made to the product", existingProduct, "")
		return
	}

	// Update and validate only the provided fields
	updatedProduct := *existingProduct
	for key, value := range updateFields {
		switch key {
		case "name":
			if name, ok := value.(string); ok {
				updatedProduct.Name = strings.ToLower(strings.TrimSpace(name))
				if err := validator.ValidateProductName(updatedProduct.Name); err != nil {
					switch err {
					case utils.ErrProductNameTooShort:
						api.SendResponse(w, http.StatusBadRequest, "Validation failed", nil, "Product name should have atleast 2 characters")
					case utils.ErrProductNameTooLong:
						api.SendResponse(w, http.StatusBadRequest, "Validation failed", nil, "Product name should not have more than 255 characters")
					}
					return
				}
			} else {
				api.SendResponse(w, http.StatusBadRequest, "Invalid field", nil, "Name must be a string")
				return
			}
		case "description":
			if description, ok := value.(string); ok {
				updatedProduct.Description = strings.TrimSpace(description)
				if err := validator.ValidateProductDescription(updatedProduct.Description); err != nil {
					api.SendResponse(w, http.StatusBadRequest, "Validation failed", nil, "Product description should be between 10 and 5000 characters")
					return
				}
			} else {
				api.SendResponse(w, http.StatusBadRequest, "Invalid field", nil, "Description must be a string")
				return
			}
		case "price":
			if price, ok := value.(float64); ok {
				updatedProduct.Price = price
				if err := validator.ValidateProductPrice(updatedProduct.Price); err != nil {
					api.SendResponse(w, http.StatusBadRequest, "Validation failed", nil, err.Error())
					return
				}
			} else {
				api.SendResponse(w, http.StatusBadRequest, "Invalid field", nil, "Price must be a number")
				return
			}
		case "stock_quantity":
			if quantity, ok := value.(float64); ok {
				updatedProduct.StockQuantity = int(quantity)
				if err := validator.ValidateProductStockQuantity(updatedProduct.StockQuantity); err != nil {
					api.SendResponse(w, http.StatusBadRequest, "Validation failed", nil, err.Error())
					return
				}
			} else {
				api.SendResponse(w, http.StatusBadRequest, "Invalid field", nil, "Stock quantity must be a number")
				return
			}
		case "sub_category_id":
			if subCategoryID, ok := value.(float64); ok {
				updatedProduct.SubCategoryID = int(subCategoryID)
				if err := validator.ValidateProductSubCategoryID(updatedProduct.SubCategoryID); err != nil {
					api.SendResponse(w, http.StatusBadRequest, "Validation failed", nil, err.Error())
					return
				}
			} else {
				api.SendResponse(w, http.StatusBadRequest, "Invalid field", nil, "Sub-category ID must be a number")
				return
			}
		default:
			// Ignore unknown fields
			continue
		}
	}

	// Update the product
	err = h.productUseCase.UpdateProduct(r.Context(), &updatedProduct)
	if err != nil {
		switch err {
		case utils.ErrProductNotFound:
			api.SendResponse(w, http.StatusNotFound, "Failed to update product", nil, "Product not found")
		case utils.ErrInvalidSubCategory:
			api.SendResponse(w, http.StatusBadRequest, "Failed to update product", nil, "Invalid sub-category")
		case utils.ErrDuplicateProductName:
			api.SendResponse(w, http.StatusConflict, "Failed to update product", nil, "Product name already exists")
		default:
			api.SendResponse(w, http.StatusInternalServerError, "Failed to update product", nil, "An unexpected error occurred")
		}
		return
	}

	api.SendResponse(w, http.StatusOK, "Product updated successfully", updatedProduct, "")

}

func (h *ProductHandler) AddProductImage(w http.ResponseWriter, r *http.Request) {
	// Parse the multipart form
	err := r.ParseMultipartForm(10 << 20) // 10 MB max , 10 << 20 is a bitwise shift operation that calculates 10 * 2^20, which equals approximately 10 MB
	if err != nil {
		api.SendResponse(w, http.StatusBadRequest, "Failed to parse form", nil, "Invalid form data")
		return
	}

	// Get the product ID from the URL
	vars := mux.Vars(r) // extracts the variables (path parameters) from the URL
	//The Vars function returns a map where the keys are the parameter names defined in the URL route, and the values are the corresponding segments in the actual request URL.
	//if your route is /products/{productId}, vars["productId"] will contain the value passed in the URL for productId.
	productID, err := strconv.ParseInt(vars["productId"], 10, 64)
	if err != nil {
		api.SendResponse(w, http.StatusBadRequest, "Invalid product ID", nil, "Product ID must be a number")
		return
	}

	// Get the file from the form data
	file, header, err := r.FormFile("image")
	if err != nil {
		api.SendResponse(w, http.StatusBadRequest, "Failed to get image", nil, "No image file provided")
		return
	}
	defer file.Close()

	// Check if the image should be set as primary
	isPrimaryStr := r.FormValue("is_primary")
	isPrimary := isPrimaryStr == "true"

	// Call the use case method to add the image
	err = h.productUseCase.AddImage(r.Context(), productID, file, header, isPrimary)
	if err != nil {
		switch err {
		case utils.ErrProductNotFound:
			api.SendResponse(w, http.StatusNotFound, "Failed to add image", nil, "Product not found")
		case utils.ErrFileTooLarge:
			api.SendResponse(w, http.StatusBadRequest, "Failed to add image", nil, "Image file is too large")
		case utils.ErrInvalidFileType:
			api.SendResponse(w, http.StatusBadRequest, "Failed to add image", nil, "Invalid image file type")
		case utils.ErrTooManyImages:
			api.SendResponse(w, http.StatusBadRequest, "Failed to add image", nil, "Maximum number of images reached for this product")
		case utils.ErrDuplicateImageURL:
			api.SendResponse(w, http.StatusConflict, "Failed to add image", nil, "This image already exists for the product")
		default:
			api.SendResponse(w, http.StatusInternalServerError, "Failed to add image", nil, "An unexpected error occurred")
		}
		return
	}

	api.SendResponse(w, http.StatusCreated, "Image added successfully", nil, "")
}
