package handlers

import (
"net/http"

"github.com/gin-gonic/gin"
"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/database"
"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/models"
"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/services"
)

// GetOrderInvoice godoc
// GET /api/v1/warehouse/orders/:id/invoice (warehouse staff only)
// Read-only view of the invoice for an order fulfilled by the caller's
// warehouse: invoice number, status, line items, and totals. Warehouse
// staff have no write access here - price, tax, discount, grand total
// and invoice number can only ever be set by GenerateInvoiceIfNotExists,
// never edited through this or any other warehouse endpoint.
func GetOrderInvoice(c *gin.Context) {
warehouseID := c.MustGet("warehouse_id").(uint)
orderID := c.Param("id")

var order models.Order
if err := database.DB.Where("warehouse_id = ?", warehouseID).First(&order, orderID).Error; err != nil {
c.JSON(http.StatusNotFound, gin.H{"error": "Order not found for your warehouse"})
return
}

var invoice models.Invoice
err := database.DB.Where("order_id = ?", order.ID).Preload("Items").First(&invoice).Error
if err != nil {
// Orders paid COD are invoiced at confirmation and online orders right
// after payment verification, so a missing invoice on an order that
// has already reached the warehouse is unexpected - generate it
// on-demand instead of forcing the warehouse to work around a gap.
generated, genErr := services.GenerateInvoiceIfNotExists(order.ID)
if genErr != nil || generated == nil {
c.JSON(http.StatusNotFound, gin.H{"error": "No invoice found for this order yet"})
return
}
database.DB.Where("order_id = ?", order.ID).Preload("Items").First(&invoice)
}

c.JSON(http.StatusOK, gin.H{
"invoice_number":  invoice.InvoiceNumber,
"order_id":        invoice.OrderID,
"order_status":    order.Status,
"customer_name":   invoice.CustomerName,
"customer_phone":  invoice.CustomerPhone,
"items_amount":    invoice.ItemsAmount,
"delivery_charge": invoice.DeliveryCharge,
"wallet_used":     invoice.WalletUsed,
"total_amount":    invoice.TotalAmount,
"payment_method":  invoice.PaymentMethod,
"payment_status":  order.PaymentStatus,
"generated_at":    invoice.GeneratedAt,
"items":           invoice.Items,
})
}
