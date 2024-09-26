package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
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
	// Create a new PDF document in portrait ("P") mode with millimeter ("mm") units, and "A4" paper size.
	pdf := gofpdf.New("P", "mm", "A4", "")

	// Add a new page to the PDF document.
	pdf.AddPage()

	// Set the font to "Arial", bold ("B"), with a font size of 16, for the title.
	pdf.SetFont("Arial", "B", 16)
	// Set the text color to black (RGB: 0, 0, 0).
	pdf.SetTextColor(0, 0, 0)

	// Add the title of the report, centered ("C"), with a cell width of 190mm and height of 10mm.
	pdf.CellFormat(190, 10, "Daily Sales Report", "", 0, "C", false, 0, "")
	// Move the cursor down 15mm for spacing after the title.
	pdf.Ln(15)

	// Set the font to "Arial", regular (""), with a font size of 12 for the date.
	pdf.SetFont("Arial", "", 12)
	// Add the date of the report (format: YYYY-MM-DD), left-aligned ("L").
	pdf.CellFormat(190, 10, fmt.Sprintf("Date: %s", data[0].Date.Format("2006-01-02")), "", 0, "L", false, 0, "")
	// Move the cursor down 10mm for spacing after the date.
	pdf.Ln(10)

	// Set the font to "Arial", bold ("B"), with a font size of 12 for the summary table header.
	pdf.SetFont("Arial", "B", 12)
	// Set the background fill color for the table header (RGB: 200, 220, 255).
	pdf.SetFillColor(200, 220, 255)
	// Add table headers for "Total Orders", "Total Amount", and "Orders with Coupons", with border and fill.
	pdf.CellFormat(60, 10, "Total Orders", "1", 0, "L", true, 0, "")
	pdf.CellFormat(60, 10, "Total Amount", "1", 0, "L", true, 0, "")
	pdf.CellFormat(70, 10, "Orders with Coupons", "1", 1, "L", true, 0, "") // The 1 at the end breaks to a new line.

	// Set the font to "Arial", regular, with a font size of 12 for the table data.
	pdf.SetFont("Arial", "", 12)

	// Add data for total orders, total amount, and coupon orders.
	pdf.CellFormat(60, 10, fmt.Sprintf("%d", data[0].OrderCount), "1", 0, "L", false, 0, "")
	pdf.CellFormat(60, 10, fmt.Sprintf("$%.2f", data[0].TotalAmount), "1", 0, "L", false, 0, "")
	pdf.CellFormat(70, 10, fmt.Sprintf("%d", data[0].CouponOrderCount), "1", 1, "L", false, 0, "")

	// Move the cursor down 10mm for spacing after the table.
	pdf.Ln(10)

	// Define some key sales metrics to be shown in the report.
	metrics := []struct {
		name  string
		value string
	}{
		{"Average Order Value", fmt.Sprintf("$%.2f", data[0].TotalAmount/float64(data[0].OrderCount))},
		{"Coupon Usage Rate", fmt.Sprintf("%.2f%%", float64(data[0].CouponOrderCount)/float64(data[0].OrderCount)*100)},
	}

	// Set the font to "Arial", bold, with a font size of 14 for the "Key Metrics" header.
	pdf.SetFont("Arial", "B", 14)

	// Add the "Key Metrics" title.
	pdf.CellFormat(190, 10, "Key Metrics", "", 1, "L", false, 0, "")

	// Set the font back to regular for the key metrics content.
	pdf.SetFont("Arial", "", 12)

	// Loop over the metrics and add each one to the PDF, with two cells in each row (name and value).
	for _, metric := range metrics {
		pdf.CellFormat(95, 10, metric.name, "", 0, "L", false, 0, "")
		pdf.CellFormat(95, 10, metric.value, "", 1, "L", false, 0, "")
	}

	// Set the Y position to -15mm from the bottom of the page for the footer.
	pdf.SetY(-15)
	// Set the font to "Arial", italic ("I"), with a font size of 8 for the footer.
	pdf.SetFont("Arial", "I", 8)
	// Add a footer showing the report generation timestamp, centered.
	pdf.CellFormat(0, 10, fmt.Sprintf("Report generated on %s", time.Now().Format("2006-01-02 15:04:05")), "", 0, "C", false, 0, "")

	// Create a buffer to hold the PDF content.
	var buf bytes.Buffer
	// Write the PDF content to the buffer.
	err := pdf.Output(&buf)
	// If there is an error while generating the PDF, return nil and the error.
	if err != nil {
		log.Printf("error while writing the pdf content to the buffer : %v", err)
		return nil, utils.ErrFailedToGeneratePDF
	}

	// Return the generated PDF content as a byte slice and nil for error.
	return buf.Bytes(), nil
}

