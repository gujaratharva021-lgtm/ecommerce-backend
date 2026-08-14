package handlers

import (
"fmt"
"net/http"

"github.com/gin-gonic/gin"
"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/config"
"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/database"
"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/models"
"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/services"
"github.com/jung-kurt/gofpdf"
)

// buildInvoicePDF renders the invoice exactly as stored - it must never
// compute or display a figure that isn't already on the Invoice/InvoiceItem
// rows, since those are the legal record. No tax line is printed: this
// project has no tax/GST calculation anywhere in the order flow, so
// inventing a rate here would make the PDF wrong, not more complete.
//
// Seller details come from config.AppConfig (env-configured, see
// internal/config/config.go) - not hardcoded, so they match whatever the
// JSON invoice API returns via sellerDetails() in warehouse_invoice_handler.go.
func buildInvoicePDF(invoice models.Invoice) ([]byte, error) {
cfg := config.AppConfig

pdf := gofpdf.New("P", "mm", "A4", "")
pdf.AddPage()
pdf.SetFont("Arial", "B", 16)
pdf.Cell(0, 10, "TAX INVOICE")
pdf.Ln(12)

pdf.SetFont("Arial", "B", 11)
pdf.Cell(0, 6, cfg.SellerCompanyName)
pdf.Ln(6)
pdf.SetFont("Arial", "", 9)
pdf.Cell(0, 5, cfg.SellerAddress)
pdf.Ln(5)
if cfg.SellerGSTIN != "" {
pdf.Cell(0, 5, "GSTIN: "+cfg.SellerGSTIN)
pdf.Ln(5)
}
if cfg.SellerFSSAINumber != "" {
pdf.Cell(0, 5, "FSSAI: "+cfg.SellerFSSAINumber)
pdf.Ln(5)
}
if cfg.SellerContactNumber != "" || cfg.SellerEmail != "" {
contactLine := cfg.SellerContactNumber
if cfg.SellerEmail != "" {
if contactLine != "" {
contactLine += " | "
}
contactLine += cfg.SellerEmail
}
pdf.Cell(0, 5, contactLine)
pdf.Ln(5)
}
pdf.Ln(4)

pdf.SetFont("Arial", "B", 10)
pdf.CellFormat(95, 6, "Invoice Number: "+invoice.InvoiceNumber, "", 0, "L", false, 0, "")
pdf.CellFormat(95, 6, "Invoice Date: "+invoice.GeneratedAt.Format("02 Jan 2006"), "", 1, "L", false, 0, "")
pdf.CellFormat(95, 6, fmt.Sprintf("Order ID: #%d", invoice.OrderID), "", 0, "L", false, 0, "")
pdf.CellFormat(95, 6, "Payment Method: "+invoice.PaymentMethod, "", 1, "L", false, 0, "")
if invoice.PaymentReference != "" {
pdf.CellFormat(95, 6, "Payment Reference: "+invoice.PaymentReference, "", 1, "L", false, 0, "")
}
if invoice.AddressState != "" {
pdf.CellFormat(95, 6, "Place of Supply: "+invoice.AddressState, "", 1, "L", false, 0, "")
}
pdf.Ln(6)

pdf.SetFont("Arial", "B", 10)
pdf.Cell(0, 6, "Bill To")
pdf.Ln(6)
pdf.SetFont("Arial", "", 9)
pdf.Cell(0, 5, invoice.CustomerName)
pdf.Ln(5)
pdf.Cell(0, 5, invoice.CustomerPhone)
pdf.Ln(5)
addressLine := invoice.AddressLine1
if invoice.AddressLine2 != "" {
addressLine += ", " + invoice.AddressLine2
}
pdf.MultiCell(0, 5, addressLine, "", "L", false)
pdf.Cell(0, 5, fmt.Sprintf("%s, %s - %s", invoice.AddressCity, invoice.AddressState, invoice.AddressPincode))
pdf.Ln(10)

// Items table
pdf.SetFont("Arial", "B", 9)
pdf.SetFillColor(230, 230, 230)
pdf.CellFormat(75, 7, "Item", "1", 0, "L", true, 0, "")
pdf.CellFormat(30, 7, "SKU", "1", 0, "L", true, 0, "")
pdf.CellFormat(20, 7, "Qty", "1", 0, "C", true, 0, "")
pdf.CellFormat(30, 7, "Unit Price", "1", 0, "R", true, 0, "")
pdf.CellFormat(35, 7, "Line Total", "1", 1, "R", true, 0, "")

pdf.SetFont("Arial", "", 9)
for _, item := range invoice.Items {
lineTotal := item.Price * float64(item.Quantity)
pdf.CellFormat(75, 7, item.ProductName, "1", 0, "L", false, 0, "")
pdf.CellFormat(30, 7, item.SKU, "1", 0, "L", false, 0, "")
pdf.CellFormat(20, 7, fmt.Sprintf("%d", item.Quantity), "1", 0, "C", false, 0, "")
pdf.CellFormat(30, 7, fmt.Sprintf("%.2f", item.Price), "1", 0, "R", false, 0, "")
pdf.CellFormat(35, 7, fmt.Sprintf("%.2f", lineTotal), "1", 1, "R", false, 0, "")
}
pdf.Ln(6)

// Totals - mirrors exactly what's stored, no recomputation.
totalsX := 120.0
rowH := 6.0
printTotalRow := func(label string, value float64, bold bool) {
pdf.SetX(totalsX)
if bold {
pdf.SetFont("Arial", "B", 10)
} else {
pdf.SetFont("Arial", "", 9)
}
pdf.CellFormat(45, rowH, label, "", 0, "L", false, 0, "")
pdf.CellFormat(30, rowH, fmt.Sprintf("%.2f", value), "", 1, "R", false, 0, "")
}
printTotalRow("Subtotal", invoice.ItemsAmount, false)
if invoice.DiscountAmount > 0 {
printTotalRow("Discount", -invoice.DiscountAmount, false)
}
printTotalRow("Delivery Charge", invoice.DeliveryCharge, false)
if invoice.WalletUsed > 0 {
printTotalRow("Wallet Used", -invoice.WalletUsed, false)
}
printTotalRow("Grand Total", invoice.TotalAmount, true)

pdf.Ln(10)
pdf.SetFont("Arial", "", 8)
pdf.Cell(0, 5, "Note: No GST/tax rate is configured for this store. Amounts above do not include tax.")

var buf []byte
w := &pdfBufferWriter{}
if err := pdf.Output(w); err != nil {
return nil, err
}
buf = w.buf
return buf, nil
}

