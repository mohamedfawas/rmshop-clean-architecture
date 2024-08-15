package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/mohamedfawas/rmshop-clean-architecture/internal/domain"
	"github.com/mohamedfawas/rmshop-clean-architecture/internal/usecase"
)

type ProductHandler struct {
	productUseCase usecase.ProductUseCase
}

func NewProductHandler(productUseCase usecase.ProductUseCase) *ProductHandler {
	return &ProductHandler{productUseCase: productUseCase}
}

// func (h *ProductHandler) CreateProduct(w http.ResponseWriter, r *http.Request) {
// 	var product domain.Product
// 	err := json.NewDecoder(r.Body).Decode(&product)
// 	if err != nil {
// 		http.Error(w, "Invalid request body", http.StatusBadRequest)
// 		return
// 	}

// 	err = h.productUseCase.CreateProduct(r.Context(), &product)
// 	if err != nil {
// 		http.Error(w, "Failed to create product", http.StatusInternalServerError)
// 		return
// 	}

// 	w.Header().Set("Content-Type", "application/json")
// 	w.WriteHeader(http.StatusCreated)
// 	json.NewEncoder(w).Encode(product)
// }

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

func (h *ProductHandler) CreateProduct(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Product domain.Product        `json:"product"`
		Images  []domain.ProductImage `json:"images"`
	}
	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate that we have at least one image
	if len(input.Images) == 0 {
		http.Error(w, "At least one image is required", http.StatusBadRequest)
		return
	}

	// Ensure only one image is set as primary
	primaryCount := 0
	for _, img := range input.Images {
		if img.IsPrimary {
			primaryCount++
		}
	}
	if primaryCount > 1 {
		http.Error(w, "Only one image can be set as primary", http.StatusBadRequest)
		return
	}

	// If no primary image is set, make the first one primary
	if primaryCount == 0 {
		input.Images[0].IsPrimary = true
	}

	err = h.productUseCase.CreateProductWithImages(r.Context(), &input.Product, input.Images)
	if err != nil {
		http.Error(w, "Failed to create product", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(input.Product)
}
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

	err = h.productUseCase.AddProductImages(r.Context(), productID, input.Images)
	if err != nil {
		if err.Error() == "product not found" {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, "Failed to add product images", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "Images added successfully"})
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
