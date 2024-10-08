package handlers

import (
	"encoding/json"
	"log"
	"mime/multipart"
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
	// Extract the product id from the URL
	vars := mux.Vars(r)
	productID, err := strconv.ParseInt(vars["productId"], 10, 64)
	if err != nil {
		api.SendResponse(w, http.StatusBadRequest, "Invalid product ID", nil, "Product ID must be a number")
		return
	}

	// Fetch the updated fields from the request body
	var updateFields map[string]interface{}
	err = json.NewDecoder(r.Body).Decode(&updateFields)
	if err != nil {
		api.SendResponse(w, http.StatusBadRequest, "Invalid request", nil, "Failed to parse request body")
		return
	}

	// Call the usecase method
	updatedProduct, err := h.productUseCase.UpdateProduct(r.Context(), productID, updateFields)
	if err != nil {
		switch err {
		case utils.ErrProductNameTooShort:
			api.SendResponse(w, http.StatusBadRequest, "Failed to update product", nil, "Product name should have more than 2 characters")
		case utils.ErrProductNameTooLong:
			api.SendResponse(w, http.StatusBadRequest, "Failed to update product", nil, "Product name should be less than 255 characters")
		case utils.ErrProductDescriptionRequired:
			api.SendResponse(w, http.StatusBadRequest, "Failed to update product", nil, "A valid product description is required")
		case utils.ErrInvalidProductPrice:
			api.SendResponse(w, http.StatusBadRequest, "Failed to update product", nil, "Please provide a valid product price")
		case utils.ErrInvalidStockQuantity:
			api.SendResponse(w, http.StatusBadRequest, "Failed to update product", nil, "Please provide a valid stock quantity for the product")
		case utils.ErrInvalidSubCategoryID:
			api.SendResponse(w, http.StatusBadRequest, "Failed to update product", nil, "Please provide a valid sub category id")
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

func (h *ProductHandler) SoftDeleteProduct(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	productID, err := strconv.ParseInt(vars["productId"], 10, 64) //base 10 int , converts to int64
	if err != nil {
		api.SendResponse(w, http.StatusBadRequest, "Invalid product ID", nil, "Product ID must be a number")
		return
	}

	err = h.productUseCase.SoftDeleteProduct(r.Context(), productID)
	if err != nil {
		switch err {
		case utils.ErrProductNotFound:
			api.SendResponse(w, http.StatusNotFound, "Failed to delete product", nil, "Product not found")
		default:
			api.SendResponse(w, http.StatusInternalServerError, "Internal server error", nil, "An unexpected error occured")
		}
		return
	}

	api.SendResponse(w, http.StatusOK, "Product deleted successfully", nil, "")
}

func (h *ProductHandler) GetProductByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	productID, err := strconv.ParseInt(vars["productId"], 10, 64)
	if err != nil {
		api.SendResponse(w, http.StatusBadRequest, "Invalid product ID", nil, "Product ID must be a number")
		return
	}

	product, err := h.productUseCase.GetProductByID(r.Context(), productID)
	if err != nil {
		switch err {
		case utils.ErrProductNotFound:
			api.SendResponse(w, http.StatusNotFound, "Product not found", nil, "The requested product does not exist or has been deleted")
		default:
			api.SendResponse(w, http.StatusInternalServerError, "Failed to retrieve product", nil, "An unexpected error occurred")
		}
		return
	}

	api.SendResponse(w, http.StatusOK, "Product retrieved successfully", product, "")

}

func (h *ProductHandler) AddProductImages(w http.ResponseWriter, r *http.Request) {
	// Parse the multipart form
	err := r.ParseMultipartForm(10 << 20) // 10 MB max
	if err != nil {
		api.SendResponse(w, http.StatusBadRequest, "Failed to parse form", nil, "Invalid form data")
		return
	}

	// Get the product ID from the URL
	vars := mux.Vars(r)
	productID, err := strconv.ParseInt(vars["productId"], 10, 64)
	if err != nil {
		api.SendResponse(w, http.StatusBadRequest, "Invalid product ID", nil, "Product ID must be a number")
		return
	}

	// Prepare slices to hold file information
	var files []multipart.File
	var headers []*multipart.FileHeader //holds the file's metadata, like filename and headers, and provides methods for accessing the actual file content.
	var isPrimaryFlags []bool

	// Handle primary image
	primaryFile, primaryHeader, err := r.FormFile("image_primary")
	if err == nil {
		files = append(files, primaryFile)
		headers = append(headers, primaryHeader)
		isPrimaryFlags = append(isPrimaryFlags, true)
	}

	// Handle non-primary images
	if multipartForm := r.MultipartForm; multipartForm != nil {
		if fileHeaders := multipartForm.File["image"]; len(fileHeaders) > 0 {
			for _, fileHeader := range fileHeaders {
				file, err := fileHeader.Open()
				if err != nil {
					api.SendResponse(w, http.StatusInternalServerError, "Failed to process image", nil, "Error opening file")
					return
				}
				defer file.Close()
				files = append(files, file)
				headers = append(headers, fileHeader)
				isPrimaryFlags = append(isPrimaryFlags, false)
			}
		}
	}

	if len(files) == 0 {
		api.SendResponse(w, http.StatusBadRequest, "Failed to get images", nil, "No image files provided")
		return
	}

	err = h.productUseCase.AddImages(r.Context(), productID, files, headers, isPrimaryFlags)
	if err != nil {
		handleImageUploadError(w, err)
		return
	}

	api.SendResponse(w, http.StatusCreated, "Images added successfully", nil, "")
}

func handleImageUploadError(w http.ResponseWriter, err error) {
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
	case utils.ErrEmptyFile:
		api.SendResponse(w, http.StatusBadRequest, "Failed to add image", nil, "Empty file provided")
	case utils.ErrMultiplePrimaryImages:
		api.SendResponse(w, http.StatusBadRequest, "Failed to add image", nil, "Only one primary image can be uploaded at a time")
	default:
		api.SendResponse(w, http.StatusInternalServerError, "Failed to add image", nil, "An unexpected error occurred")
	}
}

func (h *ProductHandler) DeleteProductImage(w http.ResponseWriter, r *http.Request) {
	// get the product id from the url
	vars := mux.Vars(r)
	productID, err := strconv.ParseInt(vars["productId"], 10, 64)
	if err != nil {
		api.SendResponse(w, http.StatusBadRequest, "Invalid product ID", nil, "Product ID must be a number")
		return
	}

	//get the image id from the url
	imageID, err := strconv.ParseInt(vars["imageId"], 10, 64)
	if err != nil {
		api.SendResponse(w, http.StatusBadRequest, "Invalid image ID", nil, "Image ID must be a number")
		return
	}

	err = h.productUseCase.DeleteProductImage(r.Context(), productID, imageID)
	if err != nil {
		switch err {
		case utils.ErrProductNotFound:
			api.SendResponse(w, http.StatusNotFound, "Failed to delete image", nil, "Product not found")
		case utils.ErrImageNotFound:
			api.SendResponse(w, http.StatusNotFound, "Failed to delete image", nil, "Image not found")
		case utils.ErrLastImage:
			api.SendResponse(w, http.StatusBadRequest, "Failed to delete image", nil, "Cannot delete the last image of a product")
		default:
			api.SendResponse(w, http.StatusInternalServerError, "Failed to delete image", nil, "An unexpected error occurred")
		}
		return
	}

	api.SendResponse(w, http.StatusOK, "Image deleted successfully", nil, "")
}

func (h *ProductHandler) GetAllProducts(w http.ResponseWriter, r *http.Request) {
	products, err := h.productUseCase.GetAllProducts(r.Context())
	if err != nil {
		api.SendResponse(w, http.StatusInternalServerError, "Failed to retrieve products", nil, "An unexpected error occurred")
		return
	}

	if len(products) == 0 {
		api.SendResponse(w, http.StatusOK, "No products found", []struct{}{}, "")
		return
	}

	api.SendResponse(w, http.StatusOK, "Products retrieved successfully", products, "")
}

func (h *ProductHandler) GetProducts(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	params := parseQueryParams(r)

	// Call use case
	products, totalCount, err := h.productUseCase.GetProducts(r.Context(), params)
	if err != nil {
		api.SendResponse(w, http.StatusInternalServerError, "Failed to retrieve products", nil, err.Error())
		return
	}

	response := map[string]interface{}{
		"products":    products,
		"total_count": totalCount,
		"page":        params.Page,
		"limit":       params.Limit,
		"total_pages": (totalCount + int64(params.Limit) - 1) / int64(params.Limit),
	}

	api.SendResponse(w, http.StatusOK, "Products retrieved successfully", response, "")
}

func parseQueryParams(r *http.Request) domain.ProductQueryParams {
	params := domain.ProductQueryParams{
		Page:  1,
		Limit: 10,
	}

	if page, err := strconv.Atoi(r.URL.Query().Get("page")); err == nil && page > 0 {
		params.Page = page
	}

	if limit, err := strconv.Atoi(r.URL.Query().Get("limit")); err == nil && limit > 0 {
		params.Limit = limit
	}

	params.Sort = r.URL.Query().Get("sort")
	params.Order = r.URL.Query().Get("order")
	params.Category = r.URL.Query().Get("category")
	params.Subcategory = r.URL.Query().Get("subcategory")
	params.Search = r.URL.Query().Get("search")

	params.MinPrice, _ = strconv.ParseFloat(r.URL.Query().Get("min_price"), 64)
	params.MaxPrice, _ = strconv.ParseFloat(r.URL.Query().Get("max_price"), 64)

	params.InStock, _ = strconv.ParseBool(r.URL.Query().Get("in_stock"))

	params.CreatedAfter = r.URL.Query().Get("created_after")
	params.CreatedBefore = r.URL.Query().Get("created_before")
	params.UpdatedAfter = r.URL.Query().Get("updated_after")
	params.UpdatedBefore = r.URL.Query().Get("updated_before")

	params.Categories = r.URL.Query()["category"]

	return params
}

func (h *ProductHandler) GetPublicProductByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	productID, err := strconv.ParseInt(vars["productId"], 10, 64)
	if err != nil {
		log.Printf("Invalid product ID: %v", err)
		api.SendResponse(w, http.StatusBadRequest, "Invalid product ID", nil, "Product ID must be a number")
		return
	}

	product, err := h.productUseCase.GetPublicProductByID(r.Context(), productID)
	if err != nil {
		switch err {
		case utils.ErrProductNotFound:
			api.SendResponse(w, http.StatusNotFound, "Product not found", nil, "The requested product does not exist or has been deleted")
		default:
			log.Printf("Error retrieving product: %v", err)
			api.SendResponse(w, http.StatusInternalServerError, "Failed to retrieve product", nil, "An unexpected error occurred")
		}
		return
	}

	api.SendResponse(w, http.StatusOK, "Product retrieved successfully", product, "")
}
