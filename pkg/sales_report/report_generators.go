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
	// Initialize a new PDF document in portrait mode (P), with mm as the unit of measurement, and A4 paper size.
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage() // Add a new page to the PDF

	// --- Adding Business Information ---
	// Set the font to Arial, bold, size 16 for the title (business name)
	pdf.SetFont("Arial", "B", 16)
	pdf.SetTextColor(0, 0, 0) // Set text color to black (RGB: 0, 0, 0)

	// Add the business name to the center of the page
	pdf.CellFormat(190, 10, "RM Sports Shop", "", 0, "C", false, 0, "")
	pdf.Ln(10) // Line break (move cursor down 10 mm)

	// Set the font to Arial, regular, size 12 for additional business details
	pdf.SetFont("Arial", "", 12)
	// Add the business location, phone number, and email, all centered
	pdf.CellFormat(190, 6, "Location: Calicut", "", 0, "C", false, 0, "")
	pdf.Ln(6)
	pdf.CellFormat(190, 6, "Phone: +911234512345", "", 0, "C", false, 0, "")
	pdf.Ln(6)
	pdf.CellFormat(190, 6, "Email: rmshop@gmail.com", "", 0, "C", false, 0, "")
	pdf.Ln(15) // Extra space before the next section

	// --- Adding Report Title ---
	pdf.SetFont("Arial", "B", 16) // Set font to bold for the report title
	pdf.CellFormat(190, 10, "Daily Sales Report", "", 0, "C", false, 0, "")
	pdf.Ln(15) // Line break

	// --- Adding Report Date ---
	pdf.SetFont("Arial", "", 12) // Regular font for the report date
	// Use the date from the first record in the data to display the report date
	pdf.CellFormat(190, 10, fmt.Sprintf("Date: %s", data[0].Date.Format("2006-01-02")), "", 0, "L", false, 0, "")
	pdf.Ln(10) // Line break

	// --- Adding Summary Information ---
	pdf.SetFont("Arial", "B", 12)   // Bold font for table headers
	pdf.SetFillColor(200, 220, 255) // Set background color for table headers (light blue)

	// Create table headers for total orders, total amount, and orders with coupons
	pdf.CellFormat(60, 10, "Total Orders", "1", 0, "L", true, 0, "")
	pdf.CellFormat(60, 10, "Total Amount", "1", 0, "L", true, 0, "")
	pdf.CellFormat(70, 10, "Orders with Coupons", "1", 1, "L", true, 0, "")

	// Set font back to regular for table content
	pdf.SetFont("Arial", "", 12)

	// Fill the table with data from the first record
	pdf.CellFormat(60, 10, fmt.Sprintf("%d", data[0].OrderCount), "1", 0, "L", false, 0, "")
	pdf.CellFormat(60, 10, fmt.Sprintf("$%.2f", data[0].TotalAmount), "1", 0, "L", false, 0, "")
	pdf.CellFormat(70, 10, fmt.Sprintf("%d", data[0].CouponOrderCount), "1", 1, "L", false, 0, "")

	pdf.Ln(10) // Extra space after the table

	// --- Adding Key Metrics Table ---
	pdf.SetFont("Arial", "B", 14)                                    // Bold font for the "Key Metrics" title
	pdf.CellFormat(190, 10, "Key Metrics", "", 1, "L", false, 0, "") // Add the title for the key metrics section
	pdf.Ln(5)                                                        // Line break

	// Set font and fill color for the key metrics table headers
	pdf.SetFont("Arial", "B", 12)
	pdf.SetFillColor(240, 240, 240) // Set background color for table headers (light gray)

	// Add headers "Metric" and "Value"
	pdf.CellFormat(95, 10, "Metric", "1", 0, "C", true, 0, "")
	pdf.CellFormat(95, 10, "Value", "1", 1, "C", true, 0, "")

	// Set font to regular for the metrics content
	pdf.SetFont("Arial", "", 12)
	// Define key metrics with calculated values (average order value, coupon usage rate)
	metrics := []struct {
		name  string
		value string
	}{
		// Calculate average order value: total amount divided by number of orders
		{"Average Order Value", fmt.Sprintf("$%.2f", data[0].TotalAmount/float64(data[0].OrderCount))},
		// Calculate coupon usage rate: (orders with coupons / total orders) * 100
		{"Coupon Usage Rate", fmt.Sprintf("%.2f%%", float64(data[0].CouponOrderCount)/float64(data[0].OrderCount)*100)},
	}

	// Add metrics rows to the table
	for _, metric := range metrics {
		pdf.CellFormat(95, 10, metric.name, "1", 0, "L", false, 0, "")
		pdf.CellFormat(95, 10, metric.value, "1", 1, "L", false, 0, "")
	}

	// --- Adding Footer ---
	// Move the cursor to the bottom of the page (-15 mm from the bottom)
	pdf.SetY(-15)
	pdf.SetFont("Arial", "I", 8)
	// Add a footer with the current date and time
	pdf.CellFormat(0, 10, fmt.Sprintf("Report generated on %s", time.Now().Format("2006-01-02 15:04:05")), "", 0, "C", false, 0, "")

	// --- Output the PDF to a buffer ---
	var buf bytes.Buffer
	// Write the PDF content to the buffer
	err := pdf.Output(&buf)
	if err != nil {
		// Log an error if the PDF generation fails
		log.Printf("error while writing the pdf content to the buffer : %v", err)
		return nil, utils.ErrFailedToGeneratePDF
	}

	// Return the generated PDF content as a byte slice
	return buf.Bytes(), nil
}

