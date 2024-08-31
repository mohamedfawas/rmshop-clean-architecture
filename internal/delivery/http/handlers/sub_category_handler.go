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
			api.SendResponse(w, http.StatusInternalServerError, "Failed to create subcategory", nil, "An unexpected error occured")
		}
		return
	}

	api.SendResponse(w, http.StatusCreated, "Subcategory created successfully", subCategory, "")
}

func (h *SubCategoryHandler) GetSubCategoriesByCategory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	categoryID, err := strconv.Atoi(vars["categoryId"])
	if err != nil {
		api.SendResponse(w, http.StatusBadRequest, "Failed to retrieve sub-categories", nil, "Invalid category ID")
		return
	}

	subCategories, err := h.subCategoryUseCase.GetSubCategoriesByCategory(r.Context(), categoryID)
	if err != nil {
		log.Printf("Error retrieving sub-categories: %v", err)
		if err == utils.ErrCategoryNotFound {
			api.SendResponse(w, http.StatusNotFound, "Failed to retrieve sub-categories", nil, "Category not found")
		} else {
			api.SendResponse(w, http.StatusInternalServerError, "Failed to retrieve sub-categories", nil, "An unexpected error occurred")
		}
		return
	}

	if len(subCategories) == 0 {
		api.SendResponse(w, http.StatusOK, "No sub-categories found", []struct{}{}, "")
		return
	}

	api.SendResponse(w, http.StatusOK, "Sub-categories retrieved successfully", subCategories, "")
}

func (h *SubCategoryHandler) GetSubCategoryByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	categoryID, err := strconv.Atoi(vars["categoryId"])
	if err != nil {
		api.SendResponse(w, http.StatusBadRequest, "Failed to retrieve sub-category", nil, "Invalid category ID")
		return
	}

	subCategoryID, err := strconv.Atoi(vars["subcategoryId"])
	if err != nil {
		api.SendResponse(w, http.StatusBadRequest, "Failed to retrieve sub-category", nil, "Invalid sub-category ID")
		return
	}

	subCategory, err := h.subCategoryUseCase.GetSubCategoryByID(r.Context(), categoryID, subCategoryID)
	if err != nil {
		switch err {
		case utils.ErrCategoryNotFound:
			api.SendResponse(w, http.StatusNotFound, "Failed to retrieve sub-category", nil, "Category not found")
		case utils.ErrSubCategoryNotFound:
			api.SendResponse(w, http.StatusNotFound, "Failed to retrieve sub-category", nil, "Sub-category not found")
		default:
			api.SendResponse(w, http.StatusInternalServerError, "Failed to retrieve sub-category", nil, "An unexpected error occurred")
		}
		return
	}

	api.SendResponse(w, http.StatusOK, "Sub-category retrieved successfully", subCategory, "")
}

func (h *SubCategoryHandler) UpdateSubCategory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	categoryID, err := strconv.Atoi(vars["categoryId"])
	if err != nil {
		api.SendResponse(w, http.StatusBadRequest, "Failed to update sub-category", nil, "Invalid category ID")
		return
	}

	subCategoryID, err := strconv.Atoi(vars["subcategoryId"])
	if err != nil {
		api.SendResponse(w, http.StatusBadRequest, "Failed to update sub-category", nil, "Invalid sub-category ID")
		return
	}

	var subCategory domain.SubCategory
	err = json.NewDecoder(r.Body).Decode(&subCategory)
	if err != nil {
		api.SendResponse(w, http.StatusBadRequest, "Failed to update sub-category", nil, "Invalid request body")
		return
	}

	subCategory.ID = subCategoryID

	err = h.subCategoryUseCase.UpdateSubCategory(r.Context(), categoryID, &subCategory)
	if err != nil {
		switch err {
		case utils.ErrCategoryNotFound:
			api.SendResponse(w, http.StatusNotFound, "Failed to update sub-category", nil, "Category not found")
		case utils.ErrSubCategoryNotFound:
			api.SendResponse(w, http.StatusNotFound, "Failed to update sub-category", nil, "Sub-category not found")
		case utils.ErrInvalidSubCategoryName:
			api.SendResponse(w, http.StatusBadRequest, "Failed to update sub-category", nil, "Invalid sub-category name")
		case utils.ErrSubCategoryNameTooLong:
			api.SendResponse(w, http.StatusBadRequest, "Failed to update sub-category", nil, "Sub-category name too long")
		default:
			api.SendResponse(w, http.StatusInternalServerError, "Failed to update sub-category", nil, "An unexpected error occurred")
		}
		return
	}

	api.SendResponse(w, http.StatusOK, "Sub-category updated successfully", subCategory, "")
}

func (h *SubCategoryHandler) SoftDeleteSubCategory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	categoryID, err := strconv.Atoi(vars["categoryId"])
	if err != nil {
		api.SendResponse(w, http.StatusBadRequest, "Failed to delete sub-category", nil, "Invalid category ID")
		return
	}

	subCategoryID, err := strconv.Atoi(vars["subcategoryId"])
	if err != nil {
		api.SendResponse(w, http.StatusBadRequest, "Failed to delete sub-category", nil, "Invalid sub-category ID")
		return
	}

	err = h.subCategoryUseCase.SoftDeleteSubCategory(r.Context(), categoryID, subCategoryID)
	if err != nil {
		switch err {
		case utils.ErrCategoryNotFound:
			api.SendResponse(w, http.StatusNotFound, "Failed to delete sub-category", nil, "Category not found")
		case utils.ErrSubCategoryNotFound:
			api.SendResponse(w, http.StatusNotFound, "Failed to delete sub-category", nil, "Sub-category not found")
		case utils.ErrSubCategoryAlreadyDeleted:
			api.SendResponse(w, http.StatusNotFound, "Failed to delete sub-category", nil, "Sub-category already deleted")
		default:
			api.SendResponse(w, http.StatusInternalServerError, "Failed to delete sub-category", nil, "An unexpected error occurred")
		}
		return
	}

	api.SendResponse(w, http.StatusOK, "Sub-category deleted successfully", nil, "")
}
