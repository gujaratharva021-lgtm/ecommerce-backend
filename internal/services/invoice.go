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

// Atomic invoice number: a dedicated Postgres sequence, not COUNT(*)+1.
// Two concurrent invoice generations for two *different* orders could
// both read the same COUNT() before either commits and then collide on
// the unique invoice_number constraint - nextval() is atomic across
// concurrent transactions and can never hand out the same value twice.
// CREATE SEQUENCE IF NOT EXISTS is self-healing: the sequence is created
// automatically the first time this runs, so there's no manual DB step
// required after deploying this - useful since AutoMigrate doesn't manage
// sequences and production migrations aren't always run immediately.
if err := tx.Exec("CREATE SEQUENCE IF NOT EXISTS invoice_number_seq START 1").Error; err != nil {
return fmt.Errorf("failed to ensure invoice sequence exists: %w", err)
}
// Same self-healing idea for columns: this project's production deploy
// runs with GIN_MODE=release, which intentionally skips GORM AutoMigrate
// (see internal/database/database.go) so schema changes go through
// versioned migrations instead - but until that migration is actually
// applied to prod, these ADD COLUMN IF NOT EXISTS statements make the
// invoice snapshot fields self-provisioning rather than a hard failure.
// Run as separate statements (not one multi-statement Exec) since not
// every Postgres driver path supports batching several DDL statements
// in a single simple-protocol call.
addColumnStatements := []string{
`ALTER TABLE invoices ADD COLUMN IF NOT EXISTS discount_amount DOUBLE PRECISION NOT NULL DEFAULT 0`,
`ALTER TABLE invoices ADD COLUMN IF NOT EXISTS address_line1 VARCHAR(255) NOT NULL DEFAULT ''`,
`ALTER TABLE invoices ADD COLUMN IF NOT EXISTS address_line2 VARCHAR(255) NOT NULL DEFAULT ''`,
`ALTER TABLE invoices ADD COLUMN IF NOT EXISTS address_city VARCHAR(100) NOT NULL DEFAULT ''`,
`ALTER TABLE invoices ADD COLUMN IF NOT EXISTS address_state VARCHAR(100) NOT NULL DEFAULT ''`,
`ALTER TABLE invoices ADD COLUMN IF NOT EXISTS address_pincode VARCHAR(20) NOT NULL DEFAULT ''`,
}
for _, stmt := range addColumnStatements {
if err := tx.Exec(stmt).Error; err != nil {
return fmt.Errorf("failed to ensure invoice columns exist: %w", err)
}
}
var seqVal int64
if err := tx.Raw("SELECT nextval('invoice_number_seq')").Scan(&seqVal).Error; err != nil {
return fmt.Errorf("failed to allocate invoice number: %w", err)
}
invoiceNumber := fmt.Sprintf("INV-%d-%06d", time.Now().Year(), seqVal)

var orderCoupon models.OrderCoupon
discountAmount := 0.0
if err := tx.Where("order_id = ?", order.ID).First(&orderCoupon).Error; err == nil {
discountAmount = orderCoupon.DiscountAmount
}

invoice = models.Invoice{
InvoiceNumber:  invoiceNumber,
OrderID:        order.ID,
CustomerName:   order.Address.FullName,
CustomerPhone:  order.Address.Phone,
AddressLine1:   order.Address.Line1,
AddressLine2:   order.Address.Line2,
AddressCity:    order.Address.City,
AddressState:   order.Address.State,
AddressPincode: order.Address.Pincode,
ItemsAmount:    order.ItemsAmount,
DiscountAmount: discountAmount,
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
