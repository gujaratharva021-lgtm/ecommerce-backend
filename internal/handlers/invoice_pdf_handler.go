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

// buildInvoicePDF renders a GST-style tax invoice matching the standard
// marketplace invoice layout: seller header, Bill To/Ship To, item table
// with HSN/GST columns, totals, and an "Order Delivered From" +
// "E-commerce Platform Info" footer.
//
// GST/CGST/SGST columns are always printed as 0.00 - this project has no
// tax calculation configured anywhere in the order flow, so showing a
// non-zero rate would be fabricated. HSN is left blank per item since the
// product catalog has no HSN field. Both are safe to wire up for real once
// that data exists - don't invent it here.
//
// warehouseAddress is the "Order Delivered From" address - passed in
// separately since Invoice doesn't snapshot the warehouse (only the
// customer's delivery address), and warehouse address can differ from
// order to order.
func buildInvoicePDF(invoice models.Invoice, warehouse models.Warehouse) ([]byte, error) {
cfg := config.AppConfig

pdf := gofpdf.New("P", "mm", "A4", "")
pdf.AddPage()
pdf.SetMargins(12, 12, 12)

// ---- Seller header ----
pdf.SetFont("Arial", "B", 12)
pdf.CellFormat(150, 6, "Seller Name: "+cfg.SellerCompanyName, "", 1, "L", false, 0, "")
pdf.SetFont("Arial", "", 9)
pdf.CellFormat(150, 5, cfg.SellerAddress, "", 1, "L", false, 0, "")
pdf.SetFont("Arial", "B", 9)
if cfg.SellerGSTIN != "" {
pdf.CellFormat(0, 5, "GSTIN: "+cfg.SellerGSTIN, "", 1, "L", false, 0, "")
}
if cfg.SellerFSSAINumber != "" {
pdf.CellFormat(0, 5, "FSSAI: "+cfg.SellerFSSAINumber, "", 1, "L", false, 0, "")
}
pdf.Ln(2)
pdf.SetDrawColor(0, 0, 0)
pdf.Line(12, pdf.GetY(), 198, pdf.GetY())
pdf.Ln(3)

// ---- Title ----
pdf.SetFont("Arial", "B", 12)
pdf.CellFormat(0, 7, "TAX INVOICE / BILL OF SUPPLY", "", 1, "C", false, 0, "")
pdf.Line(12, pdf.GetY(), 198, pdf.GetY())
pdf.Ln(3)

// ---- Invoice meta row ----
pdf.SetFont("Arial", "", 9)
metaY := pdf.GetY()
pdf.CellFormat(95, 5, "Invoice No.: "+invoice.InvoiceNumber, "", 0, "L", false, 0, "")
placeOfSupply := invoice.AddressState
if placeOfSupply == "" {
placeOfSupply = "-"
}
pdf.CellFormat(0, 5, "Place Of Supply : "+placeOfSupply, "", 1, "L", false, 0, "")
pdf.SetXY(12, metaY+5)
pdf.CellFormat(95, 5, fmt.Sprintf("Order No.: %d", invoice.OrderID), "", 0, "L", false, 0, "")
pdf.CellFormat(0, 5, "Date : "+invoice.GeneratedAt.Format("02-01-2006"), "", 1, "L", false, 0, "")
if invoice.PaymentReference != "" {
pdf.CellFormat(0, 5, "Payment Reference: "+invoice.PaymentReference, "", 1, "L", false, 0, "")
}
pdf.Ln(2)
pdf.Line(12, pdf.GetY(), 198, pdf.GetY())
pdf.Ln(3)

// ---- Bill To / Ship To ----
billShipY := pdf.GetY()
addressLine := invoice.AddressLine1
if invoice.AddressLine2 != "" {
addressLine += ", " + invoice.AddressLine2
}
cityLine := fmt.Sprintf("%s, %s, %s", invoice.AddressCity, invoice.AddressState, invoice.AddressPincode)

pdf.SetFont("Arial", "B", 9)
pdf.CellFormat(93, 5, "Bill To", "", 0, "L", false, 0, "")
pdf.CellFormat(0, 5, "Ship To", "", 1, "L", false, 0, "")
pdf.SetFont("Arial", "", 9)
pdf.SetXY(12, billShipY+5)
pdf.CellFormat(93, 5, invoice.CustomerName, "", 0, "L", false, 0, "")
pdf.CellFormat(0, 5, invoice.CustomerName, "", 1, "L", false, 0, "")
pdf.SetXY(12, billShipY+10)
pdf.CellFormat(93, 5, addressLine, "", 0, "L", false, 0, "")
pdf.CellFormat(0, 5, addressLine, "", 1, "L", false, 0, "")
pdf.SetXY(12, billShipY+15)
pdf.CellFormat(93, 5, cityLine, "", 0, "L", false, 0, "")
pdf.CellFormat(0, 5, cityLine, "", 1, "L", false, 0, "")
pdf.Ln(2)
pdf.Line(12, pdf.GetY(), 198, pdf.GetY())
pdf.Ln(3)

// ---- Items table ----
colW := []float64{8, 47, 16, 10, 8, 18, 16, 12, 10, 12, 9, 20}
headers := []string{"SR\nNo", "Item &\nDescription", "Unit\nMRP/RSP", "HSN", "Qty", "Product\nRate", "Disc.", "Taxable\nAmt.", "CGST", "S/UT\nGST", "GST\nAmt.", "Total\nAmt."}

pdf.SetFont("Arial", "B", 6.5)
pdf.SetFillColor(235, 235, 235)
headerY := pdf.GetY()
for i, h := range headers {
pdf.SetXY(pdf.GetX(), headerY)
pdf.MultiCell(colW[i], 3.5, h, "1", "C", true)
pdf.SetXY(pdf.GetX()+colW[i], headerY)
}
pdf.SetXY(12, headerY+7)

pdf.SetFont("Arial", "", 7.5)
itemsTotal := 0.0
for i, item := range invoice.Items {
lineTotal := item.Price * float64(item.Quantity)
itemsTotal += lineTotal
rowY := pdf.GetY()
hsn := item.SKU
if hsn == "" {
hsn = "-"
}
values := []string{
fmt.Sprintf("%d", i+1),
item.ProductName,
fmt.Sprintf("%.2f", item.Price),
hsn,
fmt.Sprintf("%d", item.Quantity),
fmt.Sprintf("%.2f", item.Price),
"0.00%",
fmt.Sprintf("%.2f", lineTotal),
"0.00%",
"0.00%",
"0.00",
fmt.Sprintf("%.2f", lineTotal),
}
aligns := []string{"C", "L", "R", "C", "C", "R", "C", "R", "C", "C", "R", "R"}
maxH := 5.0
for j, v := range values {
pdf.SetXY(pdf.GetX(), rowY)
pdf.MultiCell(colW[j], maxH, v, "1", aligns[j], false)
pdf.SetXY(sumWidths(colW[:j+1])+12, rowY)
}
pdf.SetXY(12, rowY+maxH)
}

// Item Total row inside the table
totalRowY := pdf.GetY()
pdf.SetFont("Arial", "B", 7.5)
pdf.SetXY(sumWidths(colW[:7])+12, totalRowY)
pdf.CellFormat(colW[7], 5, fmt.Sprintf("%.2f", itemsTotal), "1", 0, "R", false, 0, "")
pdf.CellFormat(colW[8], 5, "", "1", 0, "C", false, 0, "")
pdf.CellFormat(colW[9], 5, "", "1", 0, "C", false, 0, "")
pdf.CellFormat(colW[10], 5, "0.00", "1", 0, "R", false, 0, "")
pdf.CellFormat(colW[11], 5, fmt.Sprintf("%.2f", itemsTotal), "1", 1, "R", false, 0, "")
pdf.SetX(12)
pdf.Ln(4)

// ---- Item Total / Invoice Value ----
pdf.SetFont("Arial", "B", 9)
pdf.CellFormat(150, 5, "Item Total", "", 0, "L", false, 0, "")
pdf.CellFormat(0, 5, fmt.Sprintf("%.2f", invoice.ItemsAmount), "", 1, "R", false, 0, "")
if invoice.DiscountAmount > 0 {
pdf.SetFont("Arial", "", 9)
pdf.CellFormat(150, 5, "Discount", "", 0, "L", false, 0, "")
pdf.CellFormat(0, 5, fmt.Sprintf("-%.2f", invoice.DiscountAmount), "", 1, "R", false, 0, "")
}
if invoice.DeliveryCharge > 0 {
pdf.SetFont("Arial", "", 9)
pdf.CellFormat(150, 5, "Delivery Charge", "", 0, "L", false, 0, "")
pdf.CellFormat(0, 5, fmt.Sprintf("%.2f", invoice.DeliveryCharge), "", 1, "R", false, 0, "")
}
if invoice.WalletUsed > 0 {
pdf.SetFont("Arial", "", 9)
pdf.CellFormat(150, 5, "Wallet Used", "", 0, "L", false, 0, "")
pdf.CellFormat(0, 5, fmt.Sprintf("-%.2f", invoice.WalletUsed), "", 1, "R", false, 0, "")
}
pdf.SetFont("Arial", "B", 10)
pdf.CellFormat(150, 6, "Invoice Value", "", 0, "L", false, 0, "")
pdf.CellFormat(0, 6, fmt.Sprintf("%.2f", invoice.TotalAmount), "", 1, "R", false, 0, "")
pdf.Ln(3)

pdf.SetFont("Arial", "", 7.5)
pdf.CellFormat(0, 4, "Whether GST is payable on reverse-charge - No.", "", 1, "L", false, 0, "")
pdf.CellFormat(0, 4, "For IMEI / Serial number information, please refer to packaging / warranty slip.", "", 1, "L", false, 0, "")
pdf.Ln(3)
pdf.Line(12, pdf.GetY(), 198, pdf.GetY())
pdf.Ln(3)

// ---- Footer: Order Delivered From / E-commerce Platform Info ----
footerY := pdf.GetY()
pdf.SetFont("Arial", "B", 8)
pdf.CellFormat(93, 4, "Order Delivered From -", "", 0, "L", false, 0, "")
pdf.CellFormat(0, 4, "E-commerce Platform (FBO) Information -", "", 1, "L", false, 0, "")
pdf.SetFont("Arial", "", 8)
pdf.SetXY(12, footerY+4)
pdf.CellFormat(93, 4, warehouse.Name, "", 0, "L", false, 0, "")
pdf.CellFormat(0, 4, cfg.SellerCompanyName, "", 1, "L", false, 0, "")
pdf.SetXY(12, footerY+8)
pdf.MultiCell(90, 4, warehouse.Address, "", "L", false)
pdf.SetXY(105, footerY+8)
pdf.MultiCell(90, 4, cfg.SellerAddress, "", "L", false)
fssaiY := pdf.GetY()
if fssaiY < footerY+16 {
fssaiY = footerY + 16
}
pdf.SetXY(12, fssaiY)
pdf.CellFormat(93, 4, "FSSAI: "+valueOrDash(cfg.SellerFSSAINumber), "", 0, "L", false, 0, "")
pdf.CellFormat(0, 4, "FSSAI Lic. No.: "+valueOrDash(cfg.SellerFSSAINumber), "", 1, "L", false, 0, "")
pdf.SetXY(105, fssaiY+4)
if cfg.SellerEmail != "" {
pdf.CellFormat(0, 4, "Email: "+cfg.SellerEmail, "", 1, "L", false, 0, "")
}

var buf []byte
w := &pdfBufferWriter{}
if err := pdf.Output(w); err != nil {
return nil, err
}
buf = w.buf
return buf, nil
}

