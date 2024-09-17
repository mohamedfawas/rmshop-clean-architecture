package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/mohamedfawas/rmshop-clean-architecture/internal/usecase"
	"github.com/mohamedfawas/rmshop-clean-architecture/pkg/api"
)

type AnalyticsHandler struct {
	analyticsUseCase usecase.AnalyticsUseCase
}

func NewAnalyticsHandler(analyticsUseCase usecase.AnalyticsUseCase) *AnalyticsHandler {
	return &AnalyticsHandler{analyticsUseCase: analyticsUseCase}
}

func (h *AnalyticsHandler) GetTopProducts(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	startDate := r.URL.Query().Get("start_date")
	endDate := r.URL.Query().Get("end_date")
	limitStr := r.URL.Query().Get("limit")
	sortBy := r.URL.Query().Get("sort")

	// Validate and set default values
	limit := 10
	if limitStr != "" {
		parsedLimit, err := strconv.Atoi(limitStr)
		if err != nil || parsedLimit <= 0 {
			api.SendResponse(w, http.StatusBadRequest, "Invalid limit parameter", nil, "Limit must be a positive integer")
			return
		}
		limit = parsedLimit
	}

	if sortBy == "" {
		sortBy = "quantity" // Default sort
	} else if sortBy != "quantity" && sortBy != "revenue" {
		api.SendResponse(w, http.StatusBadRequest, "Invalid sort parameter", nil, "Sort must be either 'quantity' or 'revenue'")
		return
	}

	var start, end time.Time
	var err error

	if startDate != "" {
		start, err = time.Parse("2006-01-02", startDate)
		if err != nil {
			api.SendResponse(w, http.StatusBadRequest, "Invalid start_date", nil, "Use format YYYY-MM-DD")
			return
		}
	}

	if endDate != "" {
		end, err = time.Parse("2006-01-02", endDate)
		if err != nil {
			api.SendResponse(w, http.StatusBadRequest, "Invalid end_date", nil, "Use format YYYY-MM-DD")
			return
		}
	}

	// Call use case
	topProducts, err := h.analyticsUseCase.GetTopProducts(r.Context(), start, end, limit, sortBy)
	if err != nil {
		api.SendResponse(w, http.StatusInternalServerError, "Failed to retrieve top products", nil, "An unexpected error occurred")
		return
	}

	api.SendResponse(w, http.StatusOK, "Top products retrieved successfully", topProducts, "")
}
