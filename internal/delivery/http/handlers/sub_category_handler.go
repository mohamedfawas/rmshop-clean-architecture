package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/mohamedfawas/rmshop-clean-architecture/internal/domain"
	"github.com/mohamedfawas/rmshop-clean-architecture/internal/usecase"
	"github.com/mohamedfawas/rmshop-clean-architecture/pkg/utils"
)

type SubCategoryHandler struct {
	subCategoryUseCase usecase.SubCategoryUseCase
}

func NewSubCategoryHandler(subCategoryUseCase usecase.SubCategoryUseCase) *SubCategoryHandler {
	return &SubCategoryHandler{subCategoryUseCase: subCategoryUseCase}
}

func (h *SubCategoryHandler) CreateSubCategory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	categoryID, err := strconv.Atoi(vars["categoryId"])
	if err != nil {
		http.Error(w, "Invalid category ID", http.StatusBadRequest)
		return
	}

	var subCategory domain.SubCategory
	err = json.NewDecoder(r.Body).Decode(&subCategory)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	err = h.subCategoryUseCase.CreateSubCategory(r.Context(), categoryID, &subCategory)
	if err != nil {
		switch err {
		case utils.ErrInvalidSubCategoryName:
			http.Error(w, "Invalid subcategory name", http.StatusBadRequest)
		case utils.ErrSubCategoryNameTooLong:
			http.Error(w, "Subcategory name too long", http.StatusBadRequest)
		case utils.ErrDuplicateSubCategory:
			http.Error(w, "Subcategory already exists", http.StatusConflict)
		case utils.ErrCategoryNotFound:
			http.Error(w, "Parent category not found", http.StatusNotFound)
		default:
			http.Error(w, "Failed to create subcategory", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(subCategory)
}