// GenerateExcelReport generates an Excel report for daily sales data.
func GenerateExcelReport(data []domain.DailySales) ([]byte, error) {
	// Create a new Excel file using the excelize package
	f := excelize.NewFile()
	sheet := "Excel_sheet_daily_sales" // Define the sheet name

	// --- Adding Business Information ---
	// Set the business name, location, phone, and email in cells A1 to A4
	f.SetCellValue(sheet, "A1", "RM Sports Shop")
	f.SetCellValue(sheet, "A2", "Location: Calicut")
	f.SetCellValue(sheet, "A3", "Phone: +911234512345")
	f.SetCellValue(sheet, "A4", "Email: rmshop@gmail.com")

	// --- Adding Column Headers ---
	// Set headers for the data in row 6: Date, Order Count, Total Amount, and Coupon Order Count
	f.SetCellValue(sheet, "A6", "Date")
	f.SetCellValue(sheet, "B6", "Order Count")
	f.SetCellValue(sheet, "C6", "Total Amount")
	f.SetCellValue(sheet, "D6", "Coupon Order Count")

	// --- Adding Data Rows ---
	// Loop through the sales data and insert it into the sheet, starting from row 7
	for i, sale := range data {
		row := i + 7 // Calculate the row number (7 is the starting row for data)
		// Set the date in column A, formatted as "YYYY-MM-DD"
		f.SetCellValue(sheet, fmt.Sprintf("A%d", row), sale.Date.Format("2006-01-02"))
		// Set the order count, total amount, and coupon order count in columns B, C, and D
		f.SetCellValue(sheet, fmt.Sprintf("B%d", row), sale.OrderCount)
		f.SetCellValue(sheet, fmt.Sprintf("C%d", row), sale.TotalAmount)
		f.SetCellValue(sheet, fmt.Sprintf("D%d", row), sale.CouponOrderCount)
	}

	// --- Adding Key Metrics ---
	lastRow := len(data) + 8 // Calculate the row number for key metrics
	// Set the "Key Metrics" label in column A
	f.SetCellValue(sheet, fmt.Sprintf("A%d", lastRow), "Key Metrics")

	// Calculate and display "Average Order Value" in the next row (Total Amount / Order Count)
	f.SetCellValue(sheet, fmt.Sprintf("A%d", lastRow+1), "Average Order Value")
	f.SetCellValue(sheet, fmt.Sprintf("B%d", lastRow+1), data[0].TotalAmount/float64(data[0].OrderCount))

	// Calculate and display "Coupon Usage Rate" (Coupon Order Count / Order Count)
	f.SetCellValue(sheet, fmt.Sprintf("A%d", lastRow+2), "Coupon Usage Rate")
	f.SetCellValue(sheet, fmt.Sprintf("B%d", lastRow+2), float64(data[0].CouponOrderCount)/float64(data[0].OrderCount))

	// --- Writing the Excel File to Buffer ---
	var buf bytes.Buffer

	// Write the Excel file content to the buffer
	err := f.Write(&buf)
	if err != nil {
		// Log an error if writing the Excel content to the buffer fails
		log.Printf("error while writing the excel content to the buffer: %v", err)
		return nil, err
	}

	// Return the generated Excel file as a byte slice
	return buf.Bytes(), nil
}

