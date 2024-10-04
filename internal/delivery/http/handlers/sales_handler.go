package handlers

import (
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/mohamedfawas/rmshop-clean-architecture/internal/usecase"
	"github.com/mohamedfawas/rmshop-clean-architecture/pkg/api"
	"github.com/mohamedfawas/rmshop-clean-architecture/pkg/utils"
)

type SalesHandler struct {
	salesUseCase usecase.SalesUseCase
}

func NewSalesHandler(salesUseCase usecase.SalesUseCase) *SalesHandler {
	return &SalesHandler{salesUseCase: salesUseCase}
}

func (h *SalesHandler) GetDailySalesReport(w http.ResponseWriter, r *http.Request) {
	// Extract query parameters
	dateStr := r.URL.Query().Get("date")
	format := r.URL.Query().Get("format")

	// Validate date
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		api.SendResponse(w, http.StatusBadRequest, "Invalid date format", nil, "Use YYYY-MM-DD format")
		return
	}

	// Check if date is in the future
	if date.After(time.Now()) {
		api.SendResponse(w, http.StatusBadRequest, "Invalid date", nil, "Cannot generate report for future dates")
		return
	}

	// Validate format
	if format != "json" && format != "pdf" && format != "excel" {
		api.SendResponse(w, http.StatusBadRequest, "Invalid format", nil, "Supported formats: json, pdf, excel")
		return
	}

	// Call use case
	report, err := h.salesUseCase.GenerateDailySalesReport(r.Context(), date, format)
	if err != nil {
		if err == utils.ErrNoDataFound {
			api.SendResponse(w, http.StatusNoContent, "No sales data for the specified date", nil, "")
			return
		}
		log.Printf("error : %v", err)
		api.SendResponse(w, http.StatusInternalServerError, "Failed to generate report", nil, "An unexpected error occurred")
		return
	}

	// Set appropriate headers based on format
	switch format {
	case "pdf":
		w.Header().Set("Content-Type", "application/pdf")
		w.Header().Set("Content-Disposition", "attachment; filename=daily_sales_report.pdf")
	case "excel":
		w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
		w.Header().Set("Content-Disposition", "attachment; filename=daily_sales_report.xlsx")
	default:
		w.Header().Set("Content-Type", "application/json")
	}

	// Write the report to the response
	w.Write(report)
}

func (h *SalesHandler) GetWeeklySalesReport(w http.ResponseWriter, r *http.Request) {
	// Extract query parameters
	startDateStr := r.URL.Query().Get("start_date")
	format := r.URL.Query().Get("format")

	// Use current date if start_date is not provided
	var startDate time.Time
	var err error
	if startDateStr == "" {
		startDate = time.Now().AddDate(0, 0, -7) // Start from a week ago
	} else {
		startDate, err = time.Parse("2006-01-02", startDateStr)
		if err != nil {
			api.SendResponse(w, http.StatusBadRequest, "Invalid date format", nil, "Use YYYY-MM-DD format")
			return
		}
	}

	// Check if date is in the future
	if startDate.After(time.Now()) {
		api.SendResponse(w, http.StatusBadRequest, "Invalid date", nil, "Cannot generate report for future dates")
		return
	}

	// Validate format
	if format == "" {
		format = "json" // Default format
	}
	if format != "json" && format != "pdf" && format != "excel" {
		api.SendResponse(w, http.StatusBadRequest, "Invalid format", nil, "Supported formats: json, pdf, excel")
		return
	}

	// Call use case
	report, err := h.salesUseCase.GenerateWeeklySalesReport(r.Context(), startDate, format)
	if err != nil {
		if err == utils.ErrNoDataFound {
			api.SendResponse(w, http.StatusNoContent, "No sales data for the specified week", nil, "")
			return
		}
		api.SendResponse(w, http.StatusInternalServerError, "Failed to generate report", nil, "An unexpected error occurred")
		return
	}

	// Set appropriate headers based on format
	switch format {
	case "pdf":
		w.Header().Set("Content-Type", "application/pdf")
		w.Header().Set("Content-Disposition", "attachment; filename=weekly_sales_report.pdf")
	case "excel":
		w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
		w.Header().Set("Content-Disposition", "attachment; filename=weekly_sales_report.xlsx")
	default:
		w.Header().Set("Content-Type", "application/json")
	}

	// Write the report to the response
	w.Write(report)
}

