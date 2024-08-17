package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"regexp"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/mohamedfawas/rmshop-clean-architecture/internal/domain"
	"github.com/mohamedfawas/rmshop-clean-architecture/internal/usecase"
	"github.com/mohamedfawas/rmshop-clean-architecture/pkg/utils"
	"github.com/mohamedfawas/rmshop-clean-architecture/pkg/validator"
)

type ProductHandler struct {
	productUseCase usecase.ProductUseCase
}

func NewProductHandler(productUseCase usecase.ProductUseCase) *ProductHandler {
	return &ProductHandler{productUseCase: productUseCase}
}

func (h *ProductHandler) GetAllProducts(w http.ResponseWriter, r *http.Request) {
	log.Println("Entering GetAllProducts handler")

	products, err := h.productUseCase.GetAllProducts(r.Context())
	if err != nil {
		log.Printf("Error retrieving products: %v", err)
		http.Error(w, "Failed to retrieve products", http.StatusInternalServerError)
		return
	}

	log.Printf("Retrieved %d products", len(products))

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(products)
	if err != nil {
		log.Printf("Error encoding products to JSON: %v", err)
		http.Error(w, "Failed to encode products", http.StatusInternalServerError)
		return
	}

	log.Println("Successfully sent products response")
}

func (h *ProductHandler) GetProductByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	productID, err := strconv.ParseInt(vars["productId"], 10, 64)
	if err != nil {
		http.Error(w, "Invalid product ID", http.StatusBadRequest)
		return
	}

	product, err := h.productUseCase.GetProductByID(r.Context(), productID)
	if err != nil {
		if err.Error() == "product not found" {
			http.Error(w, "Product not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to retrieve product", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(product)
}

func (h *ProductHandler) UpdateProduct(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	productID, err := strconv.ParseInt(vars["productId"], 10, 64)
	if err != nil {
		http.Error(w, "Invalid product ID", http.StatusBadRequest)
		return
	}

	var product domain.Product
	err = json.NewDecoder(r.Body).Decode(&product)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	product.ID = productID

	err = h.productUseCase.UpdateProduct(r.Context(), &product)
	if err != nil {
		if err.Error() == "product not found or already deleted" {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, "Failed to update product", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(product)
}

func (h *ProductHandler) SoftDeleteProduct(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	productID, err := strconv.ParseInt(vars["productId"], 10, 64)
	if err != nil {
		http.Error(w, "Invalid product ID", http.StatusBadRequest)
		return
	}

	err = h.productUseCase.SoftDeleteProduct(r.Context(), productID)
	if err != nil {
		if err.Error() == "product not found or already deleted" {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, "Failed to delete product", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *ProductHandler) GetActiveProducts(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("user_id").(int64)
	if !ok {
		log.Printf("Error retrieving user ID from context")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	page, err := strconv.Atoi(r.URL.Query().Get("page"))
	if err != nil || page < 1 {
		page = 1
	}
	pageSize, err := strconv.Atoi(r.URL.Query().Get("pageSize"))
	if err != nil || pageSize < 1 {
		pageSize = 20 // Default page size
	}

	products, totalCount, err := h.productUseCase.GetActiveProducts(r.Context(), page, pageSize)
	if err != nil {
		log.Printf("Error retrieving active products for user %d: %v", userID, err)
		http.Error(w, "Failed to retrieve products", http.StatusInternalServerError)
		return
	}

	response := struct {
		Products   []*domain.Product `json:"products"`
		TotalCount int               `json:"totalCount"`
		Page       int               `json:"page"`
		PageSize   int               `json:"pageSize"`
	}{
		Products:   products,
		TotalCount: totalCount,
		Page:       page,
		PageSize:   pageSize,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding products to JSON for user %d: %v", userID, err)
		http.Error(w, "Failed to encode products", http.StatusInternalServerError)
		return
	}

	log.Printf("Successfully retrieved %d active products for user %d (page %d, pageSize %d, total %d)",
		len(products), userID, page, pageSize, totalCount)
}

// func (h *ProductHandler) CreateProduct(w http.ResponseWriter, r *http.Request) {

// 	role, ok := r.Context().Value("user_role").(string)
// 	if !ok || role != "admin" {
// 		http.Error(w, "Admin access required", http.StatusForbidden)
// 		return
// 	}

// 	var input struct {
// 		Product domain.Product        `json:"product"`
// 		Images  []domain.ProductImage `json:"images"`
// 	}

// 	err := json.NewDecoder(r.Body).Decode(&input)
// 	if err != nil {
// 		http.Error(w, "Invalid request body", http.StatusBadRequest)
// 		return
// 	}

// 	// Validate required fields
// 	if input.Product.Name == "" || input.Product.Price <= 0 || input.Product.StockQuantity < 0 {
// 		http.Error(w, "Missing or invalid required fields", http.StatusBadRequest)
// 		return
// 	}

// 	// product validations
//     if len(input.Product.Name) > 255 { // Adjust the length as per your database schema
//         http.Error(w, "Product name is too long", http.StatusBadRequest)
//         return
//     }

// 	// Validate images
// 	if len(input.Images) == 0 {
// 		http.Error(w, "At least one image is required", http.StatusBadRequest)
// 		return
// 	}

// 	primaryCount := 0
// 	for _, img := range input.Images {
// 		if img.IsPrimary {
// 			primaryCount++
// 		}
// 		if !strings.HasPrefix(img.ImageURL, "http://") && !strings.HasPrefix(img.ImageURL, "https://") {
// 			http.Error(w, "Invalid image URL", http.StatusBadRequest)
// 			return
// 		}
// 	}

// 	if primaryCount > 1 {
// 		http.Error(w, "Only one image can be set as primary", http.StatusBadRequest)
// 		return
// 	}

// 	// If no primary image is set, make the first one primary
// 	if primaryCount == 0 && len(input.Images) > 0 {
// 		input.Images[0].IsPrimary = true
// 	}

// 	err = h.productUseCase.CreateProductWithImages(r.Context(), &input.Product, input.Images)
// 	if err != nil {
// 		switch err {
// 		case usecase.ErrInvalidCategory, usecase.ErrInvalidSubCategory:
// 			http.Error(w, err.Error(), http.StatusBadRequest)
// 		default:
// 			log.Printf("Error creating product: %v", err)
// 			http.Error(w, "Failed to create product", http.StatusInternalServerError)
// 		}
// 		return
// 	}

//		w.Header().Set("Content-Type", "application/json")
//		w.WriteHeader(http.StatusCreated)
//		json.NewEncoder(w).Encode(input.Product)
//	}
func (h *ProductHandler) UpdatePrimaryImage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	productID, err := strconv.ParseInt(vars["productId"], 10, 64)
	if err != nil {
		http.Error(w, "Invalid product ID", http.StatusBadRequest)
		return
	}

	var input struct {
		ImageID int64 `json:"image_id"`
	}
	err = json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	err = h.productUseCase.UpdatePrimaryImage(r.Context(), productID, input.ImageID)
	if err != nil {
		if err.Error() == "product not found or already deleted" {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else if err.Error() == "image does not belong to the product" {
			http.Error(w, err.Error(), http.StatusBadRequest)
		} else {
			http.Error(w, "Failed to update primary image", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Primary image updated successfully"})
}

func (h *ProductHandler) AddProductImages(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	productID, err := strconv.ParseInt(vars["productId"], 10, 64)
	if err != nil {
		http.Error(w, "Invalid product ID", http.StatusBadRequest)
		return
	}

	var input struct {
		Images []domain.ProductImage `json:"images"`
	}
	err = json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if len(input.Images) == 0 {
		http.Error(w, "At least one image is required", http.StatusBadRequest)
		return
	}

	if len(input.Images) > usecase.MaxImagesPerProduct {
		http.Error(w, "Exceeded maximum number of images allowed", http.StatusBadRequest)
		return
	}

	for _, img := range input.Images {
		if !isValidImageURL(img.ImageURL) {
			http.Error(w, "Invalid image URL", http.StatusBadRequest)
			return
		}
	}

	err = h.productUseCase.AddProductImages(r.Context(), productID, input.Images)
	if err != nil {
		switch err {
		case usecase.ErrProductNotFound:
			http.Error(w, "Product not found", http.StatusNotFound)
		case usecase.ErrDuplicateImageURL:
			http.Error(w, "Duplicate image URL detected", http.StatusBadRequest)
		default:
			log.Printf("Error adding product images: %v", err)
			http.Error(w, "Failed to add product images", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "Images added successfully"})
}

func isValidImageURL(url string) bool {
	return regexp.MustCompile(`^https?://`).MatchString(url)
}

func (h *ProductHandler) RemoveProductImage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	productID, err := strconv.ParseInt(vars["productId"], 10, 64)
	if err != nil {
		http.Error(w, "Invalid product ID", http.StatusBadRequest)
		return
	}

	imageID, err := strconv.ParseInt(vars["imageId"], 10, 64)
	if err != nil {
		http.Error(w, "Invalid image ID", http.StatusBadRequest)
		return
	}

	err = h.productUseCase.RemoveProductImage(r.Context(), productID, imageID)
	if err != nil {
		switch err.Error() {
		case "product not found":
			http.Error(w, err.Error(), http.StatusNotFound)
		case "image not found":
			http.Error(w, err.Error(), http.StatusNotFound)
		case "cannot delete primary image":
			http.Error(w, err.Error(), http.StatusBadRequest)
		default:
			http.Error(w, "Failed to remove product image", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Image removed successfully"})
}

func (h *ProductHandler) CreateProduct(w http.ResponseWriter, r *http.Request) {
	// Check if the user is an admin
	role, ok := r.Context().Value("user_role").(string)
	if !ok || role != "admin" {
		http.Error(w, "Admin access required", http.StatusForbidden)
		return
	}

	// Parse the request body
	var input struct {
		Product domain.Product        `json:"product"`
		Images  []domain.ProductImage `json:"images"`
	}
	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate the product and images
	err = validator.ValidateProduct(&input.Product, input.Images)
	if err != nil {
		switch err {
		case utils.ErrInvalidProductName, utils.ErrProductNameTooLong:
			http.Error(w, "Invalid product name", http.StatusBadRequest)
		case utils.ErrInvalidProductDescription:
			http.Error(w, "Invalid product description", http.StatusBadRequest)
		case utils.ErrInvalidProductPrice:
			http.Error(w, "Invalid product price", http.StatusBadRequest)
		case utils.ErrProductDescriptionRequired:
			http.Error(w, "Product description is required", http.StatusBadRequest)
		case utils.ErrStockQuantRequired:
			http.Error(w, "Stock quantity is required", http.StatusBadRequest)
		case utils.ErrInvalidStockQuantity:
			http.Error(w, "Invalid stock quantity", http.StatusBadRequest)
		case utils.ErrInvalidCategoryID:
			http.Error(w, "Invalid category ID", http.StatusBadRequest)
		case utils.ErrInvalidSubCategoryID:
			http.Error(w, "Invalid sub-category ID", http.StatusBadRequest)
		case utils.ErrNoImages:
			http.Error(w, "At least one image is required", http.StatusBadRequest)
		case utils.ErrTooManyImages:
			http.Error(w, "Too many images", http.StatusBadRequest)
		case utils.ErrInvalidImageURL:
			http.Error(w, "Invalid image URL", http.StatusBadRequest)
		case utils.ErrMultiplePrimaryImages:
			http.Error(w, "Only one image can be set as primary", http.StatusBadRequest)
		case utils.ErrNoPrimaryImage:
			http.Error(w, "No primary image specified", http.StatusBadRequest)
		default:
			http.Error(w, "Invalid product data", http.StatusBadRequest)
		}
		return
	}

	// Create the product with images
	err = h.productUseCase.CreateProductWithImages(r.Context(), &input.Product, input.Images)
	if err != nil {
		switch err {
		case usecase.ErrInvalidCategory:
			http.Error(w, "Invalid category", http.StatusBadRequest)
		case usecase.ErrInvalidSubCategory:
			http.Error(w, "Invalid sub-category", http.StatusBadRequest)
		default:
			log.Printf("Error creating product: %v", err)
			http.Error(w, "Failed to create product", http.StatusInternalServerError)
		}
		return
	}

	// Prepare the response
	response := struct {
		Message string         `json:"message"`
		Product domain.Product `json:"product"`
	}{
		Message: "Product created successfully",
		Product: input.Product,
	}

	// Send the response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding response: %v", err)
	}
}
