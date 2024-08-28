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
		}
		return
	}

	err = h.categoryUseCase.CreateCategory(r.Context(), &category)
	if err != nil {
		switch err {
		case utils.ErrDuplicateCategory:
			api.SendResponse(w, http.StatusConflict, "Failed to create category", nil, "Category already exists")
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
			api.SendResponse(w, http.StatusNotFound, "Failed to get categories", nil, "categories not found")
		default:
			api.SendResponse(w, http.StatusInternalServerError, "Internal server error", nil, "An unexpected error occured")
		}
		return
	}

	if len(categories) == 0 {
		api.SendResponse(w, http.StatusOK, "No categories found", []struct{}{}, "")
		return
	}

	api.SendResponse(w, http.StatusOK, "Categories retrieved successfully", categories, "")
}

func (h *CategoryHandler) GetActiveCategoryByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	categoryID, err := strconv.Atoi(vars["categoryId"])
	if err != nil {
		api.SendResponse(w, http.StatusBadRequest, "Failed to get category details", nil, "Please provide a valid category id")
		return
	}

	category, err := h.categoryUseCase.GetActiveCategoryByID(r.Context(), categoryID)
	if err != nil {
		if err == utils.ErrCategoryNotFound {
			api.SendResponse(w, http.StatusNotFound, "Failed to get category details", nil, "Category not found")
		} else {
			api.SendResponse(w, http.StatusInternalServerError, "Failed to get category details", nil, "An unexpected error occured")
		}
		return
	}

	api.SendResponse(w, http.StatusOK, "category successfully retrieved", category, "")
}

func (h *CategoryHandler) UpdateCategory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	categoryID, err := strconv.Atoi(vars["categoryId"])
	if err != nil {
		api.SendResponse(w, http.StatusBadRequest, "Failed to update category", nil, "Invalid category ID")
		return
	}

	var category domain.Category
	err = json.NewDecoder(r.Body).Decode(&category)
	if err != nil {
		api.SendResponse(w, http.StatusBadRequest, "Failed to update category", nil, "Invalid request body")
		return
	}

	category.ID = categoryID

	category.Name = strings.TrimSpace(category.Name) //remove trailinng and leading white spaces

	err = validator.ValidateCategoryName(category.Name)
	if err != nil {
		switch err {
		case utils.ErrInvalidCategoryName:
			api.SendResponse(w, http.StatusBadRequest, "Failed to update category", nil, "Invalid category name")
		case utils.ErrCategoryNameTooShort:
			api.SendResponse(w, http.StatusBadRequest, "Failed to update category", nil, "Category name should have at least 2 characters")
		case utils.ErrCategoryNameTooLong:
			api.SendResponse(w, http.StatusBadRequest, "Failed to update category", nil, "Category name length should not be greater than 50 characters")
		case utils.ErrCategoryNameNumeric:
			api.SendResponse(w, http.StatusBadRequest, "Failed to update category", nil, "Category name should not be purely numeric")
		default:
			api.SendResponse(w, http.StatusInternalServerError, "Failed to update category", nil, "An unexpected error occured")
		}
		return
	}

	err = h.categoryUseCase.UpdateCategory(r.Context(), &category)
	if err != nil {
		switch err {
		case utils.ErrDuplicateCategory:
			api.SendResponse(w, http.StatusConflict, "Failed to update category", nil, "Category name already exists")
		case utils.ErrCategoryNotFound:
			api.SendResponse(w, http.StatusNotFound, "Failed to update category", nil, "Category not found")
		default:
			api.SendResponse(w, http.StatusInternalServerError, "Failed to update category", nil, "An unexpected error occurred")
		}
		return
	}

	api.SendResponse(w, http.StatusOK, "Category updated successfully", category, "")
}

func (h *CategoryHandler) SoftDeleteCategory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	categoryID, err := strconv.Atoi(vars["categoryId"])
	if err != nil {
		log.Printf("Invalid category ID: %v", err)
		api.SendResponse(w, http.StatusBadRequest, "Failed to delete category", nil, "Invalid category ID")
		return
	}

	err = h.categoryUseCase.SoftDeleteCategory(r.Context(), categoryID)
	if err != nil {
		log.Printf("Error soft deleting category: %v", err)
		if err == utils.ErrCategoryNotFound {
			api.SendResponse(w, http.StatusNotFound, "Failed to delete category", nil, "Category not found")
		} else {
			api.SendResponse(w, http.StatusInternalServerError, "Failed to delete category", nil, "An unexpected error occurred")
		}
		return
	}

	api.SendResponse(w, http.StatusOK, "Category deleted successfully", nil, "")
}
