package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/jung-kurt/gofpdf"
	"github.com/mohamedfawas/rmshop-clean-architecture/internal/domain"
	"github.com/mohamedfawas/rmshop-clean-architecture/pkg/utils"
	"github.com/xuri/excelize/v2"
)

func GenerateJSONReport(data []domain.DailySales) ([]byte, error) {
	return json.Marshal(data)
}

// GeneratePDFReport generates a pdf report for daily sales data.
func GeneratePDFReport(data []domain.DailySales) ([]byte, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

	// Add business information
	pdf.SetFont("Arial", "B", 16)
	pdf.SetTextColor(0, 0, 0)
	pdf.CellFormat(190, 10, "RM Sports Shop", "", 0, "C", false, 0, "")
	pdf.Ln(10)
	pdf.SetFont("Arial", "", 12)
	pdf.CellFormat(190, 6, "Location: Calicut", "", 0, "C", false, 0, "")
	pdf.Ln(6)
	pdf.CellFormat(190, 6, "Phone: +911234512345", "", 0, "C", false, 0, "")
	pdf.Ln(6)
	pdf.CellFormat(190, 6, "Email: rmshop@gmail.com", "", 0, "C", false, 0, "")
	pdf.Ln(15)

	pdf.SetFont("Arial", "B", 16)
	pdf.CellFormat(190, 10, "Daily Sales Report", "", 0, "C", false, 0, "")
	pdf.Ln(15)

	pdf.SetFont("Arial", "", 12)
	pdf.CellFormat(190, 10, fmt.Sprintf("Date: %s", data[0].Date.Format("2006-01-02")), "", 0, "L", false, 0, "")
	pdf.Ln(10)

	pdf.SetFont("Arial", "B", 12)
	pdf.SetFillColor(200, 220, 255)
	pdf.CellFormat(60, 10, "Total Orders", "1", 0, "L", true, 0, "")
	pdf.CellFormat(60, 10, "Total Amount", "1", 0, "L", true, 0, "")
	pdf.CellFormat(70, 10, "Orders with Coupons", "1", 1, "L", true, 0, "")

	pdf.SetFont("Arial", "", 12)
	pdf.CellFormat(60, 10, fmt.Sprintf("%d", data[0].OrderCount), "1", 0, "L", false, 0, "")
	pdf.CellFormat(60, 10, fmt.Sprintf("$%.2f", data[0].TotalAmount), "1", 0, "L", false, 0, "")
	pdf.CellFormat(70, 10, fmt.Sprintf("%d", data[0].CouponOrderCount), "1", 1, "L", false, 0, "")

	pdf.Ln(10)

	// Key Metrics Table
	pdf.SetFont("Arial", "B", 14)
	pdf.CellFormat(190, 10, "Key Metrics", "", 1, "L", false, 0, "")
	pdf.Ln(5)

	pdf.SetFont("Arial", "B", 12)
	pdf.SetFillColor(240, 240, 240)
	pdf.CellFormat(95, 10, "Metric", "1", 0, "C", true, 0, "")
	pdf.CellFormat(95, 10, "Value", "1", 1, "C", true, 0, "")

	pdf.SetFont("Arial", "", 12)
	metrics := []struct {
		name  string
		value string
	}{
		{"Average Order Value", fmt.Sprintf("$%.2f", data[0].TotalAmount/float64(data[0].OrderCount))},
		{"Coupon Usage Rate", fmt.Sprintf("%.2f%%", float64(data[0].CouponOrderCount)/float64(data[0].OrderCount)*100)},
	}

	for _, metric := range metrics {
		pdf.CellFormat(95, 10, metric.name, "1", 0, "L", false, 0, "")
		pdf.CellFormat(95, 10, metric.value, "1", 1, "L", false, 0, "")
	}

	pdf.SetY(-15)
	pdf.SetFont("Arial", "I", 8)
	pdf.CellFormat(0, 10, fmt.Sprintf("Report generated on %s", time.Now().Format("2006-01-02 15:04:05")), "", 0, "C", false, 0, "")

	var buf bytes.Buffer
	err := pdf.Output(&buf)
	if err != nil {
		log.Printf("error while writing the pdf content to the buffer : %v", err)
		return nil, utils.ErrFailedToGeneratePDF
	}

	return buf.Bytes(), nil
}

