package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"time"

	"github.com/mohamedfawas/rmshop-clean-architecture/internal/domain"
	"github.com/mohamedfawas/rmshop-clean-architecture/internal/repository"
	utils "github.com/mohamedfawas/rmshop-clean-architecture/pkg/sales_report"
	utilsVars "github.com/mohamedfawas/rmshop-clean-architecture/pkg/utils"
)

var ()

type SalesUseCase interface {
	GenerateDailySalesReport(ctx context.Context, date time.Time, format string) ([]byte, error)
	GenerateWeeklySalesReport(ctx context.Context, startDate time.Time, format string) ([]byte, error)
	GenerateMonthlySalesReport(ctx context.Context, year int, month time.Month, format string) ([]byte, error)
	GenerateCustomSalesReport(ctx context.Context, startDate, endDate time.Time, format string) ([]byte, error)
}

type salesUseCase struct {
	salesRepo repository.SalesRepository
}

func NewSalesUseCase(salesRepo repository.SalesRepository) SalesUseCase {
	return &salesUseCase{salesRepo: salesRepo}
}

func (u *salesUseCase) GenerateDailySalesReport(ctx context.Context, date time.Time, format string) ([]byte, error) {
	// Get sales data from repository
	salesData, err := u.salesRepo.GetDailySalesData(ctx, date)
	if err != nil {
		log.Printf("error while retrieving daily sales data : %v", err)
		return nil, err
	}

	// If no sales data retrieved
	if len(salesData) == 0 {
		return nil, utilsVars.ErrNoDataFound
	}

	// Generate report based on format
	switch format {
	case "json":
		return utils.GenerateJSONReport(salesData)
	case "pdf":
		return utils.GeneratePDFReport(salesData)
	case "excel":
		return utils.GenerateExcelReport(salesData)
	default:
		return nil, utilsVars.ErrInvalidFormat
	}
}

func (u *salesUseCase) GenerateWeeklySalesReport(ctx context.Context, startDate time.Time, format string) ([]byte, error) {
	// Get sales data from repository
	salesData, err := u.salesRepo.GetWeeklySalesData(ctx, startDate)
	if err != nil {
		return nil, err
	}

	// If no sales data retrieved
	if len(salesData) == 0 {
		return nil, utilsVars.ErrNoDataFound
	}

	// Calculate weekly totals
	var totalOrders int
	var totalAmount float64
	var totalCouponOrders int
	for _, day := range salesData {
		totalOrders += day.OrderCount
		totalAmount += day.TotalAmount
		totalCouponOrders += day.CouponOrderCount
	}

	weeklyReport := domain.WeeklySalesReport{
		StartDate:             startDate,
		EndDate:               startDate.AddDate(0, 0, 6),
		DailySales:            salesData,
		TotalOrderCount:       totalOrders,
		TotalAmount:           totalAmount,
		TotalCouponOrderCount: totalCouponOrders,
	}

	// Generate report based on format
	switch format {
	case "json":
		return json.Marshal(weeklyReport)
	case "pdf":
		return utils.GenerateWeeklyPDFReport(weeklyReport)
	case "excel":
		return utils.GenerateWeeklyExcelReport(weeklyReport)
	default:
		return nil, utilsVars.ErrInvalidFormat
	}
}

func (u *salesUseCase) GenerateMonthlySalesReport(ctx context.Context, year int, month time.Month, format string) ([]byte, error) {
	// Validate input
	now := time.Now()
	if year > now.Year() || (year == now.Year() && month > now.Month()) {
		return nil, errors.New("cannot generate report for future dates")
	}

	// Get sales data from repository
	salesData, err := u.salesRepo.GetMonthlySalesData(ctx, year, month)
	if err != nil {
		return nil, err
	}

	// If no sales data retrieved
	if len(salesData) == 0 {
		return nil, utilsVars.ErrNoDataFound
	}

	// Calculate monthly totals
	var totalOrders int
	var totalAmount float64
	var totalCouponOrders int
	for _, day := range salesData {
		totalOrders += day.OrderCount
		totalAmount += day.TotalAmount
		totalCouponOrders += day.CouponOrderCount
	}

	monthlyReport := domain.MonthlySalesReport{
		Year:                  year,
		Month:                 month,
		DailySales:            salesData,
		TotalOrderCount:       totalOrders,
		TotalAmount:           totalAmount,
		TotalCouponOrderCount: totalCouponOrders,
	}

	// Generate report based on format
	switch format {
	case "json":
		return json.Marshal(monthlyReport)
	case "pdf":
		return utils.GenerateMonthlyPDFReport(monthlyReport)
	case "excel":
		return utils.GenerateMonthlyExcelReport(monthlyReport)
	default:
		return nil, utilsVars.ErrInvalidFormat
	}
}

func (u *salesUseCase) GenerateCustomSalesReport(ctx context.Context, startDate, endDate time.Time, format string) ([]byte, error) {
	// Validate date range
	if endDate.Before(startDate) {
		return nil, utilsVars.ErrInvalidDateRange
	}

	// Check if date range is in the future
	if startDate.After(time.Now()) {
		return nil, utilsVars.ErrFutureDateRange
	}

	// Get sales data from repository
	salesData, err := u.salesRepo.GetCustomSalesData(ctx, startDate, endDate)
	if err != nil {
		return nil, err
	}

	// If no sales data retrieved
	if len(salesData) == 0 {
		return nil, utilsVars.ErrNoDataFound
	}

	// Calculate totals
	var totalOrders, totalCouponOrders int
	var totalAmount float64
	for _, day := range salesData {
		totalOrders += day.OrderCount
		totalAmount += day.TotalAmount
		totalCouponOrders += day.CouponOrderCount
	}

	report := domain.CustomSalesReport{
		StartDate:             startDate,
		EndDate:               endDate,
		DailySales:            salesData,
		TotalOrderCount:       totalOrders,
		TotalAmount:           totalAmount,
		TotalCouponOrderCount: totalCouponOrders,
	}

	// Generate report based on format
	switch format {
	case "json":
		return json.Marshal(report)
	case "pdf":
		return utils.GenerateCustomPDFReport(report)
	case "excel":
		return utils.GenerateCustomExcelReport(report)
	default:
		return nil, utilsVars.ErrInvalidFormat
	}
}
