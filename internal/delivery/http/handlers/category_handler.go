package handlers

import (
	"encoding/json"
	"net/http"

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