// GenerateExcelReport generates an Excel report for daily sales data.
func GenerateExcelReport(data []domain.DailySales) ([]byte, error) {
	f := excelize.NewFile()
	sheet := "Sheet1"

	// Add business information
	f.SetCellValue(sheet, "A1", "RM Sports Shop")
	f.SetCellValue(sheet, "A2", "Location: Calicut")
	f.SetCellValue(sheet, "A3", "Phone: +911234512345")
	f.SetCellValue(sheet, "A4", "Email: rmshop@gmail.com")

	// Set column headers
	f.SetCellValue(sheet, "A6", "Date")
	f.SetCellValue(sheet, "B6", "Order Count")
	f.SetCellValue(sheet, "C6", "Total Amount")
	f.SetCellValue(sheet, "D6", "Coupon Order Count")

	// Add data
	for i, sale := range data {
		row := i + 7
		f.SetCellValue(sheet, fmt.Sprintf("A%d", row), sale.Date.Format("2006-01-02"))
		f.SetCellValue(sheet, fmt.Sprintf("B%d", row), sale.OrderCount)
		f.SetCellValue(sheet, fmt.Sprintf("C%d", row), sale.TotalAmount)
		f.SetCellValue(sheet, fmt.Sprintf("D%d", row), sale.CouponOrderCount)
	}

	// Add key metrics
	lastRow := len(data) + 8
	f.SetCellValue(sheet, fmt.Sprintf("A%d", lastRow), "Key Metrics")
	f.SetCellValue(sheet, fmt.Sprintf("A%d", lastRow+1), "Average Order Value")
	f.SetCellValue(sheet, fmt.Sprintf("B%d", lastRow+1), data[0].TotalAmount/float64(data[0].OrderCount))
	f.SetCellValue(sheet, fmt.Sprintf("A%d", lastRow+2), "Coupon Usage Rate")
	f.SetCellValue(sheet, fmt.Sprintf("B%d", lastRow+2), float64(data[0].CouponOrderCount)/float64(data[0].OrderCount))

	var buf bytes.Buffer
	err := f.Write(&buf)
	if err != nil {
		log.Printf("error while writing the excel content to the buffer: %v", err)
		return nil, err
	}

	return buf.Bytes(), nil
}

