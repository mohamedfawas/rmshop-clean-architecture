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

func (h *SalesHandler) GetSalesReport(w http.ResponseWriter, r *http.Request) {
	// Extract query parameters
	reportType := r.URL.Query().Get("report_type")
	startDate := r.URL.Query().Get("start_date")
	endDate := r.URL.Query().Get("end_date")
	date := r.URL.Query().Get("date")
	year := r.URL.Query().Get("year")
	couponApplied := r.URL.Query().Get("coupon_applied")
	includeMetrics := r.URL.Query().Get("include_metrics")
	format := r.URL.Query().Get("format")

	// Validate report type
	if !isValidReportType(reportType) {
		api.SendResponse(w, http.StatusBadRequest, "Invalid report type", nil, "Supported types: daily, weekly, monthly, yearly, custom")
		return
	}

	// Parse dates
	var startTime, endTime time.Time
	var err error

	switch reportType {
	case "daily":
		if date == "" {
			api.SendResponse(w, http.StatusBadRequest, "Missing date for daily report", nil, "Please provide a date")
			return
		}
		startTime, err = time.Parse("2006-01-02", date)
		if err != nil {
			api.SendResponse(w, http.StatusBadRequest, "Invalid date format", nil, "Use YYYY-MM-DD format")
			return
		}
		endTime = startTime.AddDate(0, 0, 1)
	case "weekly":
		startTime, endTime, err = parseWeeklyDates(startDate, endDate)
	case "monthly":
		startTime, endTime, err = parseMonthlyDates(date)
	case "yearly":
		startTime, endTime, err = parseYearlyDates(year)
	case "custom":
		startTime, endTime, err = parseCustomDates(startDate, endDate)
	}

	if err != nil {
		api.SendResponse(w, http.StatusBadRequest, "Invalid date parameters", nil, err.Error())
		return
	}

	// Parse boolean parameters
	couponAppliedBool, _ := strconv.ParseBool(couponApplied)

	// Generate the report
	report, err := h.salesUseCase.GenerateSalesReport(r.Context(), reportType, startTime, endTime, couponAppliedBool, includeMetrics)
	if err != nil {
		api.SendResponse(w, http.StatusInternalServerError, "Failed to generate report", nil, "An unexpected error occurred")
		return
	}

	// Handle different output formats
	switch format {
	case "pdf":
		pdfData, err := h.salesUseCase.GeneratePDFReport(report)
		if err != nil {
			api.SendResponse(w, http.StatusInternalServerError, "Failed to generate PDF", nil, "An unexpected error occurred")
			return
		}
		w.Header().Set("Content-Type", "application/pdf")
		w.Header().Set("Content-Disposition", "attachment; filename=sales_report.pdf")
		_, err = w.Write(pdfData)
		if err != nil {
			log.Printf("Error writing PDF data to response: %v", err)
		}
		return
	case "excel":
		excelData, err := h.salesUseCase.GenerateExcelReport(report)
		if err != nil {
			api.SendResponse(w, http.StatusInternalServerError, "Failed to generate Excel file", nil, "An unexpected error occurred")
			return
		}
		w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
		w.Header().Set("Content-Disposition", "attachment; filename=sales_report.xlsx")
		_, err = w.Write(excelData)
		if err != nil {
			log.Printf("Error writing Excel data to response: %v", err)
		}
		return
	case "json", "":
		// Default to JSON if no format is specified
		api.SendResponse(w, http.StatusOK, "Sales report generated successfully", report, "")
	default:
		api.SendResponse(w, http.StatusBadRequest, "Invalid format", nil, "Supported formats: json, pdf, excel")
	}
}

func isValidReportType(reportType string) bool {
	validTypes := map[string]bool{
		"daily":   true,
		"weekly":  true,
		"monthly": true,
		"yearly":  true,
		"custom":  true,
	}
	return validTypes[reportType]
}

func parseWeeklyDates(startDate, endDate string) (time.Time, time.Time, error) {
	start, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	end, err := time.Parse("2006-01-02", endDate)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	if end.Sub(start) > 7*24*time.Hour {
		return time.Time{}, time.Time{}, utils.ErrInvalidDateRange
	}
	return start, end.AddDate(0, 0, 1), nil
}

func parseMonthlyDates(date string) (time.Time, time.Time, error) {
	start, err := time.Parse("2006-01", date)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	end := start.AddDate(0, 1, 0)
	return start, end, nil
}

func parseYearlyDates(year string) (time.Time, time.Time, error) {
	start, err := time.Parse("2006", year)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	end := start.AddDate(1, 0, 0)
	return start, end, nil
}

func parseCustomDates(startDate, endDate string) (time.Time, time.Time, error) {
	start, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	end, err := time.Parse("2006-01-02", endDate)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	if end.Before(start) {
		return time.Time{}, time.Time{}, utils.ErrInvalidDateRange
	}
	return start, end.AddDate(0, 0, 1), nil
}
