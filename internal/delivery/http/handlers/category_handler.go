package handlers

import (
	"encoding/json"
	"log"
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
		api.SendResponse(w, http.StatusBadRequest, "Failed to create category", nil, "Invalid request body")
		return
	}

	category.Name = strings.ToLower(strings.TrimSpace(category.Name))

	err = validator.ValidateCategoryName(category.Name)
	if err != nil {
		switch err {
		case utils.ErrInvalidCategoryName:
			api.SendResponse(w, http.StatusBadRequest, "Failed to create category", nil, "Please provide a valid category name")
		case utils.ErrCategoryNameTooShort:
			api.SendResponse(w, http.StatusBadRequest, "Failed to create category", nil, "Category name should have at least 2 characters")
		case utils.ErrCategoryNameTooLong:
			api.SendResponse(w, http.StatusBadRequest, "Failed to create category", nil, "Category name length should not be greater than 50 characters")
		case utils.ErrCategoryNameNumeric:
			api.SendResponse(w, http.StatusBadRequest, "Failed to create category", nil, "Category name should not be numeric")
		default:
			log.Printf("Unexpected error in category name validation: %v", err)
			api.SendResponse(w, http.StatusBadRequest, "Failed to create category", nil, "Invalid category name")
		}
		return
	}

	err = h.categoryUseCase.CreateCategory(r.Context(), &category)
	if err != nil {
		switch err {
		case utils.ErrDuplicateCategory:
			api.SendResponse(w, http.StatusConflict, "Failed to create category", nil, "Category already exists")
		case utils.ErrDBCreateCategory:
			api.SendResponse(w, http.StatusInternalServerError, "Failed to create category", nil, "Failed to create the category entry in database")
		default:
			api.SendResponse(w, http.StatusInternalServerError, "Failed to create category", nil, "An unexpected error occurred")
		}
		return
	}

	api.SendResponse(w, http.StatusCreated, "Category created successfully", category, "")
}

func (h *CategoryHandler) GetAllCategories(w http.ResponseWriter, r *http.Request) {
	categories, err := h.categoryUseCase.GetAllCategories(r.Context())
	if err != nil {
		switch err {
		case utils.ErrNoCategoriesFound:
			http.Error(w, "No categories found", http.StatusNotFound)
		case utils.ErrQueryExecution, utils.ErrRowScan:
			http.Error(w, utils.InternalServerErrorString, http.StatusInternalServerError)
		}

		http.Error(w, "Failed to retrieve categories", http.StatusInternalServerError)
		return
	}

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

	category.Name = strings.TrimSpace(category.Name) //remove trailinng and leading white spaces

	err = validator.ValidateCategoryName(category.Name)
	if err != nil {
		switch err {
		case utils.ErrInvalidCategoryName:
			http.Error(w, "Invalid category name", http.StatusBadRequest)
		case utils.ErrCategoryNameTooShort:
			http.Error(w, "Category name should have atleast 2 characters", http.StatusBadRequest)
		case utils.ErrCategoryNameTooLong:
			http.Error(w, "Category name length should not be greater than 50 characters", http.StatusBadRequest)
		case utils.ErrCategoryNameNumeric:
			http.Error(w, "Category name should not be purely numeric", http.StatusBadRequest)
		default:
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
		}
		return
	}

	err = h.categoryUseCase.UpdateCategory(r.Context(), &category)
	if err != nil {
		//log.Printf("Error updating category: %v", err)
		switch err {
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

func (h *CategoryHandler) SoftDeleteCategory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	categoryID, err := strconv.Atoi(vars["categoryId"])
	if err != nil {
		log.Printf("Invalid category ID: %v", err)
		http.Error(w, "Invalid category ID", http.StatusBadRequest)
		return
	}

	err = h.categoryUseCase.SoftDeleteCategory(r.Context(), categoryID)
	if err != nil {
		log.Printf("Error soft deleting category: %v", err)
		if err == utils.ErrCategoryNotFound {
			http.Error(w, "Category not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to delete category", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
