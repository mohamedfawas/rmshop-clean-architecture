package usecase

import (
	"context"
	"sort"
	"strings"
	"time"

	"github.com/mohamedfawas/rmshop-clean-architecture/internal/domain"
	"github.com/mohamedfawas/rmshop-clean-architecture/internal/repository"
)

type SalesUseCase interface {
	GenerateSalesReport(ctx context.Context, reportType string, startDate, endDate time.Time, couponApplied bool, includeMetrics string) (*domain.SalesReport, error)
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