func GenerateWeeklyPDFReport(data domain.WeeklySalesReport) ([]byte, error) {
	// Create a new PDF document with A4 page size, measured in millimeters (mm)
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage() // Add a new page to the PDF

	// --- Title and Business Info ---
	// Set font to Arial, bold, size 16 for the title, and center it on the page
	pdf.SetFont("Arial", "B", 16)
	pdf.CellFormat(190, 10, "Weekly Sales Report", "", 1, "C", false, 0, "")
	pdf.Ln(5) // Add a line break for spacing

	// Set font to Arial, normal, size 12 for business info, and left-align it
	pdf.SetFont("Arial", "", 12)
	pdf.CellFormat(190, 8, "RM Sports Shop", "", 1, "L", false, 0, "")
	pdf.CellFormat(190, 8, "Location: Calicut", "", 1, "L", false, 0, "")
	pdf.CellFormat(190, 8, "Phone: +911234512345", "", 1, "L", false, 0, "")
	pdf.CellFormat(190, 8, "Email: rmshop@gmail.com", "", 1, "L", false, 0, "")
	pdf.Ln(5) // Add spacing

	// --- Date Range ---
	// Display the report's date range using a bold font, size 12
	pdf.SetFont("Arial", "B", 12)
	pdf.CellFormat(190, 10, fmt.Sprintf("From: %s To: %s", data.StartDate.Format("2006-01-02"), data.EndDate.Format("2006-01-02")), "", 1, "L", false, 0, "")
	pdf.Ln(5) // Add spacing

	// --- Summary Table ---
	pdf.SetFont("Arial", "B", 14)
	pdf.CellFormat(190, 10, "Summary", "", 1, "L", false, 0, "")
	pdf.Ln(5)

	// Add table header for "Metric" and "Value" columns, with a light blue background
	pdf.SetFillColor(200, 220, 255) // Light blue
	pdf.SetFont("Arial", "B", 12)
	pdf.CellFormat(95, 10, "Metric", "1", 0, "C", true, 0, "")
	pdf.CellFormat(95, 10, "Value", "1", 1, "C", true, 0, "")

	// --- Table Content ---
	pdf.SetFont("Arial", "", 12)
	pdf.SetFillColor(255, 255, 255) // White background for table rows

	// Total Orders
	pdf.CellFormat(95, 10, "Total Orders", "1", 0, "L", true, 0, "")
	pdf.CellFormat(95, 10, fmt.Sprintf("%d", data.TotalOrderCount), "1", 1, "R", true, 0, "")

	// Total Amount
	pdf.CellFormat(95, 10, "Total Amount", "1", 0, "L", true, 0, "")
	pdf.CellFormat(95, 10, fmt.Sprintf("$%.2f", data.TotalAmount), "1", 1, "R", true, 0, "")

	// Orders with Coupons
	pdf.CellFormat(95, 10, "Orders with Coupons", "1", 0, "L", true, 0, "")
	pdf.CellFormat(95, 10, fmt.Sprintf("%d", data.TotalCouponOrderCount), "1", 1, "R", true, 0, "")

	pdf.Ln(10) // Add spacing

	// --- Daily Breakdown Table ---
	pdf.SetFont("Arial", "B", 14)
	pdf.CellFormat(190, 10, "Daily Breakdown", "", 1, "L", false, 0, "")
	pdf.Ln(5)

	// Add table header for daily breakdown, with a light blue background
	pdf.SetFillColor(200, 220, 255) // Light blue
	pdf.SetFont("Arial", "B", 12)
	pdf.CellFormat(47.5, 10, "Date", "1", 0, "C", true, 0, "")
	pdf.CellFormat(47.5, 10, "Orders", "1", 0, "C", true, 0, "")
	pdf.CellFormat(47.5, 10, "Amount", "1", 0, "C", true, 0, "")
	pdf.CellFormat(47.5, 10, "Coupon Orders", "1", 1, "C", true, 0, "")

	// --- Daily Sales Data ---
	pdf.SetFont("Arial", "", 12)
	for i, day := range data.DailySales {
		// Alternate background colors: white for even rows and light gray for odd rows
		fillColor := 255 // White background
		if i%2 == 1 {
			fillColor = 240 // Light gray for alternating rows
		}
		pdf.SetFillColor(fillColor, fillColor, fillColor) // Apply background color

		// Add daily data for date, order count, total amount, and coupon order count
		pdf.CellFormat(47.5, 10, day.Date.Format("2006-01-02"), "1", 0, "L", true, 0, "")
		pdf.CellFormat(47.5, 10, fmt.Sprintf("%d", day.OrderCount), "1", 0, "R", true, 0, "")
		pdf.CellFormat(47.5, 10, fmt.Sprintf("$%.2f", day.TotalAmount), "1", 0, "R", true, 0, "")
		pdf.CellFormat(47.5, 10, fmt.Sprintf("%d", day.CouponOrderCount), "1", 1, "R", true, 0, "")
	}

	// --- Output the PDF ---
	var buf bytes.Buffer
	err := pdf.Output(&buf) // Generate PDF content and write it to the buffer
	if err != nil {
		log.Printf("failed to generate pdf : %v", err)
		return nil, err // Return an error if PDF generation fails
	}

	// Return the PDF as a byte slice
	return buf.Bytes(), nil
}

