package services

import (
"fmt"

"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/database"
"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/models"
"gorm.io/gorm"
)

// PostSalesLedgerEntry records the double-entry ledger lines for one sale,
// triggered at the exact moment revenue is recognized:
//   - Online: right after payment verification (payment_handler.go)
//   - COD:    right after delivery confirmation (ConfirmDelivery)
// It is idempotent per order (checked via reference_type="sale",
// reference_id=orderID) so a retry or duplicate call never double-posts.
//
// Balanced entry:
//   Debit  Bank/Cash (order total actually collected)
//   Debit  Customer Wallet Liability (wallet amount redeemed)
//   Debit  Discount Given (discount amount)
//   Credit Product Sales (taxable amount)
//   Credit GST Payable (CGST+SGST+IGST)
// Debit total = TotalAmount + WalletUsed + DiscountAmount
// Credit total = TaxableAmount + totalGST
// These are equal because TaxableAmount + GST = ItemsAmount + DeliveryCharge
// + PlatformFee = TotalAmount + WalletUsed + DiscountAmount (invoice math).
func PostSalesLedgerEntry(orderID uint) error {
var invoice models.Invoice
if err := database.DB.Where("order_id = ?", orderID).First(&invoice).Error; err != nil {
return fmt.Errorf("no invoice found for order %d: %w", orderID, err)
}

// Idempotency check.
var existing models.LedgerEntry
if err := database.DB.Where("reference_type = ? AND reference_id = ?", "sale", orderID).First(&existing).Error; err == nil {
return nil
}

var order models.Order
if err := database.DB.First(&order, orderID).Error; err != nil {
return fmt.Errorf("order not found: %w", err)
}

cashOrBankCode := "1002" // Bank
if order.PaymentMethod == models.PaymentMethodCOD {
cashOrBankCode = "1001" // Cash
}

totalGST := invoice.CGSTAmount + invoice.SGSTAmount + invoice.IGSTAmount
transactionRef := fmt.Sprintf("SALE-%d", orderID)

type line struct {
code   string
lType  string
amount float64
desc   string
}
var lines []line

lines = append(lines, line{cashOrBankCode, "debit", invoice.TotalAmount, fmt.Sprintf("Sale for order #%d", orderID)})
if invoice.WalletUsed > 0 {
lines = append(lines, line{"2005", "debit", invoice.WalletUsed, fmt.Sprintf("Wallet redeemed on order #%d", orderID)})
}
if invoice.DiscountAmount > 0 {
lines = append(lines, line{"5002", "debit", invoice.DiscountAmount, fmt.Sprintf("Discount on order #%d", orderID)})
}
lines = append(lines, line{"4001", "credit", invoice.TaxableAmount, fmt.Sprintf("Sale for order #%d", orderID)})
if totalGST > 0 {
lines = append(lines, line{"2002", "credit", totalGST, fmt.Sprintf("GST on order #%d", orderID)})
}

return database.DB.Transaction(func(tx *gorm.DB) error {
for _, l := range lines {
if l.amount <= 0 {
continue
}
var account models.Account
if err := tx.Where("code = ?", l.code).First(&account).Error; err != nil {
return fmt.Errorf("chart of accounts missing code %s: %w", l.code, err)
}
entry := models.LedgerEntry{
TransactionRef: transactionRef,
AccountID:      account.ID,
Type:           l.lType,
Amount:         l.amount,
Description:    l.desc,
ReferenceType:  "sale",
ReferenceID:    &orderID,
EntryDate:      invoice.GeneratedAt,
CreatedByID:    nil, // system-generated, same convention as invoice audit log
}
if err := tx.Create(&entry).Error; err != nil {
return fmt.Errorf("failed to create ledger entry: %w", err)
}
}
return nil
})
}
