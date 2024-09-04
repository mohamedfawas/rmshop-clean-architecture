package handlers

import (
	"log"
	"net/http"
	"strconv"

	"github.com/mohamedfawas/rmshop-clean-architecture/internal/domain"
	"github.com/mohamedfawas/rmshop-clean-architecture/internal/usecase"
	"github.com/mohamedfawas/rmshop-clean-architecture/pkg/api"
)

type InventoryHandler struct {
	inventoryUseCase usecase.InventoryUseCase
}

func NewInventoryHandler(inventoryUseCase usecase.InventoryUseCase) *InventoryHandler {
	return &InventoryHandler{inventoryUseCase: inventoryUseCase}
}

func (h *InventoryHandler) GetInventory(w http.ResponseWriter, r *http.Request) {
	params := domain.InventoryQueryParams{
		Page:  1,
		Limit: 10,
	}

	if productID, err := strconv.ParseInt(r.URL.Query().Get("product_id"), 10, 64); err == nil {
		params.ProductID = productID
	}

	params.ProductName = r.URL.Query().Get("product_name")

	if categoryID, err := strconv.ParseInt(r.URL.Query().Get("category_id"), 10, 64); err == nil {
		params.CategoryID = categoryID
	}

	params.CategoryName = r.URL.Query().Get("category_name")

	if stockLessThan, err := strconv.Atoi(r.URL.Query().Get("stock_less_than")); err == nil {
		params.StockLessThan = &stockLessThan
	}

	if stockMoreThan, err := strconv.Atoi(r.URL.Query().Get("stock_more_than")); err == nil {
		params.StockMoreThan = &stockMoreThan
	}

	if page, err := strconv.Atoi(r.URL.Query().Get("page")); err == nil && page > 0 {
		params.Page = page
	}

	if limit, err := strconv.Atoi(r.URL.Query().Get("limit")); err == nil && limit > 0 {
		params.Limit = limit
	}

	params.SortBy = r.URL.Query().Get("sort_by")
	params.SortOrder = r.URL.Query().Get("order")

	inventory, total, err := h.inventoryUseCase.GetInventory(r.Context(), params)
	if err != nil {
		log.Printf("error : %v", err)
		api.SendResponse(w, http.StatusInternalServerError, "Failed to retrieve inventory", nil, "An unexpected error occurred")
		return
	}

	response := map[string]interface{}{
		"inventory":   inventory,
		"total_count": total,
		"page":        params.Page,
		"limit":       params.Limit,
		"total_pages": (total + int64(params.Limit) - 1) / int64(params.Limit),
	}

	api.SendResponse(w, http.StatusOK, "Inventory retrieved successfully", response, "")
}