func GenerateWeeklyExcelReport(data domain.WeeklySalesReport) ([]byte, error) {
	// Create a new Excel file
	f := excelize.NewFile()

	// Define the sheet name for the report
	sheet := "Weekly_Sales_Report"
	// Rename the default sheet to the report sheet name
	f.SetSheetName("Weekly_Sales_Report1", sheet)

	// Add the title and business information
	f.SetCellValue(sheet, "A1", "Weekly Sales Report")
	f.MergeCell(sheet, "A1", "D1") // Merge cells for the title
	f.SetCellValue(sheet, "A2", "RM Sports Shop")
	f.SetCellValue(sheet, "A3", "Location: Calicut")
	f.SetCellValue(sheet, "A4", "Phone: +911234512345")
	f.SetCellValue(sheet, "A5", "Email: rmshop@gmail.com")

	// Add the date range of the report
	f.SetCellValue(sheet, "A7", fmt.Sprintf("From: %s To: %s", data.StartDate.Format("2006-01-02"), data.EndDate.Format("2006-01-02")))
	f.MergeCell(sheet, "A7", "D7") // Merge cells for the date range

	// Add the summary section header
	f.SetCellValue(sheet, "A9", "Summary")
	f.MergeCell(sheet, "A9", "D9") // Merge cells for the summary header

	// Add summary metrics (Total Orders, Total Amount, Orders with Coupons)
	f.SetCellValue(sheet, "A10", "Metric") // Column for metric names
	f.SetCellValue(sheet, "B10", "Value")  // Column for metric values
	f.SetCellValue(sheet, "A11", "Total Orders")
	f.SetCellValue(sheet, "B11", data.TotalOrderCount)
	f.SetCellValue(sheet, "A12", "Total Amount")
	f.SetCellValue(sheet, "B12", data.TotalAmount)
	f.SetCellValue(sheet, "A13", "Orders with Coupons")
	f.SetCellValue(sheet, "B13", data.TotalCouponOrderCount)

	// Add the daily sales breakdown section header
	f.SetCellValue(sheet, "A15", "Daily Breakdown")
	f.MergeCell(sheet, "A15", "D15") // Merge cells for the daily breakdown header

	// Add the headers for the daily breakdown table (Date, Orders, Amount, Coupon Orders)
	f.SetCellValue(sheet, "A16", "Date")
	f.SetCellValue(sheet, "B16", "Orders")
	f.SetCellValue(sheet, "C16", "Amount")
	f.SetCellValue(sheet, "D16", "Coupon Orders")

	// Loop through the daily sales data and fill in the breakdown table
	for i, day := range data.DailySales {
		row := 17 + i
		f.SetCellValue(sheet, fmt.Sprintf("A%d", row), day.Date.Format("2006-01-02"))
		f.SetCellValue(sheet, fmt.Sprintf("B%d", row), day.OrderCount)
		f.SetCellValue(sheet, fmt.Sprintf("C%d", row), day.TotalAmount)
		f.SetCellValue(sheet, fmt.Sprintf("D%d", row), day.CouponOrderCount)
	}

	// Apply styles to the Excel report
	// Style for the title (bold, size 16, centered)
	titleStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 16},
		Alignment: &excelize.Alignment{Horizontal: "center"},
	})
	f.SetCellStyle(sheet, "A1", "D1", titleStyle)

	// Style for the headers (bold, light blue background, centered, with borders)
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

	// Style for the content (with borders)
	contentStyle, _ := f.NewStyle(&excelize.Style{
		Border: []excelize.Border{{Type: "top", Color: "000000", Style: 1},
			{Type: "left", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1}},
	})
	f.SetCellStyle(sheet, "A11", "B13", contentStyle)
	f.SetCellStyle(sheet, "A17", fmt.Sprintf("D%d", 16+len(data.DailySales)), contentStyle)

	// Set the column widths for the table
	f.SetColWidth(sheet, "A", "D", 20)

	// Write the Excel content to a buffer
	var buf bytes.Buffer
	err := f.Write(&buf)
	if err != nil {
		log.Printf("error while writing excel content to a buffer %v", err)
		return nil, err // Return an error if writing fails
	}

	// Return the generated Excel report as a byte slice
	return buf.Bytes(), nil
}

