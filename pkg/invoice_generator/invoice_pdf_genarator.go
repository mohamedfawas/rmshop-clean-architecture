package invoicegenerator

import (
	"bytes"
	"fmt"
	"time"

	"github.com/jung-kurt/gofpdf"
	"github.com/mohamedfawas/rmshop-clean-architecture/internal/domain"
)

func GenerateInvoicePDF(order *domain.Order) ([]byte, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

	// Add business information
	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(190, 10, "RM Sports Shop")
	pdf.Ln(8)
	pdf.SetFont("Arial", "", 12)
	pdf.Cell(190, 6, "Location: Calicut")
	pdf.Ln(6)
	pdf.Cell(190, 6, "Phone: +911234512345")
	pdf.Ln(6)
	pdf.Cell(190, 6, "Email: rmshop@gmail.com")
	pdf.Ln(15)

	// Add invoice header
	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(0, 10, "Invoice")
	pdf.Ln(10)

	// Add order details
	pdf.SetFont("Arial", "", 12)
	pdf.Cell(40, 8, fmt.Sprintf("Order ID: %d", order.ID))
	pdf.Ln(8)
	pdf.Cell(40, 8, fmt.Sprintf("Date: %s", order.CreatedAt.Format("2006-01-02 15:04:05")))
	pdf.Ln(8)
	pdf.Cell(40, 8, fmt.Sprintf("Status: %s", order.OrderStatus))
	pdf.Ln(15)

	// Add shipping address
	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(0, 8, "Shipping Address:")
	pdf.Ln(8)
	pdf.SetFont("Arial", "", 10)

	// Create address string
	addressStr := fmt.Sprintf("%s\n", order.ShippingAddress.AddressLine1)
	if order.ShippingAddress.AddressLine2 != "" {
		addressStr += fmt.Sprintf("%s\n", order.ShippingAddress.AddressLine2)
	}
	addressStr += fmt.Sprintf("%s, %s, %s\n",
		order.ShippingAddress.City,
		order.ShippingAddress.State,
		order.ShippingAddress.PinCode)
	addressStr += fmt.Sprintf("Phone: %s", order.ShippingAddress.PhoneNumber)

	// Use MultiCell with correct arguments
	pdf.MultiCell(0, 6, addressStr, "", "", false)
	pdf.Ln(10)

	// Add items table
	pdf.SetFont("Arial", "B", 12)
	pdf.SetFillColor(200, 220, 255)
	pdf.CellFormat(90, 8, "Product", "1", 0, "C", true, 0, "")
	pdf.CellFormat(30, 8, "Quantity", "1", 0, "C", true, 0, "")
	pdf.CellFormat(35, 8, "Price", "1", 0, "C", true, 0, "")
	pdf.CellFormat(35, 8, "Subtotal", "1", 1, "C", true, 0, "")

	pdf.SetFont("Arial", "", 10)
	for _, item := range order.Items {
		pdf.CellFormat(90, 8, fmt.Sprintf("%s (ID: %d)", item.ProductName, item.ProductID), "1", 0, "L", false, 0, "")
		pdf.CellFormat(30, 8, fmt.Sprintf("%d", item.Quantity), "1", 0, "C", false, 0, "")
		pdf.CellFormat(35, 8, fmt.Sprintf("$%.2f", item.Price), "1", 0, "R", false, 0, "")
		pdf.CellFormat(35, 8, fmt.Sprintf("$%.2f", float64(item.Quantity)*item.Price), "1", 1, "R", false, 0, "")
	}

	// Add total, discount, and final amount
	pdf.Ln(5)
	pdf.SetFont("Arial", "B", 12)
	pdf.CellFormat(155, 8, "Total Amount", "1", 0, "R", false, 0, "")
	pdf.CellFormat(35, 8, fmt.Sprintf("$%.2f", order.TotalAmount), "1", 1, "R", false, 0, "")

	if order.DiscountAmount > 0 {
		pdf.CellFormat(155, 8, "Discount", "1", 0, "R", false, 0, "")
		pdf.CellFormat(35, 8, fmt.Sprintf("-$%.2f", order.DiscountAmount), "1", 1, "R", false, 0, "")
	}

	if order.CouponApplied {
		pdf.CellFormat(155, 8, "Coupon Applied", "1", 0, "R", false, 0, "")
		pdf.CellFormat(35, 8, "Yes", "1", 1, "R", false, 0, "")
	}

	pdf.SetFont("Arial", "B", 14)
	pdf.CellFormat(155, 10, "Final Amount", "1", 0, "R", false, 0, "")
	pdf.CellFormat(35, 10, fmt.Sprintf("$%.2f", order.FinalAmount), "1", 1, "R", false, 0, "")

	// Add footer
	pdf.SetY(-15)
	pdf.SetFont("Arial", "I", 8)
	pdf.Cell(0, 10, fmt.Sprintf("Invoice generated on %s", time.Now().Format("2006-01-02 15:04:05")))

	var buf bytes.Buffer
	err := pdf.Output(&buf)
	if err != nil {
		return nil, fmt.Errorf("failed to generate PDF: %w", err)
	}

	return buf.Bytes(), nil
}