// pdfBufferWriter is a minimal io.Writer that gofpdf writes the finished
// PDF bytes into, so we can hand them back as an in-memory []byte instead
// of writing to disk.
type pdfBufferWriter struct{ buf []byte }

func (w *pdfBufferWriter) Write(p []byte) (int, error) {
w.buf = append(w.buf, p...)
return len(p), nil
}

func servePDF(c *gin.Context, invoice models.Invoice) {
pdfBytes, err := buildInvoicePDF(invoice)
if err != nil {
c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate invoice PDF"})
return
}
filename := invoice.InvoiceNumber + ".pdf"
c.Header("Content-Disposition", "attachment; filename=\""+filename+"\"")
c.Data(http.StatusOK, "application/pdf", pdfBytes)
}

// GetMyOrderInvoicePDF godoc
// GET /api/v1/orders/:id/invoice/pdf (protected - order owner only)
func GetMyOrderInvoicePDF(c *gin.Context) {
userID := c.MustGet("user_id").(uint)
orderID := c.Param("id")

var order models.Order
if err := database.DB.Where("id = ? AND user_id = ?", orderID, userID).First(&order).Error; err != nil {
c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
return
}

var invoice models.Invoice
if err := database.DB.Where("order_id = ?", order.ID).Preload("Items").First(&invoice).Error; err != nil {
generated, genErr := services.GenerateInvoiceIfNotExists(order.ID)
if genErr != nil || generated == nil {
c.JSON(http.StatusNotFound, gin.H{"error": "No invoice found for this order yet"})
return
}
database.DB.Where("order_id = ?", order.ID).Preload("Items").First(&invoice)
}
servePDF(c, invoice)
}

// GetWarehouseOrderInvoicePDF godoc
// GET /api/v1/warehouse/orders/:id/invoice/pdf (warehouse staff only)
func GetWarehouseOrderInvoicePDF(c *gin.Context) {
warehouseID := c.MustGet("warehouse_id").(uint)
orderID := c.Param("id")

var order models.Order
if err := database.DB.Where("warehouse_id = ?", warehouseID).First(&order, orderID).Error; err != nil {
c.JSON(http.StatusNotFound, gin.H{"error": "Order not found for your warehouse"})
return
}

var invoice models.Invoice
if err := database.DB.Where("order_id = ?", order.ID).Preload("Items").First(&invoice).Error; err != nil {
c.JSON(http.StatusNotFound, gin.H{"error": "No invoice found for this order yet"})
return
}
servePDF(c, invoice)
}

// GetAdminInvoicePDF godoc
// GET /api/v1/admin/invoices/:id/pdf (admin only)
func GetAdminInvoicePDF(c *gin.Context) {
id := c.Param("id")

var invoice models.Invoice
if err := database.DB.Preload("Items").First(&invoice, id).Error; err != nil {
c.JSON(http.StatusNotFound, gin.H{"error": "Invoice not found"})
return
}
servePDF(c, invoice)
}
