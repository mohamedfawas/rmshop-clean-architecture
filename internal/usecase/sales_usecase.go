package usecase

import (
	"bytes"
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/jung-kurt/gofpdf"
	"github.com/mohamedfawas/rmshop-clean-architecture/internal/domain"
	"github.com/mohamedfawas/rmshop-clean-architecture/internal/repository"
	"github.com/xuri/excelize/v2"
)

type SalesUseCase interface {
	GenerateSalesReport(ctx context.Context, reportType string, startDate, endDate time.Time, couponApplied bool, includeMetrics string) (*domain.SalesReport, error)
	GeneratePDFReport(report *domain.SalesReport) ([]byte, error)
	GenerateExcelReport(report *domain.SalesReport) ([]byte, error)
}

type salesUseCase struct {
	salesRepo repository.SalesRepository
}

func NewSalesUseCase(salesRepo repository.SalesRepository) SalesUseCase {
	return &salesUseCase{salesRepo: salesRepo}
}

func (u *salesUseCase) GenerateSalesReport(ctx context.Context, reportType string, startDate, endDate time.Time, couponApplied bool, includeMetrics string) (*domain.SalesReport, error) {
	dailyData, err := u.salesRepo.GetSalesData(ctx, startDate, endDate, couponApplied)
	if err != nil {
		return nil, err
	}

	report := &domain.SalesReport{
		ReportType:    reportType,
		StartDate:     startDate,
		EndDate:       endDate,
		CouponApplied: couponApplied,
	}

	// Process the report based on the report type
	switch reportType {
	case "daily":
		report.DailyData = dailyData
	case "weekly":
		report.WeeklyData = u.aggregateWeeklyData(dailyData)
	case "monthly":
		report.MonthlyData = u.aggregateMonthlyData(dailyData)
	case "yearly":
		report.YearlyData = u.aggregateYearlyData(dailyData)
	case "custom":
		report.CustomData = u.aggregateCustomData(dailyData, startDate, endDate)
	}

	u.calculateTotals(report)

	// Include additional metrics if requested
	if includeMetrics != "" {
		err = u.includeAdditionalMetrics(ctx, report, includeMetrics, startDate, endDate)
		if err != nil {
			return nil, err
		}
	}

	return report, nil
}

func (u *salesUseCase) aggregateWeeklyData(dailyData []domain.DailySalesData) []domain.WeeklySalesData {
	weeklyMap := make(map[int]domain.WeeklySalesData)

	for _, day := range dailyData {
		_, week := day.Date.ISOWeek()
		weekData := weeklyMap[week]
		weekData.WeekNumber = week
		weekData.StartDate = minDate(weekData.StartDate, day.Date)
		weekData.EndDate = maxDate(weekData.EndDate, day.Date)
		weekData.OrderCount += day.OrderCount
		weekData.TotalAmount += day.TotalAmount
		weekData.CouponOrderCount += day.CouponOrderCount
		weeklyMap[week] = weekData
	}

	var weeklyData []domain.WeeklySalesData
	for _, week := range weeklyMap {
		weeklyData = append(weeklyData, week)
	}

	sort.Slice(weeklyData, func(i, j int) bool {
		return weeklyData[i].WeekNumber < weeklyData[j].WeekNumber
	})

	return weeklyData
}

func (u *salesUseCase) aggregateMonthlyData(dailyData []domain.DailySalesData) []domain.MonthlySalesData {
	monthlyMap := make(map[int]domain.MonthlySalesData)

	for _, day := range dailyData {
		month := day.Date.Month()
		monthData := monthlyMap[int(month)]
		monthData.Month = month
		monthData.Year = day.Date.Year()
		monthData.OrderCount += day.OrderCount
		monthData.TotalAmount += day.TotalAmount
		monthData.CouponOrderCount += day.CouponOrderCount
		monthlyMap[int(month)] = monthData
	}

	var monthlyData []domain.MonthlySalesData
	for _, month := range monthlyMap {
		monthlyData = append(monthlyData, month)
	}

	sort.Slice(monthlyData, func(i, j int) bool {
		return monthlyData[i].Month < monthlyData[j].Month
	})

	return monthlyData
}

func (u *salesUseCase) aggregateYearlyData(dailyData []domain.DailySalesData) []domain.YearlySalesData {
	yearlyMap := make(map[int]domain.YearlySalesData)

	for _, day := range dailyData {
		year := day.Date.Year()
		yearData := yearlyMap[year]
		yearData.Year = year
		yearData.OrderCount += day.OrderCount
		yearData.TotalAmount += day.TotalAmount
		yearData.CouponOrderCount += day.CouponOrderCount
		yearlyMap[year] = yearData
	}

	var yearlyData []domain.YearlySalesData
	for _, year := range yearlyMap {
		yearlyData = append(yearlyData, year)
	}

	sort.Slice(yearlyData, func(i, j int) bool {
		return yearlyData[i].Year < yearlyData[j].Year
	})

	return yearlyData
}