func sumWidths(widths []float64) float64 {
total := 0.0
for _, w := range widths {
total += w
}
return total
}

func valueOrDash(v string) string {
if v == "" {
return "-"
}
return v
}

// pdfBufferWriter is a minimal io.Writer that gofpdf writes the finished
// PDF bytes into, so we can hand them back as an in-memory []byte instead
// of writing to disk.
type pdfBufferWriter struct{ buf []byte }

func (w *pdfBufferWriter) Write(p []byte) (int, error) {
w.buf = append(w.buf, p...)
return len(p), nil
}

func servePDF(c *gin.Context, invoice models.Invoice, warehouse models.Warehouse) {
pdfBytes, err := buildInvoicePDF(invoice, warehouse)
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
if err := database.DB.Preload("Warehouse").Where("id = ? AND user_id = ?", orderID, userID).First(&order).Error; err != nil {
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
servePDF(c, invoice, order.Warehouse)
}

// GetWarehouseOrderInvoicePDF godoc
// GET /api/v1/warehouse/orders/:id/invoice/pdf (warehouse staff only)
func GetWarehouseOrderInvoicePDF(c *gin.Context) {
warehouseID := c.MustGet("warehouse_id").(uint)
orderID := c.Param("id")

var order models.Order
if err := database.DB.Preload("Warehouse").Where("warehouse_id = ?", warehouseID).First(&order, orderID).Error; err != nil {
c.JSON(http.StatusNotFound, gin.H{"error": "Order not found for your warehouse"})
return
}

var invoice models.Invoice
if err := database.DB.Where("order_id = ?", order.ID).Preload("Items").First(&invoice).Error; err != nil {
c.JSON(http.StatusNotFound, gin.H{"error": "No invoice found for this order yet"})
return
}
servePDF(c, invoice, order.Warehouse)
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

var order models.Order
database.DB.Preload("Warehouse").First(&order, invoice.OrderID)

servePDF(c, invoice, order.Warehouse)
}