func GenerateMonthlyPDFReport(data domain.MonthlySalesReport) ([]byte, error) {
	// Create a new PDF document with portrait orientation ("P"), millimeter units ("mm"), and A4 size.
	pdf := gofpdf.New("P", "mm", "A4", "")
	// Add a new page to the PDF.
	pdf.AddPage()

	// Add business information at the top of the page.
	// Set the font to Arial, bold, size 16 for the business name.
	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(0, 10, "RM Sports Shop") // Add the business name.
	pdf.Ln(8)                         // Move to the next line with a line break.

	// Add additional business info: location, phone number, and email.
	pdf.SetFont("Arial", "", 12)        // Set the font to Arial, regular, size 12.
	pdf.Cell(0, 6, "Location: Calicut") // Add the business location.
	pdf.Ln(6)
	pdf.Cell(0, 6, "Phone: +911234512345") // Add the business phone number.
	pdf.Ln(6)
	pdf.Cell(0, 6, "Email: rmshop@gmail.com") // Add the business email.
	pdf.Ln(15)                                // Add some extra space after the contact details.

	// Add the report title (Monthly Sales Report) with the selected month and year.
	pdf.SetFont("Arial", "B", 16) // Set the font to Arial, bold, size 16 for the title.
	pdf.Cell(0, 10, fmt.Sprintf("Monthly Sales Report - %s %d", data.Month.String(), data.Year))
	pdf.Ln(15)

	// Add the summary section of the report.
	pdf.SetFont("Arial", "B", 14) // Set the font to Arial, bold, size 14 for the section header.
	pdf.Cell(0, 10, "Summary")    // Add the "Summary" header.
	pdf.Ln(10)

	// Define a table for the summary section with the metric names and their corresponding values.
	summaryTable := [][]string{
		{"Metric", "Value"},
		{"Total Orders", fmt.Sprintf("%d", data.TotalOrderCount)},              // Total number of orders.
		{"Total Amount", fmt.Sprintf("$%.2f", data.TotalAmount)},               // Total sales amount.
		{"Orders with Coupons", fmt.Sprintf("%d", data.TotalCouponOrderCount)}, // Orders made with coupons.
	}

	// Set colors for the table header.
	pdf.SetFillColor(200, 220, 255) // Set fill color to a light blue shade.
	pdf.SetTextColor(0, 0, 0)       // Set text color to black.
	pdf.SetFont("Arial", "B", 12)   // Set the font to Arial, bold, size 12 for the table headers.

	// Loop through each row in the summary table to create the table structure.
	for i, row := range summaryTable {
		for _, col := range row {
			// Set fill color for the first row (header) or white for other rows.
			if i == 0 {
				pdf.SetFillColor(200, 220, 255)
			} else {
				pdf.SetFillColor(255, 255, 255)
			}
			pdf.CellFormat(60, 8, col, "1", 0, "LM", true, 0, "") // Create a table cell for each column.
		}
		pdf.Ln(-1) // Move to the next row in the table.
	}
	pdf.Ln(10) // Add some extra space after the summary table.

	// Add the daily breakdown section of the report.
	pdf.SetFont("Arial", "B", 14)      // Set the font to Arial, bold, size 14 for the section header.
	pdf.Cell(0, 10, "Daily Breakdown") // Add the "Daily Breakdown" header.
	pdf.Ln(10)

	// Define the table structure for the daily breakdown.
	breakdownHeader := []string{"Date", "Orders", "Amount", "Coupon Orders"} // Define the headers for each column.

	// Set table header styling.
	pdf.SetFillColor(200, 220, 255) // Set the fill color for the header.
	pdf.SetTextColor(0, 0, 0)       // Set text color to black.
	pdf.SetFont("Arial", "B", 12)   // Set the font to Arial, bold, size 12 for the table headers.

	// Create the header row for the daily breakdown table.
	for _, col := range breakdownHeader {
		pdf.CellFormat(45, 8, col, "1", 0, "CM", true, 0, "") // Create a table cell for each header.
	}
	pdf.Ln(-1) // Move to the next row.

	// Populate the daily breakdown table with data.
	pdf.SetFont("Arial", "", 12)    // Set the font to Arial, regular, size 12 for the table content.
	pdf.SetFillColor(255, 255, 255) // Set the fill color for the table content rows.
	for _, day := range data.DailySales {
		// Add data for each day: date, order count, total amount, and coupon order count.
		pdf.CellFormat(45, 8, day.Date.Format("2006-01-02"), "1", 0, "LM", true, 0, "")           // Date.
		pdf.CellFormat(45, 8, fmt.Sprintf("%d", day.OrderCount), "1", 0, "RM", true, 0, "")       // Orders
		pdf.CellFormat(45, 8, fmt.Sprintf("$%.2f", day.TotalAmount), "1", 0, "RM", true, 0, "")   // Amount
		pdf.CellFormat(45, 8, fmt.Sprintf("%d", day.CouponOrderCount), "1", 0, "RM", true, 0, "") // Coupon Orders
		pdf.Ln(-1)                                                                                // Move to the next row.
	}

	// Add a footer with the report generation date and time.
	pdf.SetY(-15)                // Set the position to 15 mm from the bottom of the page.
	pdf.SetFont("Arial", "I", 8) // Set the font to Arial, italic, size 8 for the footer
	pdf.Cell(0, 10, fmt.Sprintf("Report generated on %s", time.Now().Format("2006-01-02 15:04:05")))

	// Write the PDF content to a buffer
	var buf bytes.Buffer
	err := pdf.Output(&buf) // Generate the PDF and store it in the buffer
	if err != nil {
		log.Printf("failed to generate pdf : %v", err)
		return nil, err
	}

	// Return the generated PDF as a byte slice.
	return buf.Bytes(), nil
}

