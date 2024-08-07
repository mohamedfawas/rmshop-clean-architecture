package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/mohamedfawas/rmshop-clean-architecture/internal/domain"
	"github.com/mohamedfawas/rmshop-clean-architecture/internal/usecase"
	"github.com/mohamedfawas/rmshop-clean-architecture/pkg/utils"
)

type CategoryHandler struct {
	categoryUseCase usecase.CategoryUseCase
}

func NewCategoryHandler(categoryUseCase usecase.CategoryUseCase) *CategoryHandler {
	return &CategoryHandler{categoryUseCase: categoryUseCase}
}

func (h *CategoryHandler) CreateCategory(w http.ResponseWriter, r *http.Request) {
	var category domain.Category
	err := json.NewDecoder(r.Body).Decode(&category)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	err = h.categoryUseCase.CreateCategory(r.Context(), &category)
	if err != nil {
		switch err {
		case utils.ErrInvalidCategoryName:
			http.Error(w, "Invalid category name", http.StatusBadRequest)
		case utils.ErrCategoryNameTooLong:
			http.Error(w, "Category name too long", http.StatusBadRequest)
		case utils.ErrDuplicateCategory:
			http.Error(w, "Category already exists", http.StatusConflict)
		default:
			http.Error(w, "Failed to create category", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(category)
}

func (h *CategoryHandler) GetAllCategories(w http.ResponseWriter, r *http.Request) {
	log.Println("Entering GetAllCategories handler")

	// Log request details
	log.Printf("Request Method: %s", r.Method)
	log.Printf("Request URL: %s", r.URL)
	log.Printf("Request Headers: %+v", r.Header)

	categories, err := h.categoryUseCase.GetAllCategories(r.Context())
	if err != nil {
		log.Printf("Error retrieving categories: %v", err)
		http.Error(w, "Failed to retrieve categories", http.StatusInternalServerError)
		return
	}

	log.Printf("Retrieved %d categories", len(categories))

	w.Header().Set("Content-Type", "application/json")

	encodedData, err := json.Marshal(categories)
	if err != nil {
		log.Printf("Error marshaling categories: %v", err)
		http.Error(w, "Failed to encode categories", http.StatusInternalServerError)
		return
	}

	log.Printf("Marshaled data: %s", string(encodedData))

	_, err = w.Write(encodedData)
	if err != nil {
		log.Printf("Error writing response: %v", err)
		http.Error(w, "Failed to send response", http.StatusInternalServerError)
		return
	}

	log.Println("Successfully sent categories response")
}

func (h *CategoryHandler) GetActiveCategoryByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	categoryID, err := strconv.Atoi(vars["categoryId"])
	if err != nil {
		http.Error(w, "Invalid category ID", http.StatusBadRequest)
		return
	}

	category, err := h.categoryUseCase.GetActiveCategoryByID(r.Context(), categoryID)
	if err != nil {
		if err == utils.ErrCategoryNotFound {
			http.Error(w, "Category not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to retrieve category", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(category)
}

func (h *CategoryHandler) UpdateCategory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	categoryID, err := strconv.Atoi(vars["categoryId"])
	if err != nil {
		log.Printf("Invalid category ID: %v", err)
		http.Error(w, "Invalid category ID", http.StatusBadRequest)
		return
	}

	var category domain.Category
	err = json.NewDecoder(r.Body).Decode(&category)
	if err != nil {
		log.Printf("Invalid request body: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	category.ID = categoryID

	err = h.categoryUseCase.UpdateCategory(r.Context(), &category)
	if err != nil {
		log.Printf("Error updating category: %v", err)
		switch err {
		case utils.ErrInvalidCategoryName:
			http.Error(w, "Invalid category name", http.StatusBadRequest)
		case utils.ErrCategoryNameTooLong:
			http.Error(w, "Category name too long", http.StatusBadRequest)
		case utils.ErrDuplicateCategory:
			http.Error(w, "Category name already exists", http.StatusConflict)
		case utils.ErrCategoryNotFound:
			http.Error(w, "Category not found", http.StatusNotFound)
		default:
			http.Error(w, "Failed to update category", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(category)
}