func GenerateWeeklyPDFReport(data domain.WeeklySalesReport) ([]byte, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

	// Add title and business info
	pdf.SetFont("Arial", "B", 16)
	pdf.CellFormat(190, 10, "Weekly Sales Report", "", 1, "C", false, 0, "")
	pdf.Ln(5)

	pdf.SetFont("Arial", "", 12)
	pdf.CellFormat(190, 8, "RM Sports Shop", "", 1, "L", false, 0, "")
	pdf.CellFormat(190, 8, "Location: Calicut", "", 1, "L", false, 0, "")
	pdf.CellFormat(190, 8, "Phone: +911234512345", "", 1, "L", false, 0, "")
	pdf.CellFormat(190, 8, "Email: rmshop@gmail.com", "", 1, "L", false, 0, "")
	pdf.Ln(5)

	// Add date range
	pdf.SetFont("Arial", "B", 12)
	pdf.CellFormat(190, 10, fmt.Sprintf("From: %s To: %s", data.StartDate.Format("2006-01-02"), data.EndDate.Format("2006-01-02")), "", 1, "L", false, 0, "")
	pdf.Ln(5)

	// Add summary table
	pdf.SetFont("Arial", "B", 14)
	pdf.CellFormat(190, 10, "Summary", "", 1, "L", false, 0, "")
	pdf.Ln(5)

	// Table header
	pdf.SetFillColor(200, 220, 255) // Light blue
	pdf.SetFont("Arial", "B", 12)
	pdf.CellFormat(95, 10, "Metric", "1", 0, "C", true, 0, "")
	pdf.CellFormat(95, 10, "Value", "1", 1, "C", true, 0, "")

	// Table content
	pdf.SetFont("Arial", "", 12)
	pdf.SetFillColor(255, 255, 255) // White
	pdf.CellFormat(95, 10, "Total Orders", "1", 0, "L", true, 0, "")
	pdf.CellFormat(95, 10, fmt.Sprintf("%d", data.TotalOrderCount), "1", 1, "R", true, 0, "")
	pdf.CellFormat(95, 10, "Total Amount", "1", 0, "L", true, 0, "")
	pdf.CellFormat(95, 10, fmt.Sprintf("$%.2f", data.TotalAmount), "1", 1, "R", true, 0, "")
	pdf.CellFormat(95, 10, "Orders with Coupons", "1", 0, "L", true, 0, "")
	pdf.CellFormat(95, 10, fmt.Sprintf("%d", data.TotalCouponOrderCount), "1", 1, "R", true, 0, "")
	pdf.Ln(10)

	// Add daily breakdown table
	pdf.SetFont("Arial", "B", 14)
	pdf.CellFormat(190, 10, "Daily Breakdown", "", 1, "L", false, 0, "")
	pdf.Ln(5)

	// Table header
	pdf.SetFillColor(200, 220, 255) // Light blue
	pdf.SetFont("Arial", "B", 12)
	pdf.CellFormat(47.5, 10, "Date", "1", 0, "C", true, 0, "")
	pdf.CellFormat(47.5, 10, "Orders", "1", 0, "C", true, 0, "")
	pdf.CellFormat(47.5, 10, "Amount", "1", 0, "C", true, 0, "")
	pdf.CellFormat(47.5, 10, "Coupon Orders", "1", 1, "C", true, 0, "")

	// Table content
	pdf.SetFont("Arial", "", 12)
	for i, day := range data.DailySales {
		fillColor := 255 // White
		if i%2 == 1 {
			fillColor = 240 // Light gray for alternating rows
		}
		pdf.SetFillColor(fillColor, fillColor, fillColor) // Now using int values directly

		pdf.CellFormat(47.5, 10, day.Date.Format("2006-01-02"), "1", 0, "L", true, 0, "")
		pdf.CellFormat(47.5, 10, fmt.Sprintf("%d", day.OrderCount), "1", 0, "R", true, 0, "")
		pdf.CellFormat(47.5, 10, fmt.Sprintf("$%.2f", day.TotalAmount), "1", 0, "R", true, 0, "")
		pdf.CellFormat(47.5, 10, fmt.Sprintf("%d", day.CouponOrderCount), "1", 1, "R", true, 0, "")
	}

	var buf bytes.Buffer
	err := pdf.Output(&buf)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func GenerateWeeklyExcelReport(data domain.WeeklySalesReport) ([]byte, error) {
	f := excelize.NewFile()
	sheet := "Weekly Sales Report"
	f.SetSheetName("Sheet1", sheet)

	// Add title and business info
	f.SetCellValue(sheet, "A1", "Weekly Sales Report")
	f.MergeCell(sheet, "A1", "D1")
	f.SetCellValue(sheet, "A2", "RM Sports Shop")
	f.SetCellValue(sheet, "A3", "Location: Calicut")
	f.SetCellValue(sheet, "A4", "Phone: +911234512345")
	f.SetCellValue(sheet, "A5", "Email: rmshop@gmail.com")

	// Add date range
	f.SetCellValue(sheet, "A7", fmt.Sprintf("From: %s To: %s", data.StartDate.Format("2006-01-02"), data.EndDate.Format("2006-01-02")))
	f.MergeCell(sheet, "A7", "D7")

	// Add summary
	f.SetCellValue(sheet, "A9", "Summary")
	f.MergeCell(sheet, "A9", "D9")
	f.SetCellValue(sheet, "A10", "Metric")
	f.SetCellValue(sheet, "B10", "Value")
	f.SetCellValue(sheet, "A11", "Total Orders")
	f.SetCellValue(sheet, "B11", data.TotalOrderCount)
	f.SetCellValue(sheet, "A12", "Total Amount")
	f.SetCellValue(sheet, "B12", data.TotalAmount)
	f.SetCellValue(sheet, "A13", "Orders with Coupons")
	f.SetCellValue(sheet, "B13", data.TotalCouponOrderCount)

	// Add daily breakdown
	f.SetCellValue(sheet, "A15", "Daily Breakdown")
	f.MergeCell(sheet, "A15", "D15")
	f.SetCellValue(sheet, "A16", "Date")
	f.SetCellValue(sheet, "B16", "Orders")
	f.SetCellValue(sheet, "C16", "Amount")
	f.SetCellValue(sheet, "D16", "Coupon Orders")

	for i, day := range data.DailySales {
		row := 17 + i
		f.SetCellValue(sheet, fmt.Sprintf("A%d", row), day.Date.Format("2006-01-02"))
		f.SetCellValue(sheet, fmt.Sprintf("B%d", row), day.OrderCount)
		f.SetCellValue(sheet, fmt.Sprintf("C%d", row), day.TotalAmount)
		f.SetCellValue(sheet, fmt.Sprintf("D%d", row), day.CouponOrderCount)
	}

	// Apply styles
	titleStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 16},
		Alignment: &excelize.Alignment{Horizontal: "center"},
	})
	f.SetCellStyle(sheet, "A1", "D1", titleStyle)

	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"#C8DCFF"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center"},
		Border: []excelize.Border{{Type: "top", Color: "000000", Style: 1},
			{Type: "left", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1}},
	})
	f.SetCellStyle(sheet, "A10", "B10", headerStyle)
	f.SetCellStyle(sheet, "A16", "D16", headerStyle)

	contentStyle, _ := f.NewStyle(&excelize.Style{
		Border: []excelize.Border{{Type: "top", Color: "000000", Style: 1},
			{Type: "left", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1}},
	})
	f.SetCellStyle(sheet, "A11", "B13", contentStyle)
	f.SetCellStyle(sheet, "A17", fmt.Sprintf("D%d", 16+len(data.DailySales)), contentStyle)

	// Set column widths
	f.SetColWidth(sheet, "A", "D", 20)

	var buf bytes.Buffer
	err := f.Write(&buf)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func GenerateMonthlyPDFReport(data domain.MonthlySalesReport) ([]byte, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

	// Add business info
	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(0, 10, "RM Sports Shop")
	pdf.Ln(8)
	pdf.SetFont("Arial", "", 12)
	pdf.Cell(0, 6, "Location: Calicut")
	pdf.Ln(6)
	pdf.Cell(0, 6, "Phone: +911234512345")
	pdf.Ln(6)
	pdf.Cell(0, 6, "Email: rmshop@gmail.com")
	pdf.Ln(15)

	// Add title
	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(0, 10, fmt.Sprintf("Monthly Sales Report - %s %d", data.Month.String(), data.Year))
	pdf.Ln(15)

	// Add summary table
	pdf.SetFont("Arial", "B", 14)
	pdf.Cell(0, 10, "Summary")
	pdf.Ln(10)

	// Define table structure
	summaryTable := [][]string{
		{"Metric", "Value"},
		{"Total Orders", fmt.Sprintf("%d", data.TotalOrderCount)},
		{"Total Amount", fmt.Sprintf("$%.2f", data.TotalAmount)},
		{"Orders with Coupons", fmt.Sprintf("%d", data.TotalCouponOrderCount)},
	}

	// Set table header colors
	pdf.SetFillColor(200, 220, 255)
	pdf.SetTextColor(0, 0, 0)
	pdf.SetFont("Arial", "B", 12)

	// Create summary table
	for i, row := range summaryTable {
		for _, col := range row {
			if i == 0 {
				pdf.SetFillColor(200, 220, 255)
			} else {
				pdf.SetFillColor(255, 255, 255)
			}
			pdf.CellFormat(60, 8, col, "1", 0, "LM", true, 0, "")
		}
		pdf.Ln(-1)
	}
	pdf.Ln(10)

	// Add daily breakdown table
	pdf.SetFont("Arial", "B", 14)
	pdf.Cell(0, 10, "Daily Breakdown")
	pdf.Ln(10)

	// Define daily breakdown table structure
	breakdownHeader := []string{"Date", "Orders", "Amount", "Coupon Orders"}
	pdf.SetFillColor(200, 220, 255)
	pdf.SetTextColor(0, 0, 0)
	pdf.SetFont("Arial", "B", 12)

	// Create daily breakdown table header
	for _, col := range breakdownHeader {
		pdf.CellFormat(45, 8, col, "1", 0, "CM", true, 0, "")
	}
	pdf.Ln(-1)

	// Create daily breakdown table content
	pdf.SetFont("Arial", "", 12)
	pdf.SetFillColor(255, 255, 255)
	for _, day := range data.DailySales {
		pdf.CellFormat(45, 8, day.Date.Format("2006-01-02"), "1", 0, "LM", true, 0, "")
		pdf.CellFormat(45, 8, fmt.Sprintf("%d", day.OrderCount), "1", 0, "RM", true, 0, "")
		pdf.CellFormat(45, 8, fmt.Sprintf("$%.2f", day.TotalAmount), "1", 0, "RM", true, 0, "")
		pdf.CellFormat(45, 8, fmt.Sprintf("%d", day.CouponOrderCount), "1", 0, "RM", true, 0, "")
		pdf.Ln(-1)
	}

	// Add footer
	pdf.SetY(-15)
	pdf.SetFont("Arial", "I", 8)
	pdf.Cell(0, 10, fmt.Sprintf("Report generated on %s", time.Now().Format("2006-01-02 15:04:05")))

	var buf bytes.Buffer
	err := pdf.Output(&buf)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func GenerateMonthlyExcelReport(data domain.MonthlySalesReport) ([]byte, error) {
	f := excelize.NewFile()
	sheet := "Monthly Sales Report"
	f.SetSheetName("Sheet1", sheet)

	// Add business info
	f.SetCellValue(sheet, "A1", "RM Sports Shop")
	f.SetCellValue(sheet, "A2", "Location: Calicut")
	f.SetCellValue(sheet, "A3", "Phone: +911234512345")
	f.SetCellValue(sheet, "A4", "Email: rmshop@gmail.com")

	// Add title
	f.SetCellValue(sheet, "A6", fmt.Sprintf("Monthly Sales Report - %s %d", data.Month.String(), data.Year))

	// Add summary table
	f.SetCellValue(sheet, "A8", "Summary")
	summaryData := [][]interface{}{
		{"Metric", "Value"},
		{"Total Orders", data.TotalOrderCount},
		{"Total Amount", data.TotalAmount},
		{"Orders with Coupons", data.TotalCouponOrderCount},
	}

	// Create summary table
	for i, row := range summaryData {
		for j, cellValue := range row {
			cell, _ := excelize.CoordinatesToCellName(j+1, i+9)
			f.SetCellValue(sheet, cell, cellValue)
		}
	}

	// Format summary table
	summaryStyle, _ := f.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
		},
	})
	f.SetCellStyle(sheet, "A9", "B12", summaryStyle)

	summaryHeaderStyle, _ := f.NewStyle(&excelize.Style{
		Fill: excelize.Fill{Type: "pattern", Color: []string{"C8E6FF"}, Pattern: 1},
		Font: &excelize.Font{Bold: true},
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
		},
	})
	f.SetCellStyle(sheet, "A9", "B9", summaryHeaderStyle)

	// Add daily breakdown table
	f.SetCellValue(sheet, "A14", "Daily Breakdown")
	breakdownHeaders := []string{"Date", "Orders", "Amount", "Coupon Orders"}
	for i, header := range breakdownHeaders {
		cell, _ := excelize.CoordinatesToCellName(i+1, 15)
		f.SetCellValue(sheet, cell, header)
	}

	// Create daily breakdown table
	for i, day := range data.DailySales {
		row := 16 + i
		f.SetCellValue(sheet, fmt.Sprintf("A%d", row), day.Date.Format("2006-01-02"))
		f.SetCellValue(sheet, fmt.Sprintf("B%d", row), day.OrderCount)
		f.SetCellValue(sheet, fmt.Sprintf("C%d", row), day.TotalAmount)
		f.SetCellValue(sheet, fmt.Sprintf("D%d", row), day.CouponOrderCount)
	}

	// Format daily breakdown table
	lastRow := 15 + len(data.DailySales)
	breakdownStyle, _ := f.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
		},
	})
	f.SetCellStyle(sheet, "A15", fmt.Sprintf("D%d", lastRow), breakdownStyle)

	breakdownHeaderStyle, _ := f.NewStyle(&excelize.Style{
		Fill: excelize.Fill{Type: "pattern", Color: []string{"C8E6FF"}, Pattern: 1},
		Font: &excelize.Font{Bold: true},
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
		},
	})
	f.SetCellStyle(sheet, "A15", "D15", breakdownHeaderStyle)

	// Auto-fit column width
	for col := 'A'; col <= 'D'; col++ {
		f.SetColWidth(sheet, string(col), string(col), 20)
	}

	// Add footer
	footerRow := lastRow + 2
	f.SetCellValue(sheet, fmt.Sprintf("A%d", footerRow), fmt.Sprintf("Report generated on %s", time.Now().Format("2006-01-02 15:04:05")))

	var buf bytes.Buffer
	err := f.Write(&buf)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// func GenerateCustomPDFReport(data domain.CustomSalesReport) ([]byte, error) {