// GenerateExcelReport generates an Excel report for daily sales data.
func GenerateExcelReport(data []domain.DailySales) ([]byte, error) {
	// Create a new Excel file using the "excelize" package.
	f := excelize.NewFile()
	// Define the name of the sheet where the data will be written.
	sheet := "Sheet1"

	// Set column headers in the first row of the Excel sheet.
	// "A1" refers to the first cell of column A, "B1" for column B, etc.
	f.SetCellValue(sheet, "A1", "Date")               // Column for Date
	f.SetCellValue(sheet, "B1", "Order Count")        // Column for total orders
	f.SetCellValue(sheet, "C1", "Total Amount")       // Column for total amount
	f.SetCellValue(sheet, "D1", "Coupon Order Count") // Column for orders with coupons

	// Loop through the sales data to add each record as a new row in the Excel sheet.
	for i, sale := range data {
		// Calculate the row number (starting from row 2 since row 1 contains the headers).
		row := i + 2
		// Set cell values for each field in the current row.
		// "A" column gets the date formatted as "YYYY-MM-DD".
		f.SetCellValue(sheet, "A"+strconv.Itoa(row), sale.Date.Format("2006-01-02"))
		// "B" column gets the total number of orders for that date.
		f.SetCellValue(sheet, "B"+strconv.Itoa(row), sale.OrderCount)
		// "C" column gets the total sales amount for that date.
		f.SetCellValue(sheet, "C"+strconv.Itoa(row), sale.TotalAmount)
		// "D" column gets the number of orders that used coupons for that date.
		f.SetCellValue(sheet, "D"+strconv.Itoa(row), sale.CouponOrderCount)
	}

	// Create a buffer to hold the Excel file content.
	var buf bytes.Buffer
	// Write the Excel file content to the buffer.
	err := f.Write(&buf)
	if err != nil {
		log.Printf("error while writing the excel content to the buffer: %v", err)
		return nil, err
	}

	// Return the Excel content as a byte slice and nil for error.
	return buf.Bytes(), nil
}