func (u *salesUseCase) aggregateCustomData(dailyData []domain.DailySalesData, startDate, endDate time.Time) domain.CustomSalesData {
	customData := domain.CustomSalesData{
		StartDate: startDate,
		EndDate:   endDate,
	}

	for _, day := range dailyData {
		customData.OrderCount += day.OrderCount
		customData.TotalAmount += day.TotalAmount
		customData.CouponOrderCount += day.CouponOrderCount
	}

	return customData
}

func (u *salesUseCase) calculateTotals(report *domain.SalesReport) {
	switch report.ReportType {
	case "daily":
		for _, day := range report.DailyData {
			report.TotalOrderCount += day.OrderCount
			report.TotalAmount += day.TotalAmount
			report.TotalCouponOrderCount += day.CouponOrderCount
		}
	case "weekly":
		for _, week := range report.WeeklyData {
			report.TotalOrderCount += week.OrderCount
			report.TotalAmount += week.TotalAmount
			report.TotalCouponOrderCount += week.CouponOrderCount
		}
	case "monthly":
		for _, month := range report.MonthlyData {
			report.TotalOrderCount += month.OrderCount
			report.TotalAmount += month.TotalAmount
			report.TotalCouponOrderCount += month.CouponOrderCount
		}
	case "yearly":
		for _, year := range report.YearlyData {
			report.TotalOrderCount += year.OrderCount
			report.TotalAmount += year.TotalAmount
			report.TotalCouponOrderCount += year.CouponOrderCount
		}
	case "custom":
		report.TotalOrderCount = report.CustomData.OrderCount
		report.TotalAmount = report.CustomData.TotalAmount
		report.TotalCouponOrderCount = report.CustomData.CouponOrderCount
	}
}

func (u *salesUseCase) includeAdditionalMetrics(ctx context.Context, report *domain.SalesReport, metrics string, startDate, endDate time.Time) error {
	if strings.Contains(metrics, "average_order_value") {
		report.AverageOrderValue = report.TotalAmount / float64(report.TotalOrderCount)
	}

	if strings.Contains(metrics, "top_selling_products") {
		topProducts, err := u.salesRepo.GetTopSellingProducts(ctx, startDate, endDate, 10)
		if err != nil {
			return err
		}
		report.TopSellingProducts = topProducts
	}

	return nil
}

func minDate(a, b time.Time) time.Time {
	if a.IsZero() || b.Before(a) {
		return b
	}
	return a
}

func maxDate(a, b time.Time) time.Time {
	if a.After(b) {
		return a
	}
	return b
}