func (h *SalesHandler) GetMonthlySalesReport(w http.ResponseWriter, r *http.Request) {
	// Extract query parameters
	yearStr := r.URL.Query().Get("year")
	monthStr := r.URL.Query().Get("month")
	format := r.URL.Query().Get("format")

	// Validate and parse year
	year, err := strconv.Atoi(yearStr)
	if err != nil {
		api.SendResponse(w, http.StatusBadRequest, "Invalid year", nil, "Year must be a valid number")
		return
	}

	// Validate and parse month
	month, err := strconv.Atoi(monthStr)
	if err != nil || month < 1 || month > 12 {
		api.SendResponse(w, http.StatusBadRequest, "Invalid month", nil, "Month must be a number between 1 and 12")
		return
	}

	// Validate format
	if format == "" {
		format = "json" // Default format
	}
	if format != "json" && format != "pdf" && format != "excel" {
		api.SendResponse(w, http.StatusBadRequest, "Invalid format", nil, "Supported formats: json, pdf, excel")
		return
	}

	// Call use case
	report, err := h.salesUseCase.GenerateMonthlySalesReport(r.Context(), year, time.Month(month), format)
	if err != nil {
		if err == utils.ErrNoDataFound {
			api.SendResponse(w, http.StatusNotFound, "No sales data for the specified month", nil, "")
			return
		}
		api.SendResponse(w, http.StatusInternalServerError, "Failed to generate report", nil, "An unexpected error occurred")
		return
	}

	// Set appropriate headers based on format
	switch format {
	case "pdf":
		w.Header().Set("Content-Type", "application/pdf")
		w.Header().Set("Content-Disposition", "attachment; filename=monthly_sales_report.pdf")
	case "excel":
		w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
		w.Header().Set("Content-Disposition", "attachment; filename=monthly_sales_report.xlsx")
	default:
		w.Header().Set("Content-Type", "application/json")
	}

	// Write the report to the response
	w.Write(report)
}

func (h *SalesHandler) GetCustomSalesReport(w http.ResponseWriter, r *http.Request) {
	// Extract query parameters
	startDateStr := r.URL.Query().Get("start_date")
	endDateStr := r.URL.Query().Get("end_date")
	format := r.URL.Query().Get("format")

	// Validate and parse dates
	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		api.SendResponse(w, http.StatusBadRequest, "Invalid start date", nil, "Use YYYY-MM-DD format")
		return
	}

	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		api.SendResponse(w, http.StatusBadRequest, "Invalid end date", nil, "Use YYYY-MM-DD format")
		return
	}

	// Validate format
	if format == "" {
		format = "json" // Default format
	}
	if format != "json" && format != "pdf" && format != "excel" {
		api.SendResponse(w, http.StatusBadRequest, "Invalid format", nil, "Supported formats: json, pdf, excel")
		return
	}

	// Call use case
	report, err := h.salesUseCase.GenerateCustomSalesReport(r.Context(), startDate, endDate, format)
	if err != nil {
		switch err {
		case utils.ErrNoDataFound:
			api.SendResponse(w, http.StatusNotFound, "No sales data for the specified period", nil, "")
		case utils.ErrInvalidDateRange:
			api.SendResponse(w, http.StatusBadRequest, "Invalid date range", nil, "End date must be after start date")
		case utils.ErrFutureDateRange:
			api.SendResponse(w, http.StatusBadRequest, "Invalid date range", nil, "Cannot generate report for future dates")
		default:
			api.SendResponse(w, http.StatusInternalServerError, "Failed to generate report", nil, "An unexpected error occurred")
		}
		return
	}

	// Set appropriate headers based on format
	switch format {
	case "pdf":
		w.Header().Set("Content-Type", "application/pdf")
		w.Header().Set("Content-Disposition", "attachment; filename=custom_sales_report.pdf")
	case "excel":
		w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
		w.Header().Set("Content-Disposition", "attachment; filename=custom_sales_report.xlsx")
	default:
		w.Header().Set("Content-Type", "application/json")
	}

	// Write the report to the response
	w.Write(report)
}
