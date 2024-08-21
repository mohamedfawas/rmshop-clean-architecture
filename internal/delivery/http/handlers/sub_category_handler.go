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
		api.SendResponse(w, http.StatusBadRequest, "Failed to create subcategory", nil, "Invalid category ID")
		return
	}

	var subCategory domain.SubCategory
	err = json.NewDecoder(r.Body).Decode(&subCategory)
	if err != nil {
		api.SendResponse(w, http.StatusBadRequest, "Failed to create subcategory", nil, "Invalid request body")
		return
	}

	// Trim whitespace from subcategory name
	subCategory.Name = strings.ToLower(strings.TrimSpace(subCategory.Name))

	err = validator.ValidateSubCategoryName(subCategory.Name)
	if err != nil {
		switch err {
		case utils.ErrInvalidSubCategoryName:
			api.SendResponse(w, http.StatusBadRequest, "Failed to create subcategory", nil, "Please provide a valid sub category name")
		case utils.ErrSubCategoryNameTooShort:
			api.SendResponse(w, http.StatusBadRequest, "Failed to create subcategory", nil, "Sub category name too short: should have at least two characters")
		case utils.ErrSubCategoryNameTooLong:
			api.SendResponse(w, http.StatusBadRequest, "Failed to create subcategory", nil, "Sub category name too long: should be less than 50 characters")
		case utils.ErrSubCategoryNameNumeric:
			api.SendResponse(w, http.StatusBadRequest, "Failed to create subcategory", nil, "Sub category name should not be numeric")
		default:
			api.SendResponse(w, http.StatusInternalServerError, "Failed to create subcategory", nil, "Internal server error")
		}
		return
	}

	err = h.subCategoryUseCase.CreateSubCategory(r.Context(), categoryID, &subCategory)
	if err != nil {
		switch err {
		case utils.ErrDuplicateSubCategory:
			api.SendResponse(w, http.StatusConflict, "Failed to create subcategory", nil, "Subcategory already exists")
		case utils.ErrCategoryNotFound:
			api.SendResponse(w, http.StatusNotFound, "Failed to create subcategory", nil, "Parent category not found")
		default:
			api.SendResponse(w, http.StatusInternalServerError, "Failed to create subcategory", nil, "Failed to create subcategory")
		}
		return
	}

	api.SendResponse(w, http.StatusCreated, "Subcategory created successfully", subCategory, "")
}

func (h *SubCategoryHandler) GetSubCategoriesByCategory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	categoryID, err := strconv.Atoi(vars["categoryId"])
	if err != nil {
		http.Error(w, "Invalid category ID", http.StatusBadRequest)
		return
	}

	subCategories, err := h.subCategoryUseCase.GetSubCategoriesByCategory(r.Context(), categoryID)
	if err != nil {
		log.Printf("Error retrieving sub-categories: %v", err)
		if err == utils.ErrCategoryNotFound {
			http.Error(w, "Category not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to retrieve sub-categories", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(subCategories)
}

func (h *SubCategoryHandler) GetSubCategoryByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	categoryID, err := strconv.Atoi(vars["categoryId"])
	if err != nil {
		http.Error(w, "Invalid category ID", http.StatusBadRequest)
		return
	}

	subCategoryID, err := strconv.Atoi(vars["subcategoryId"])
	if err != nil {
		http.Error(w, "Invalid sub-category ID", http.StatusBadRequest)
		return
	}

	subCategory, err := h.subCategoryUseCase.GetSubCategoryByID(r.Context(), categoryID, subCategoryID)
	if err != nil {
		switch err {
		case utils.ErrCategoryNotFound:
			http.Error(w, "Category not found", http.StatusNotFound)
		case utils.ErrSubCategoryNotFound:
			http.Error(w, "Sub-category not found", http.StatusNotFound)
		default:
			http.Error(w, "Failed to retrieve sub-category", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(subCategory)
}

func (h *SubCategoryHandler) UpdateSubCategory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	categoryID, err := strconv.Atoi(vars["categoryId"])
	if err != nil {
		http.Error(w, "Invalid category ID", http.StatusBadRequest)
		return
	}

	subCategoryID, err := strconv.Atoi(vars["subcategoryId"])
	if err != nil {
		http.Error(w, "Invalid sub-category ID", http.StatusBadRequest)
		return
	}

	var subCategory domain.SubCategory
	err = json.NewDecoder(r.Body).Decode(&subCategory)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	subCategory.ID = subCategoryID

	err = h.subCategoryUseCase.UpdateSubCategory(r.Context(), categoryID, &subCategory)
	if err != nil {
		switch err {
		case utils.ErrCategoryNotFound:
			http.Error(w, "Category not found", http.StatusNotFound)
		case utils.ErrSubCategoryNotFound:
			http.Error(w, "Sub-category not found", http.StatusNotFound)
		case utils.ErrInvalidSubCategoryName:
			http.Error(w, "Invalid sub-category name", http.StatusBadRequest)
		case utils.ErrSubCategoryNameTooLong:
			http.Error(w, "Sub-category name too long", http.StatusBadRequest)
		default:
			http.Error(w, "Failed to update sub-category", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(subCategory)
}

func (h *SubCategoryHandler) SoftDeleteSubCategory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	categoryID, err := strconv.Atoi(vars["categoryId"])
	if err != nil {
		http.Error(w, "Invalid category ID", http.StatusBadRequest)
		return
	}

	subCategoryID, err := strconv.Atoi(vars["subcategoryId"])
	if err != nil {
		http.Error(w, "Invalid sub-category ID", http.StatusBadRequest)
		return
	}

	err = h.subCategoryUseCase.SoftDeleteSubCategory(r.Context(), categoryID, subCategoryID)
	if err != nil {
		switch err {
		case utils.ErrCategoryNotFound:
			http.Error(w, "Category not found", http.StatusNotFound)
		case utils.ErrSubCategoryNotFound:
			http.Error(w, "Sub-category not found", http.StatusNotFound)
		default:
			http.Error(w, "Failed to delete sub-category", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