// 	pdf := gofpdf.New("P", "mm", "A4", "")
// 	pdf.AddPage()

// 	// Add title
// 	pdf.SetFont("Arial", "B", 16)
// 	pdf.Cell(40, 10, "Custom Sales Report")
// 	pdf.Ln(10)

// 	// Add date range
// 	pdf.SetFont("Arial", "", 12)
// 	pdf.Cell(40, 10, fmt.Sprintf("From: %s To: %s", data.StartDate.Format("2006-01-02"), data.EndDate.Format("2006-01-02")))
// 	pdf.Ln(10)

// 	// Add summary
// 	pdf.SetFont("Arial", "B", 14)
// 	pdf.Cell(40, 10, "Summary")
// 	pdf.Ln(8)
// 	pdf.SetFont("Arial", "", 12)
// 	pdf.Cell(40, 10, fmt.Sprintf("Total Orders: %d", data.TotalOrderCount))
// 	pdf.Ln(8)
// 	pdf.Cell(40, 10, fmt.Sprintf("Total Amount: $%.2f", data.TotalAmount))
// 	pdf.Ln(8)
// 	pdf.Cell(40, 10, fmt.Sprintf("Orders with Coupons: %d", data.TotalCouponOrderCount))
// 	pdf.Ln(10)

// 	// Add daily breakdown
// 	pdf.SetFont("Arial", "B", 14)
// 	pdf.Cell(40, 10, "Daily Breakdown")
// 	pdf.Ln(8)
// 	pdf.SetFont("Arial", "B", 12)
// 	pdf.Cell(40, 10, "Date")
// 	pdf.Cell(30, 10, "Orders")
// 	pdf.Cell(40, 10, "Amount")
// 	pdf.Cell(40, 10, "Coupon Orders")
// 	pdf.Ln(8)