func GenerateMonthlyExcelReport(data domain.MonthlySalesReport) ([]byte, error) {
	// Create a new Excel file
	f := excelize.NewFile()

	// Define the sheet name for the report
	sheet := "Monthly Sales Report"
	// Rename the default sheet to "Monthly Sales Report"
	f.SetSheetName("Sheet1", sheet)

	// Add business information to the top of the sheet
	f.SetCellValue(sheet, "A1", "RM Sports Shop")
	f.SetCellValue(sheet, "A2", "Location: Calicut")
	f.SetCellValue(sheet, "A3", "Phone: +911234512345")
	f.SetCellValue(sheet, "A4", "Email: rmshop@gmail.com")

	// Add the report title with the month and year from the data
	f.SetCellValue(sheet, "A6", fmt.Sprintf("Monthly Sales Report - %s %d", data.Month.String(), data.Year))

	// Add a section for the summary of sales
	f.SetCellValue(sheet, "A8", "Summary")

	// Define the summary data to be added to the sheet
	summaryData := [][]interface{}{
		{"Metric", "Value"},
		{"Total Orders", data.TotalOrderCount},
		{"Total Amount", data.TotalAmount},
		{"Orders with Coupons", data.TotalCouponOrderCount},
	}

	// Add the summary data to the sheet starting at row 9
	for i, row := range summaryData {
		for j, cellValue := range row {
			// Convert coordinates to cell name and set the cell value
			cell, _ := excelize.CoordinatesToCellName(j+1, i+9)
			f.SetCellValue(sheet, cell, cellValue)
		}
	}

	// Apply a border style to the summary table
	summaryStyle, _ := f.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
		},
	})
	// Apply the style to the summary table's data range
	f.SetCellStyle(sheet, "A9", "B12", summaryStyle)

	// Create a bold and colored header style for the summary table
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
	// Apply the header style to the summary table header
	f.SetCellStyle(sheet, "A9", "B9", summaryHeaderStyle)

	// Add a section for daily breakdown data
	f.SetCellValue(sheet, "A14", "Daily Breakdown")
	// Define the headers for the daily breakdown table
	breakdownHeaders := []string{"Date", "Orders", "Amount", "Coupon Orders"}
	// Add the headers to the sheet
	for i, header := range breakdownHeaders {
		cell, _ := excelize.CoordinatesToCellName(i+1, 15)
		f.SetCellValue(sheet, cell, header)
	}

	// Add the daily sales data to the sheet
	for i, day := range data.DailySales {
		row := 16 + i
		f.SetCellValue(sheet, fmt.Sprintf("A%d", row), day.Date.Format("2006-01-02"))
		f.SetCellValue(sheet, fmt.Sprintf("B%d", row), day.OrderCount)
		f.SetCellValue(sheet, fmt.Sprintf("C%d", row), day.TotalAmount)
		f.SetCellValue(sheet, fmt.Sprintf("D%d", row), day.CouponOrderCount)
	}

	// Format the daily breakdown table with borders
	lastRow := 15 + len(data.DailySales)
	breakdownStyle, _ := f.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
		},
	})
	// Apply the style to the daily breakdown table
	f.SetCellStyle(sheet, "A15", fmt.Sprintf("D%d", lastRow), breakdownStyle)

	// Apply a bold and colored header style for the daily breakdown table
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
	// Apply the header style to the daily breakdown header
	f.SetCellStyle(sheet, "A15", "D15", breakdownHeaderStyle)

	// Adjust the column width to fit the content
	for col := 'A'; col <= 'D'; col++ {
		f.SetColWidth(sheet, string(col), string(col), 20)
	}

	// Add a footer with the report generation date
	footerRow := lastRow + 2
	f.SetCellValue(sheet, fmt.Sprintf("A%d", footerRow), fmt.Sprintf("Report generated on %s", time.Now().Format("2006-01-02 15:04:05")))

	// Write the Excel file content to a buffer
	var buf bytes.Buffer
	err := f.Write(&buf)
	if err != nil {
		log.Printf("error while adding excel content to the buffer : %v", err)
		return nil, err
	}

	// Return the generated Excel report as bytes
	return buf.Bytes(), nil
}

