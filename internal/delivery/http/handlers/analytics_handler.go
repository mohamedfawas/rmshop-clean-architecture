package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/mohamedfawas/rmshop-clean-architecture/internal/domain"
	"github.com/mohamedfawas/rmshop-clean-architecture/internal/usecase"
	"github.com/mohamedfawas/rmshop-clean-architecture/pkg/api"
	"github.com/mohamedfawas/rmshop-clean-architecture/pkg/utils"
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

func (h *AnalyticsHandler) GetTopCategories(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	startDate, endDate, err := parseTimeRange(r)
	if err != nil {
		api.SendResponse(w, http.StatusBadRequest, "Invalid date format", nil, err.Error())
		return
	}

	limit, err := parseLimit(r)
	if err != nil {
		api.SendResponse(w, http.StatusBadRequest, "Invalid limit", nil, err.Error())
		return
	}

	sortOrder := r.URL.Query().Get("sort")
	if sortOrder != "asc" && sortOrder != "desc" {
		sortOrder = "desc"
	}

	// Call use case
	categories, err := h.analyticsUseCase.GetTopCategories(r.Context(), startDate, endDate, limit, sortOrder)
	if err != nil {
		api.SendResponse(w, http.StatusInternalServerError, "Failed to retrieve top categories", nil, "An unexpected error occurred")
		return
	}

	api.SendResponse(w, http.StatusOK, "Top categories retrieved successfully", categories, "")
}

func parseTimeRange(r *http.Request) (time.Time, time.Time, error) {
	startDateStr := r.URL.Query().Get("start_date")
	endDateStr := r.URL.Query().Get("end_date")

	var startDate, endDate time.Time
	var err error

	if startDateStr != "" {
		startDate, err = time.Parse("2006-01-02", startDateStr)
		if err != nil {
			return time.Time{}, time.Time{}, err
		}
	} else {
		startDate = time.Now().AddDate(0, -1, 0) // Default to last month
	}

	if endDateStr != "" {
		endDate, err = time.Parse("2006-01-02", endDateStr)
		if err != nil {
			return time.Time{}, time.Time{}, err
		}
	} else {
		endDate = time.Now()
	}

	return startDate, endDate, nil
}

func parseLimit(r *http.Request) (int, error) {
	limitStr := r.URL.Query().Get("limit")
	if limitStr == "" {
		return 10, nil // Default limit
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		return 0, err
	}

	if limit <= 0 || limit > 10 {
		return 10, nil // Enforce maximum limit
	}

	return limit, nil
}

func (h *AnalyticsHandler) GetTopSubcategories(w http.ResponseWriter, r *http.Request) {
	params := domain.SubcategoryAnalyticsParams{
		Page:  1,
		Limit: 10,
	}

	// Parse query parameters
	if page, err := strconv.Atoi(r.URL.Query().Get("page")); err == nil && page > 0 {
		params.Page = page
	}
	if limit, err := strconv.Atoi(r.URL.Query().Get("limit")); err == nil && limit > 0 {
		params.Limit = limit
	}
	if categoryID, err := strconv.Atoi(r.URL.Query().Get("category_id")); err == nil {
		params.CategoryID = categoryID
	}
	if sortBy := r.URL.Query().Get("sort_by"); sortBy == "revenue" || sortBy == "quantity" {
		params.SortBy = sortBy
	}

	// Parse date range
	if startDate, err := time.Parse("2006-01-02", r.URL.Query().Get("start_date")); err == nil {
		params.StartDate = startDate
	}
	if endDate, err := time.Parse("2006-01-02", r.URL.Query().Get("end_date")); err == nil {
		params.EndDate = endDate
	}

	// Call use case
	subcategories, err := h.analyticsUseCase.GetTopSubcategories(r.Context(), params)
	if err != nil {
		if err == utils.ErrInvalidDateRange {
			api.SendResponse(w, http.StatusBadRequest, "Invalid date range", nil, err.Error())
		} else {
			api.SendResponse(w, http.StatusInternalServerError, "Failed to retrieve top subcategories", nil, "An unexpected error occurred")
		}
		return
	}

	api.SendResponse(w, http.StatusOK, "Top subcategories retrieved successfully", subcategories, "")
}
