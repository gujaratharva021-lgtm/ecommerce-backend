package services

import (
"fmt"
"time"

"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/database"
"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/models"
"gorm.io/gorm"
"gorm.io/gorm/clause"
)

// GenerateInvoiceIfNotExists creates the Invoice + InvoiceItem snapshot
// for an order, the first time it's called for that order. Safe to call
// more than once (e.g. once from Checkout for COD, once from
// VerifyPayment for online) - OrderID has a unique index, and this
// function checks for an existing row inside the transaction under a
// lock before creating one, so concurrent calls cannot create duplicates.
//
// Fire-and-forget by design: an invoice failure should never block the
// order/payment flow that triggered it. Errors are returned so the caller
// can log them, but callers should not fail the request on error.
func GenerateInvoiceIfNotExists(orderID uint) (*models.Invoice, error) {
var invoice models.Invoice

txErr := database.DB.Transaction(func(tx *gorm.DB) error {
// Already generated - nothing to do. This check happens inside
// the transaction so a concurrent call blocks on the row lock
// below rather than racing this SELECT.
if err := tx.Where("order_id = ?", orderID).First(&invoice).Error; err == nil {
return nil
}

var order models.Order
if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
Preload("Items.Product").Preload("Address").
First(&order, orderID).Error; err != nil {
return fmt.Errorf("order not found: %w", err)
}

// Re-check after acquiring the lock in case another goroutine
// created it between the first check and the lock being granted.
if err := tx.Where("order_id = ?", orderID).First(&invoice).Error; err == nil {
return nil
}

var count int64
tx.Model(&models.Invoice{}).Count(&count)
invoiceNumber := fmt.Sprintf("INV-%d-%06d", time.Now().Year(), count+1)

invoice = models.Invoice{
InvoiceNumber:  invoiceNumber,
OrderID:        order.ID,
CustomerName:   order.Address.FullName,
CustomerPhone:  order.Address.Phone,
ItemsAmount:    order.ItemsAmount,
DeliveryCharge: order.DeliveryCharge,
WalletUsed:     order.WalletAmountUsed,
TotalAmount:    order.TotalAmount,
PaymentMethod:  order.PaymentMethod,
GeneratedAt:    time.Now(),
}
if err := tx.Create(&invoice).Error; err != nil {
return fmt.Errorf("failed to create invoice: %w", err)
}

for _, item := range order.Items {
invoiceItem := models.InvoiceItem{
InvoiceID:   invoice.ID,
ProductID:   item.ProductID,
ProductName: item.Product.Name,
Quantity:    item.Quantity,
Price:       item.Price,
}
if err := tx.Create(&invoiceItem).Error; err != nil {
return fmt.Errorf("failed to create invoice item: %w", err)
}
}

return nil
})

if txErr != nil {
return nil, txErr
}
return &invoice, nil
}