func GenerateCustomPDFReport(data domain.CustomSalesReport) ([]byte, error) {
	// Create a new PDF document in portrait mode ("P"), using millimeters ("mm") for measurements, and A4 page size.
	pdf := gofpdf.New("P", "mm", "A4", "")
	// Add a new page to the PDF
	pdf.AddPage()

	// Set the font to Arial, bold, size 16 for the title
	pdf.SetFont("Arial", "B", 16)

	// Add the title "Custom Sales Report" centered on the page.
	pdf.CellFormat(190, 10, "Custom Sales Report", "", 1, "C", false, 0, "")
	pdf.Ln(5) // Add a line break (5mm height).

	// Set the font to Arial, regular, size 12 for the business information section.
	pdf.SetFont("Arial", "", 12)

	// Add business name, address, phone, and email, each centered on the page
	pdf.CellFormat(190, 7, "RM Sports Shop", "", 1, "C", false, 0, "")
	pdf.CellFormat(190, 7, "Calicut", "", 1, "C", false, 0, "")
	pdf.CellFormat(190, 7, "Phone: +911234512345", "", 1, "C", false, 0, "")
	pdf.CellFormat(190, 7, "Email: rmshop@gmail.com", "", 1, "C", false, 0, "")
	pdf.Ln(10) // Add a line break (10mm height).

	// Add the date range for the sales report, formatted as "From: YYYY-MM-DD To: YYYY-MM-DD".
	pdf.SetFont("Arial", "", 12)
	pdf.CellFormat(190, 10, fmt.Sprintf("From: %s To: %s", data.StartDate.Format("2006-01-02"), data.EndDate.Format("2006-01-02")), "", 1, "C", false, 0, "")
	pdf.Ln(10) // Add a line break (10mm height).

	// Add the "Summary" section heading in bold.
	pdf.SetFont("Arial", "B", 14)
	pdf.CellFormat(190, 10, "Summary", "", 1, "L", false, 0, "")
	pdf.Ln(5) // Add a line break (5mm height)

	// Define the summary table data as a 2D array with labels and corresponding values
	summaryTable := [][]string{
		{"Total Orders", fmt.Sprintf("%d", data.TotalOrderCount)},
		{"Total Amount", fmt.Sprintf("$%.2f", data.TotalAmount)},
		{"Orders with Coupons", fmt.Sprintf("%d", data.TotalCouponOrderCount)},
	}

	// Set the background color for the summary table to light gray
	pdf.SetFillColor(240, 240, 240) // RGB color (240, 240, 240)
	pdf.SetTextColor(0, 0, 0)       // Set text color to black (RGB: 0, 0, 0).

	// Loop through the summaryTable array and create table rows with alternating row colors.
	for i, row := range summaryTable {
		pdf.SetFont("Arial", "B", 12)                            // Set the font for the label (first column) to bold.
		pdf.CellFormat(95, 10, row[0], "1", 0, "L", true, 0, "") // Create the label cell.
		pdf.SetFont("Arial", "", 12)                             // Set the font for the value (second column) to regular.
		pdf.CellFormat(95, 10, row[1], "1", 1, "R", true, 0, "") // Create the value cell.

		// Alternate row colors: white for even rows and light gray for odd rows.
		if i%2 == 1 {
			pdf.SetFillColor(255, 255, 255) // Set to white background.
		} else {
			pdf.SetFillColor(240, 240, 240) // Set to light gray background
		}
	}
	pdf.Ln(10) // Add a line break (10mm height)

	// Add the "Daily Breakdown" section heading in bold.
	pdf.SetFont("Arial", "B", 14)
	pdf.CellFormat(190, 10, "Daily Breakdown", "", 1, "L", false, 0, "")
	pdf.Ln(5) // Add a line break (5mm height)

	// Define table headers for the daily breakdown section and set a darker gray fill for the header background
	pdf.SetFont("Arial", "B", 12)
	pdf.SetFillColor(200, 200, 200)                                     // Set to darker gray background for headers
	pdf.CellFormat(47.5, 10, "Date", "1", 0, "C", true, 0, "")          // Date header.
	pdf.CellFormat(47.5, 10, "Orders", "1", 0, "C", true, 0, "")        // Orders header
	pdf.CellFormat(47.5, 10, "Amount", "1", 0, "C", true, 0, "")        // Amount header
	pdf.CellFormat(47.5, 10, "Coupon Orders", "1", 1, "C", true, 0, "") // Coupon Orders header

	// Loop through the DailySales data to populate the daily breakdown table
	pdf.SetFont("Arial", "", 12)
	for i, day := range data.DailySales {
		// Alternate row colors: white for even rows and light gray for odd rows
		if i%2 == 0 {
			pdf.SetFillColor(255, 255, 255) // White background for even rows
		} else {
			pdf.SetFillColor(240, 240, 240) // Light gray background for odd rows
		}

		// Add the row data: Date, Orders, Amount, and Coupon Orders
		pdf.CellFormat(47.5, 10, day.Date.Format("2006-01-02"), "1", 0, "C", true, 0, "")
		pdf.CellFormat(47.5, 10, fmt.Sprintf("%d", day.OrderCount), "1", 0, "C", true, 0, "")
		pdf.CellFormat(47.5, 10, fmt.Sprintf("$%.2f", day.TotalAmount), "1", 0, "C", true, 0, "")
		pdf.CellFormat(47.5, 10, fmt.Sprintf("%d", day.CouponOrderCount), "1", 1, "C", true, 0, "")
	}

	// Add a footer to the page with the report generation date and time
	pdf.SetY(-15)                // Move to 15mm from the bottom of the page
	pdf.SetFont("Arial", "I", 8) // Set the font to Arial, italic, size 8.
	pdf.CellFormat(0, 10, fmt.Sprintf("Report generated on %s", time.Now().Format("2006-01-02 15:04:05")), "", 0, "C", false, 0, "")

	// Create a buffer to hold the generated PDF content
	var buf bytes.Buffer
	// Output the PDF into the buffer and check for errors
	err := pdf.Output(&buf)
	if err != nil {
		log.Printf("error while generating pdf : %v", err)
		return nil, err // Return an error if PDF generation fails
	}

	// Return the PDF content as a byte slice
	return buf.Bytes(), nil
}