// GenerateWeeklyPDFReport generates a weekly sales report in PDF format.
func GenerateWeeklyPDFReport(data domain.WeeklySalesReport) ([]byte, error) {
	// Create a new PDF document with default settings (Portrait orientation, millimeter units, and A4 size).
	pdf := gofpdf.New("P", "mm", "A4", "")
	// Add a new page to the PDF
	pdf.AddPage()

	// Add title
	pdf.SetFont("Arial", "B", 16)           // Set font to Arial, bold, size 16 for the title.
	pdf.Cell(40, 10, "Weekly Sales Report") // Create a cell to hold the title text.
	pdf.Ln(10)                              // Move to the next line after the title with a 10mm line break.

	// Add date range
	pdf.SetFont("Arial", "", 12) // Set font to Arial, normal, size 12 for the date range.
	pdf.Cell(40, 10, fmt.Sprintf("From: %s To: %s",
		data.StartDate.Format("2006-01-02"),
		data.EndDate.Format("2006-01-02")))
	pdf.Ln(10) // Move to the next line after the date range with a 10mm line break.

	// Add summary
	pdf.SetFont("Arial", "B", 14)
	pdf.Cell(40, 10, "Summary")
	pdf.Ln(8)
	pdf.SetFont("Arial", "", 12)
	pdf.Cell(40, 10, fmt.Sprintf("Total Orders: %d", data.TotalOrderCount))
	pdf.Ln(8)
	pdf.Cell(40, 10, fmt.Sprintf("Total Amount: $%.2f", data.TotalAmount))
	pdf.Ln(8)
	pdf.Cell(40, 10, fmt.Sprintf("Orders with Coupons: %d", data.TotalCouponOrderCount))
	pdf.Ln(10)

	// Add daily breakdown
	pdf.SetFont("Arial", "B", 14)
	pdf.Cell(40, 10, "Daily Breakdown")
	pdf.Ln(8)
	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(40, 10, "Date")
	pdf.Cell(40, 10, "Orders")
	pdf.Cell(40, 10, "Amount")
	pdf.Cell(40, 10, "Coupon Orders")
	pdf.Ln(8)

	pdf.SetFont("Arial", "", 12)
	for _, day := range data.DailySales {
		pdf.Cell(40, 10, day.Date.Format("2006-01-02"))
		pdf.Cell(40, 10, fmt.Sprintf("%d", day.OrderCount))
		pdf.Cell(40, 10, fmt.Sprintf("$%.2f", day.TotalAmount))
		pdf.Cell(40, 10, fmt.Sprintf("%d", day.CouponOrderCount))
		pdf.Ln(8)
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

	// Add title
	f.SetCellValue(sheet, "A1", "Weekly Sales Report")
	f.SetCellValue(sheet, "A2", fmt.Sprintf("From: %s To: %s", data.StartDate.Format("2006-01-02"), data.EndDate.Format("2006-01-02")))

	// Add summary
	f.SetCellValue(sheet, "A4", "Summary")
	f.SetCellValue(sheet, "A5", "Total Orders")
	f.SetCellValue(sheet, "B5", data.TotalOrderCount)
	f.SetCellValue(sheet, "A6", "Total Amount")
	f.SetCellValue(sheet, "B6", data.TotalAmount)
	f.SetCellValue(sheet, "A7", "Orders with Coupons")
	f.SetCellValue(sheet, "B7", data.TotalCouponOrderCount)

	// Add daily breakdown
	f.SetCellValue(sheet, "A9", "Daily Breakdown")
	f.SetCellValue(sheet, "A10", "Date")
	f.SetCellValue(sheet, "B10", "Orders")
	f.SetCellValue(sheet, "C10", "Amount")
	f.SetCellValue(sheet, "D10", "Coupon Orders")

	for i, day := range data.DailySales {
		row := 11 + i
		f.SetCellValue(sheet, fmt.Sprintf("A%d", row), day.Date.Format("2006-01-02"))
		f.SetCellValue(sheet, fmt.Sprintf("B%d", row), day.OrderCount)
		f.SetCellValue(sheet, fmt.Sprintf("C%d", row), day.TotalAmount)
		f.SetCellValue(sheet, fmt.Sprintf("D%d", row), day.CouponOrderCount)
	}

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

	// Add title
	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(40, 10, fmt.Sprintf("Monthly Sales Report - %s %d", data.Month.String(), data.Year))
	pdf.Ln(10)

	// Add summary
	pdf.SetFont("Arial", "B", 14)
	pdf.Cell(40, 10, "Summary")
	pdf.Ln(8)
	pdf.SetFont("Arial", "", 12)
	pdf.Cell(40, 10, fmt.Sprintf("Total Orders: %d", data.TotalOrderCount))
	pdf.Ln(8)
	pdf.Cell(40, 10, fmt.Sprintf("Total Amount: $%.2f", data.TotalAmount))
	pdf.Ln(8)
	pdf.Cell(40, 10, fmt.Sprintf("Orders with Coupons: %d", data.TotalCouponOrderCount))
	pdf.Ln(10)

	// Add daily breakdown
	pdf.SetFont("Arial", "B", 14)
	pdf.Cell(40, 10, "Daily Breakdown")
	pdf.Ln(8)
	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(30, 10, "Date")
	pdf.Cell(30, 10, "Orders")
	pdf.Cell(40, 10, "Amount")
	pdf.Cell(40, 10, "Coupon Orders")
	pdf.Ln(8)

	pdf.SetFont("Arial", "", 12)
	for _, day := range data.DailySales {
		pdf.Cell(30, 10, day.Date.Format("2006-01-02"))
		pdf.Cell(30, 10, fmt.Sprintf("%d", day.OrderCount))
		pdf.Cell(40, 10, fmt.Sprintf("$%.2f", day.TotalAmount))
		pdf.Cell(40, 10, fmt.Sprintf("%d", day.CouponOrderCount))
		pdf.Ln(8)
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

	// Add title
	f.SetCellValue(sheet, "A1", fmt.Sprintf("Monthly Sales Report - %s %d", data.Month.String(), data.Year))

	// Add summary
	f.SetCellValue(sheet, "A3", "Summary")
	f.SetCellValue(sheet, "A4", "Total Orders")
	f.SetCellValue(sheet, "B4", data.TotalOrderCount)
	f.SetCellValue(sheet, "A5", "Total Amount")
	f.SetCellValue(sheet, "B5", data.TotalAmount)
	f.SetCellValue(sheet, "A6", "Orders with Coupons")
	f.SetCellValue(sheet, "B6", data.TotalCouponOrderCount)

	// Add daily breakdown
	f.SetCellValue(sheet, "A8", "Daily Breakdown")
	f.SetCellValue(sheet, "A9", "Date")
	f.SetCellValue(sheet, "B9", "Orders")
	f.SetCellValue(sheet, "C9", "Amount")
	f.SetCellValue(sheet, "D9", "Coupon Orders")

	for i, day := range data.DailySales {
		row := 10 + i
		f.SetCellValue(sheet, fmt.Sprintf("A%d", row), day.Date.Format("2006-01-02"))
		f.SetCellValue(sheet, fmt.Sprintf("B%d", row), day.OrderCount)
		f.SetCellValue(sheet, fmt.Sprintf("C%d", row), day.TotalAmount)
		f.SetCellValue(sheet, fmt.Sprintf("D%d", row), day.CouponOrderCount)
	}

	// Auto-fit column width
	for col := 'A'; col <= 'D'; col++ {
		f.SetColWidth(sheet, string(col), string(col), 20)
	}

	// Add footer
	lastRow := 10 + len(data.DailySales) + 2
	f.SetCellValue(sheet, fmt.Sprintf("A%d", lastRow), fmt.Sprintf("Report generated on %s", time.Now().Format("2006-01-02 15:04:05")))

	var buf bytes.Buffer
	err := f.Write(&buf)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func GenerateCustomPDFReport(data domain.CustomSalesReport) ([]byte, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

	// Add title
	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(40, 10, "Custom Sales Report")
	pdf.Ln(10)

	// Add date range
	pdf.SetFont("Arial", "", 12)
	pdf.Cell(40, 10, fmt.Sprintf("From: %s To: %s", data.StartDate.Format("2006-01-02"), data.EndDate.Format("2006-01-02")))
	pdf.Ln(10)

	// Add summary
	pdf.SetFont("Arial", "B", 14)
	pdf.Cell(40, 10, "Summary")
	pdf.Ln(8)
	pdf.SetFont("Arial", "", 12)
	pdf.Cell(40, 10, fmt.Sprintf("Total Orders: %d", data.TotalOrderCount))
	pdf.Ln(8)
	pdf.Cell(40, 10, fmt.Sprintf("Total Amount: $%.2f", data.TotalAmount))
	pdf.Ln(8)
	pdf.Cell(40, 10, fmt.Sprintf("Orders with Coupons: %d", data.TotalCouponOrderCount))
	pdf.Ln(10)

	// Add daily breakdown
	pdf.SetFont("Arial", "B", 14)
	pdf.Cell(40, 10, "Daily Breakdown")
	pdf.Ln(8)
	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(40, 10, "Date")
	pdf.Cell(30, 10, "Orders")
	pdf.Cell(40, 10, "Amount")
	pdf.Cell(40, 10, "Coupon Orders")
	pdf.Ln(8)

	pdf.SetFont("Arial", "", 12)
	for _, day := range data.DailySales {
		pdf.Cell(40, 10, day.Date.Format("2006-01-02"))
		pdf.Cell(30, 10, fmt.Sprintf("%d", day.OrderCount))
		pdf.Cell(40, 10, fmt.Sprintf("$%.2f", day.TotalAmount))
		pdf.Cell(40, 10, fmt.Sprintf("%d", day.CouponOrderCount))
		pdf.Ln(8)
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

func GenerateCustomExcelReport(data domain.CustomSalesReport) ([]byte, error) {
	f := excelize.NewFile()
	sheet := "Custom Sales Report"
	f.SetSheetName("Sheet1", sheet)

	// Add title
	f.SetCellValue(sheet, "A1", "Custom Sales Report")
	f.SetCellValue(sheet, "A2", fmt.Sprintf("From: %s To: %s", data.StartDate.Format("2006-01-02"), data.EndDate.Format("2006-01-02")))

	// Add summary
	f.SetCellValue(sheet, "A4", "Summary")
	f.SetCellValue(sheet, "A5", "Total Orders")
	f.SetCellValue(sheet, "B5", data.TotalOrderCount)
	f.SetCellValue(sheet, "A6", "Total Amount")
	f.SetCellValue(sheet, "B6", data.TotalAmount)
	f.SetCellValue(sheet, "A7", "Orders with Coupons")
	f.SetCellValue(sheet, "B7", data.TotalCouponOrderCount)

	// Add daily breakdown
	f.SetCellValue(sheet, "A9", "Daily Breakdown")
	f.SetCellValue(sheet, "A10", "Date")
	f.SetCellValue(sheet, "B10", "Orders")
	f.SetCellValue(sheet, "C10", "Amount")
	f.SetCellValue(sheet, "D10", "Coupon Orders")

	for i, day := range data.DailySales {
		row := 11 + i
		f.SetCellValue(sheet, fmt.Sprintf("A%d", row), day.Date.Format("2006-01-02"))
		f.SetCellValue(sheet, fmt.Sprintf("B%d", row), day.OrderCount)
		f.SetCellValue(sheet, fmt.Sprintf("C%d", row), day.TotalAmount)
		f.SetCellValue(sheet, fmt.Sprintf("D%d", row), day.CouponOrderCount)
	}

	// Auto-fit column width
	for col := 'A'; col <= 'D'; col++ {
		f.SetColWidth(sheet, string(col), string(col), 20)
	}

	// Add footer
	lastRow := 11 + len(data.DailySales) + 2
	f.SetCellValue(sheet, fmt.Sprintf("A%d", lastRow), fmt.Sprintf("Report generated on %s", time.Now().Format("2006-01-02 15:04:05")))

	var buf bytes.Buffer
	err := f.Write(&buf)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