// 	pdf.SetFont("Arial", "", 12)
// 	for _, day := range data.DailySales {
// 		pdf.Cell(40, 10, day.Date.Format("2006-01-02"))
// 		pdf.Cell(30, 10, fmt.Sprintf("%d", day.OrderCount))
// 		pdf.Cell(40, 10, fmt.Sprintf("$%.2f", day.TotalAmount))
// 		pdf.Cell(40, 10, fmt.Sprintf("%d", day.CouponOrderCount))
// 		pdf.Ln(8)
// 	}

// 	// Add footer
// 	pdf.SetY(-15)
// 	pdf.SetFont("Arial", "I", 8)
// 	pdf.Cell(0, 10, fmt.Sprintf("Report generated on %s", time.Now().Format("2006-01-02 15:04:05")))

// 	var buf bytes.Buffer
// 	err := pdf.Output(&buf)
// 	if err != nil {
// 		return nil, err
// 	}

// 	return buf.Bytes(), nil
// }

func GenerateCustomPDFReport(data domain.CustomSalesReport) ([]byte, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

	// Add title
	pdf.SetFont("Arial", "B", 16)
	pdf.CellFormat(190, 10, "Custom Sales Report", "", 1, "C", false, 0, "")
	pdf.Ln(5)

	// Add business information
	pdf.SetFont("Arial", "", 12)
	pdf.CellFormat(190, 7, "RM Sports Shop", "", 1, "C", false, 0, "")
	pdf.CellFormat(190, 7, "Calicut", "", 1, "C", false, 0, "")
	pdf.CellFormat(190, 7, "Phone: +911234512345", "", 1, "C", false, 0, "")
	pdf.CellFormat(190, 7, "Email: rmshop@gmail.com", "", 1, "C", false, 0, "")
	pdf.Ln(10)

	// Add date range
	pdf.SetFont("Arial", "", 12)
	pdf.CellFormat(190, 10, fmt.Sprintf("From: %s To: %s", data.StartDate.Format("2006-01-02"), data.EndDate.Format("2006-01-02")), "", 1, "C", false, 0, "")
	pdf.Ln(10)

	// Add summary table
	pdf.SetFont("Arial", "B", 14)
	pdf.CellFormat(190, 10, "Summary", "", 1, "L", false, 0, "")
	pdf.Ln(5)

	// Define summary table
	summaryTable := [][]string{
		{"Total Orders", fmt.Sprintf("%d", data.TotalOrderCount)},
		{"Total Amount", fmt.Sprintf("$%.2f", data.TotalAmount)},
		{"Orders with Coupons", fmt.Sprintf("%d", data.TotalCouponOrderCount)},
	}

	// Set colors for summary table
	pdf.SetFillColor(240, 240, 240) // Light gray background
	pdf.SetTextColor(0, 0, 0)       // Black text

	// Create summary table
	for i, row := range summaryTable {
		pdf.SetFont("Arial", "B", 12)
		pdf.CellFormat(95, 10, row[0], "1", 0, "L", true, 0, "")
		pdf.SetFont("Arial", "", 12)
		pdf.CellFormat(95, 10, row[1], "1", 1, "R", true, 0, "")
		if i%2 == 1 {
			pdf.SetFillColor(255, 255, 255) // White background for even rows
		} else {
			pdf.SetFillColor(240, 240, 240) // Light gray background for odd rows
		}
	}
	pdf.Ln(10)

	// Add daily breakdown table
	pdf.SetFont("Arial", "B", 14)
	pdf.CellFormat(190, 10, "Daily Breakdown", "", 1, "L", false, 0, "")
	pdf.Ln(5)

	// Define daily breakdown table headers
	pdf.SetFont("Arial", "B", 12)
	pdf.SetFillColor(200, 200, 200) // Darker gray for header
	pdf.CellFormat(47.5, 10, "Date", "1", 0, "C", true, 0, "")
	pdf.CellFormat(47.5, 10, "Orders", "1", 0, "C", true, 0, "")
	pdf.CellFormat(47.5, 10, "Amount", "1", 0, "C", true, 0, "")
	pdf.CellFormat(47.5, 10, "Coupon Orders", "1", 1, "C", true, 0, "")

	// Create daily breakdown table
	pdf.SetFont("Arial", "", 12)
	for i, day := range data.DailySales {
		if i%2 == 0 {
			pdf.SetFillColor(255, 255, 255) // White background for even rows
		} else {
			pdf.SetFillColor(240, 240, 240) // Light gray background for odd rows
		}
		pdf.CellFormat(47.5, 10, day.Date.Format("2006-01-02"), "1", 0, "C", true, 0, "")
		pdf.CellFormat(47.5, 10, fmt.Sprintf("%d", day.OrderCount), "1", 0, "C", true, 0, "")
		pdf.CellFormat(47.5, 10, fmt.Sprintf("$%.2f", day.TotalAmount), "1", 0, "C", true, 0, "")
		pdf.CellFormat(47.5, 10, fmt.Sprintf("%d", day.CouponOrderCount), "1", 1, "C", true, 0, "")
	}

	// Add footer
	pdf.SetY(-15)
	pdf.SetFont("Arial", "I", 8)
	pdf.CellFormat(0, 10, fmt.Sprintf("Report generated on %s", time.Now().Format("2006-01-02 15:04:05")), "", 0, "C", false, 0, "")

	var buf bytes.Buffer
	err := pdf.Output(&buf)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// func GenerateCustomExcelReport(data domain.CustomSalesReport) ([]byte, error) {
// 	f := excelize.NewFile()
// 	sheet := "Custom Sales Report"
// 	f.SetSheetName("Sheet1", sheet)

// 	// Add title
// 	f.SetCellValue(sheet, "A1", "Custom Sales Report")
// 	f.SetCellValue(sheet, "A2", fmt.Sprintf("From: %s To: %s", data.StartDate.Format("2006-01-02"), data.EndDate.Format("2006-01-02")))

// 	// Add summary
// 	f.SetCellValue(sheet, "A4", "Summary")
// 	f.SetCellValue(sheet, "A5", "Total Orders")
// 	f.SetCellValue(sheet, "B5", data.TotalOrderCount)
// 	f.SetCellValue(sheet, "A6", "Total Amount")
// 	f.SetCellValue(sheet, "B6", data.TotalAmount)
// 	f.SetCellValue(sheet, "A7", "Orders with Coupons")
// 	f.SetCellValue(sheet, "B7", data.TotalCouponOrderCount)

// 	// Add daily breakdown
// 	f.SetCellValue(sheet, "A9", "Daily Breakdown")
// 	f.SetCellValue(sheet, "A10", "Date")
// 	f.SetCellValue(sheet, "B10", "Orders")
// 	f.SetCellValue(sheet, "C10", "Amount")
// 	f.SetCellValue(sheet, "D10", "Coupon Orders")

// 	for i, day := range data.DailySales {
// 		row := 11 + i
// 		f.SetCellValue(sheet, fmt.Sprintf("A%d", row), day.Date.Format("2006-01-02"))
// 		f.SetCellValue(sheet, fmt.Sprintf("B%d", row), day.OrderCount)
// 		f.SetCellValue(sheet, fmt.Sprintf("C%d", row), day.TotalAmount)
// 		f.SetCellValue(sheet, fmt.Sprintf("D%d", row), day.CouponOrderCount)
// 	}

// 	// Auto-fit column width
// 	for col := 'A'; col <= 'D'; col++ {
// 		f.SetColWidth(sheet, string(col), string(col), 20)
// 	}

// 	// Add footer
// 	lastRow := 11 + len(data.DailySales) + 2
// 	f.SetCellValue(sheet, fmt.Sprintf("A%d", lastRow), fmt.Sprintf("Report generated on %s", time.Now().Format("2006-01-02 15:04:05")))

// 	var buf bytes.Buffer
// 	err := f.Write(&buf)
// 	if err != nil {
// 		return nil, err
// 	}

// 	return buf.Bytes(), nil
// }

func GenerateCustomExcelReport(data domain.CustomSalesReport) ([]byte, error) {
	f := excelize.NewFile()
	sheet := "Custom Sales Report"
	f.SetSheetName("Sheet1", sheet)

	// Add title and business information
	f.SetCellValue(sheet, "A1", "Custom Sales Report")
	f.SetCellValue(sheet, "A2", "RM Sports Shop")
	f.SetCellValue(sheet, "A3", "Calicut")
	f.SetCellValue(sheet, "A4", "Phone: +911234512345")
	f.SetCellValue(sheet, "A5", "Email: rmshop@gmail.com")
	f.SetCellValue(sheet, "A7", fmt.Sprintf("From: %s To: %s", data.StartDate.Format("2006-01-02"), data.EndDate.Format("2006-01-02")))

	// Create styles
	titleStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 16},
		Alignment: &excelize.Alignment{Horizontal: "center"},
	})

	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Color: "#FFFFFF"},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"#4472C4"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center"},
	})

	evenRowStyle, _ := f.NewStyle(&excelize.Style{
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"#E9EFF7"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center"},
	})

	oddRowStyle, _ := f.NewStyle(&excelize.Style{
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"#D5E2F3"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center"},
	})

	borderStyle, _ := f.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
		},
	})

	// Apply title style
	f.SetCellStyle(sheet, "A1", "A1", titleStyle)

	// Add summary
	f.SetCellValue(sheet, "A9", "Summary")
	f.SetCellStyle(sheet, "A9", "B9", headerStyle)

	summaryData := [][]interface{}{
		{"Total Orders", data.TotalOrderCount},
		{"Total Amount", data.TotalAmount},
		{"Orders with Coupons", data.TotalCouponOrderCount},
	}

	for i, row := range summaryData {
		rowNum := 10 + i
		f.SetCellValue(sheet, fmt.Sprintf("A%d", rowNum), row[0])
		f.SetCellValue(sheet, fmt.Sprintf("B%d", rowNum), row[1])
		if i%2 == 0 {
			f.SetCellStyle(sheet, fmt.Sprintf("A%d", rowNum), fmt.Sprintf("B%d", rowNum), evenRowStyle)
		} else {
			f.SetCellStyle(sheet, fmt.Sprintf("A%d", rowNum), fmt.Sprintf("B%d", rowNum), oddRowStyle)
		}
	}

	// Add daily breakdown
	f.SetCellValue(sheet, "A14", "Daily Breakdown")
	f.SetCellStyle(sheet, "A14", "D14", headerStyle)

	f.SetCellValue(sheet, "A15", "Date")
	f.SetCellValue(sheet, "B15", "Orders")
	f.SetCellValue(sheet, "C15", "Amount")
	f.SetCellValue(sheet, "D15", "Coupon Orders")
	f.SetCellStyle(sheet, "A15", "D15", headerStyle)

	for i, day := range data.DailySales {
		row := 16 + i
		f.SetCellValue(sheet, fmt.Sprintf("A%d", row), day.Date.Format("2006-01-02"))
		f.SetCellValue(sheet, fmt.Sprintf("B%d", row), day.OrderCount)
		f.SetCellValue(sheet, fmt.Sprintf("C%d", row), day.TotalAmount)
		f.SetCellValue(sheet, fmt.Sprintf("D%d", row), day.CouponOrderCount)
		if i%2 == 0 {
			f.SetCellStyle(sheet, fmt.Sprintf("A%d", row), fmt.Sprintf("D%d", row), evenRowStyle)
		} else {
			f.SetCellStyle(sheet, fmt.Sprintf("A%d", row), fmt.Sprintf("D%d", row), oddRowStyle)
		}
	}

	// Auto-fit column width
	for col := 'A'; col <= 'D'; col++ {
		f.SetColWidth(sheet, string(col), string(col), 20)
	}

	// Add footer
	lastRow := 16 + len(data.DailySales) + 2
	f.SetCellValue(sheet, fmt.Sprintf("A%d", lastRow), fmt.Sprintf("Report generated on %s", time.Now().Format("2006-01-02 15:04:05")))

	// Set overall styles
	f.SetCellStyle(sheet, "A1", fmt.Sprintf("D%d", lastRow), borderStyle)

	var buf bytes.Buffer
	err := f.Write(&buf)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