func GenerateCustomExcelReport(data domain.CustomSalesReport) ([]byte, error) {
	// Create a new Excel file
	f := excelize.NewFile()
	sheet := "Custom Sales Report"  // Define the name for the new sheet
	f.SetSheetName("Sheet1", sheet) // Rename the default sheet to "Custom Sales Report"

	// Add title and business information
	f.SetCellValue(sheet, "A1", "Custom Sales Report")     // Set the title in cell A1
	f.SetCellValue(sheet, "A2", "RM Sports Shop")          // Set the shop name in cell A2
	f.SetCellValue(sheet, "A3", "Calicut")                 // Set the location in cell A3
	f.SetCellValue(sheet, "A4", "Phone: +911234512345")    // Set the phone number in cell A4
	f.SetCellValue(sheet, "A5", "Email: rmshop@gmail.com") // Set the email in cell A5
	f.SetCellValue(sheet, "A7", fmt.Sprintf("From: %s To: %s", data.StartDate.Format("2006-01-02"), data.EndDate.Format("2006-01-02")))
	// Add the date range in cell A7

	// Create styles for different sections of the report
	titleStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 16},      // Bold title font of size 16
		Alignment: &excelize.Alignment{Horizontal: "center"}, // Center alignment
	})

	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Color: "#FFFFFF"},                           // Bold white font for headers
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"#4472C4"}, Pattern: 1}, // Blue background for headers
		Alignment: &excelize.Alignment{Horizontal: "center"},                              // Center alignment
	})

	evenRowStyle, _ := f.NewStyle(&excelize.Style{
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"#E9EFF7"}, Pattern: 1}, // Light color for even rows
		Alignment: &excelize.Alignment{Horizontal: "center"},                              // Center alignment
	})

	oddRowStyle, _ := f.NewStyle(&excelize.Style{
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"#D5E2F3"}, Pattern: 1}, // Slightly darker for odd rows
		Alignment: &excelize.Alignment{Horizontal: "center"},                              // Center alignment
	})

	borderStyle, _ := f.NewStyle(&excelize.Style{
		Border: []excelize.Border{ // Define borders for the cells
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
		},
	})

	// Apply the title style to the title cell
	f.SetCellStyle(sheet, "A1", "A1", titleStyle)

	// Add summary section title
	f.SetCellValue(sheet, "A9", "Summary")         // Set the summary title in cell A9
	f.SetCellStyle(sheet, "A9", "B9", headerStyle) // Apply header style to the summary title

	// Prepare summary data
	summaryData := [][]interface{}{
		{"Total Orders", data.TotalOrderCount},              // Total order count
		{"Total Amount", data.TotalAmount},                  // Total sales amount
		{"Orders with Coupons", data.TotalCouponOrderCount}, // Orders using coupons
	}

	// Fill the summary data into the sheet
	for i, row := range summaryData {
		rowNum := 10 + i                                          // Calculate the row number for summary data
		f.SetCellValue(sheet, fmt.Sprintf("A%d", rowNum), row[0]) // Set the first column value
		f.SetCellValue(sheet, fmt.Sprintf("B%d", rowNum), row[1]) // Set the second column value

		// Apply even or odd row style based on the row index
		if i%2 == 0 {
			f.SetCellStyle(sheet, fmt.Sprintf("A%d", rowNum), fmt.Sprintf("B%d", rowNum), evenRowStyle) // Even row style
		} else {
			f.SetCellStyle(sheet, fmt.Sprintf("A%d", rowNum), fmt.Sprintf("B%d", rowNum), oddRowStyle) // Odd row style
		}
	}

	// Add daily breakdown section title
	f.SetCellValue(sheet, "A14", "Daily Breakdown")  // Set title for daily breakdown
	f.SetCellStyle(sheet, "A14", "D14", headerStyle) // Apply header style

	// Set column headers for daily breakdown
	f.SetCellValue(sheet, "A15", "Date")             // Date column header
	f.SetCellValue(sheet, "B15", "Orders")           // Orders column header
	f.SetCellValue(sheet, "C15", "Amount")           // Amount column header
	f.SetCellValue(sheet, "D15", "Coupon Orders")    // Coupon orders column header
	f.SetCellStyle(sheet, "A15", "D15", headerStyle) // Apply header style to all column headers

	// Fill daily sales data into the sheet
	for i, day := range data.DailySales {
		row := 16 + i                                                                 // Calculate the row number for daily sales data
		f.SetCellValue(sheet, fmt.Sprintf("A%d", row), day.Date.Format("2006-01-02")) // Set the date
		f.SetCellValue(sheet, fmt.Sprintf("B%d", row), day.OrderCount)                // Set the order count
		f.SetCellValue(sheet, fmt.Sprintf("C%d", row), day.TotalAmount)               // Set the total amount
		f.SetCellValue(sheet, fmt.Sprintf("D%d", row), day.CouponOrderCount)          // Set coupon order count

		// Apply even or odd row style based on the row index
		if i%2 == 0 {
			f.SetCellStyle(sheet, fmt.Sprintf("A%d", row), fmt.Sprintf("D%d", row), evenRowStyle)
		} else {
			f.SetCellStyle(sheet, fmt.Sprintf("A%d", row), fmt.Sprintf("D%d", row), oddRowStyle)
		}
	}

	// Auto-fit column width for better readability
	for col := 'A'; col <= 'D'; col++ {
		f.SetColWidth(sheet, string(col), string(col), 20)
	}

	// Add footer with the report generation date
	lastRow := 16 + len(data.DailySales) + 2
	f.SetCellValue(sheet, fmt.Sprintf("A%d", lastRow), fmt.Sprintf("Report generated on %s", time.Now().Format("2006-01-02 15:04:05")))

	// Set overall border styles for the report
	f.SetCellStyle(sheet, "A1", fmt.Sprintf("D%d", lastRow), borderStyle)

	var buf bytes.Buffer // Create a buffer to hold the Excel file data
	err := f.Write(&buf) // Write the file data to the buffer
	if err != nil {
		log.Printf("error while adding excel content to buffer : %v", err)
		return nil, err // Return an error if writing fails
	}

	return buf.Bytes(), nil // Return the generated Excel file data as a byte slice
}