func (u *salesUseCase) GeneratePDFReport(report *domain.SalesReport) ([]byte, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

	// Title
	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(190, 10, "Sales Report")
	pdf.Ln(20)

	// Report details
	pdf.SetFont("Arial", "", 12)
	pdf.CellFormat(90, 10, fmt.Sprintf("Report Type: %s", report.ReportType), "", 0, "", false, 0, "")
	pdf.CellFormat(100, 10, fmt.Sprintf("Date Range: %s - %s", report.StartDate.Format("2006-01-02"), report.EndDate.Format("2006-01-02")), "", 1, "", false, 0, "")

	pdf.CellFormat(90, 10, fmt.Sprintf("Total Orders: %d", report.TotalOrderCount), "", 0, "", false, 0, "")
	pdf.CellFormat(100, 10, fmt.Sprintf("Total Amount: $%.2f", report.TotalAmount), "", 1, "", false, 0, "")

	pdf.CellFormat(90, 10, fmt.Sprintf("Coupon Orders: %d", report.TotalCouponOrderCount), "", 0, "", false, 0, "")
	if report.AverageOrderValue > 0 {
		pdf.CellFormat(100, 10, fmt.Sprintf("Avg Order Value: $%.2f", report.AverageOrderValue), "", 1, "", false, 0, "")
	}
	pdf.Ln(10)

	// Table header
	pdf.SetFillColor(200, 200, 200)
	pdf.SetFont("Arial", "B", 12)
	pdf.CellFormat(40, 10, "Date", "1", 0, "", true, 0, "")
	pdf.CellFormat(50, 10, "Order Count", "1", 0, "", true, 0, "")
	pdf.CellFormat(50, 10, "Total Amount", "1", 0, "", true, 0, "")
	pdf.CellFormat(50, 10, "Coupon Orders", "1", 1, "", true, 0, "")

	// Table content
	pdf.SetFont("Arial", "", 12)
	for _, data := range report.DailyData {
		pdf.CellFormat(40, 10, data.Date.Format("2006-01-02"), "1", 0, "", false, 0, "")
		pdf.CellFormat(50, 10, fmt.Sprintf("%d", data.OrderCount), "1", 0, "", false, 0, "")
		pdf.CellFormat(50, 10, fmt.Sprintf("$%.2f", data.TotalAmount), "1", 0, "", false, 0, "")
		pdf.CellFormat(50, 10, fmt.Sprintf("%d", data.CouponOrderCount), "1", 1, "", false, 0, "")
	}

	// Top Selling Products
	if len(report.TopSellingProducts) > 0 {
		pdf.Ln(10)
		pdf.SetFont("Arial", "B", 14)
		pdf.Cell(190, 10, "Top Selling Products")
		pdf.Ln(10)

		pdf.SetFont("Arial", "B", 12)
		pdf.CellFormat(90, 10, "Product Name", "1", 0, "", true, 0, "")
		pdf.CellFormat(50, 10, "Quantity Sold", "1", 0, "", true, 0, "")
		pdf.CellFormat(50, 10, "Revenue", "1", 1, "", true, 0, "")

		pdf.SetFont("Arial", "", 12)
		for _, product := range report.TopSellingProducts {
			pdf.CellFormat(90, 10, product.Name, "1", 0, "", false, 0, "")
			pdf.CellFormat(50, 10, fmt.Sprintf("%d", product.Quantity), "1", 0, "", false, 0, "")
			pdf.CellFormat(50, 10, fmt.Sprintf("$%.2f", product.Revenue), "1", 1, "", false, 0, "")
		}
	}

	var buf bytes.Buffer
	err := pdf.Output(&buf)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (u *salesUseCase) GenerateExcelReport(report *domain.SalesReport) ([]byte, error) {
	f := excelize.NewFile()
	sheetName := "Sales Report"
	index, err := f.NewSheet(sheetName)
	if err != nil {
		return nil, fmt.Errorf("failed to create new sheet: %w", err)
	}

	// Set column widths
	f.SetColWidth(sheetName, "A", "A", 15)
	f.SetColWidth(sheetName, "B", "D", 20)

	// Report header
	f.SetCellValue(sheetName, "A1", "Sales Report")
	f.SetCellValue(sheetName, "A2", fmt.Sprintf("Report Type: %s", report.ReportType))
	f.SetCellValue(sheetName, "A3", fmt.Sprintf("Date Range: %s - %s", report.StartDate.Format("2006-01-02"), report.EndDate.Format("2006-01-02")))
	f.SetCellValue(sheetName, "A4", fmt.Sprintf("Total Orders: %d", report.TotalOrderCount))
	f.SetCellValue(sheetName, "A5", fmt.Sprintf("Total Amount: $%.2f", report.TotalAmount))
	f.SetCellValue(sheetName, "A6", fmt.Sprintf("Coupon Orders: %d", report.TotalCouponOrderCount))
	if report.AverageOrderValue > 0 {
		f.SetCellValue(sheetName, "A7", fmt.Sprintf("Average Order Value: $%.2f", report.AverageOrderValue))
	}

	// Table header
	headerStyle, err := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"#CCCCCC"}, Pattern: 1},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create header style: %w", err)
	}
	f.SetCellStyle(sheetName, "A9", "D9", headerStyle)
	f.SetCellValue(sheetName, "A9", "Date")
	f.SetCellValue(sheetName, "B9", "Order Count")
	f.SetCellValue(sheetName, "C9", "Total Amount")
	f.SetCellValue(sheetName, "D9", "Coupon Orders")

	// Table content
	for i, data := range report.DailyData {
		row := i + 10
		f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), data.Date.Format("2006-01-02"))
		f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), data.OrderCount)
		f.SetCellValue(sheetName, fmt.Sprintf("C%d", row), data.TotalAmount)
		f.SetCellValue(sheetName, fmt.Sprintf("D%d", row), data.CouponOrderCount)
	}

	// Top Selling Products
	if len(report.TopSellingProducts) > 0 {
		topProductsStartRow := len(report.DailyData) + 12
		f.SetCellValue(sheetName, fmt.Sprintf("A%d", topProductsStartRow), "Top Selling Products")
		f.SetCellStyle(sheetName, fmt.Sprintf("A%d", topProductsStartRow+1), fmt.Sprintf("C%d", topProductsStartRow+1), headerStyle)
		f.SetCellValue(sheetName, fmt.Sprintf("A%d", topProductsStartRow+1), "Product Name")
		f.SetCellValue(sheetName, fmt.Sprintf("B%d", topProductsStartRow+1), "Quantity Sold")
		f.SetCellValue(sheetName, fmt.Sprintf("C%d", topProductsStartRow+1), "Revenue")

		for i, product := range report.TopSellingProducts {
			row := topProductsStartRow + i + 2
			f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), product.Name)
			f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), product.Quantity)
			f.SetCellValue(sheetName, fmt.Sprintf("C%d", row), product.Revenue)
		}
	}

	f.SetActiveSheet(index)

	// Write to buffer
	buffer, err := f.WriteToBuffer()
	if err != nil {
		return nil, fmt.Errorf("failed to write Excel to buffer: %w", err)
	}

	return buffer.Bytes(), nil
}
