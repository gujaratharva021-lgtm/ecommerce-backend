package models

import "time"

// Invoice is generated once per Order, after payment is confirmed (COD
// orders get one as soon as they're confirmed; online orders get one
// right after VerifyPayment marks the order paid). OrderID is unique so
// a second generation attempt for the same order is a no-op rather than
// a duplicate row.
//
// Amounts and item lines are a snapshot taken at generation time - if the
// underlying Order/Product data changes later, the invoice does not
// change, which is what makes it a legal record rather than a live view.
type Invoice struct {
ID             uint          `gorm:"primaryKey" json:"id"`
InvoiceNumber  string        `gorm:"uniqueIndex;not null" json:"invoice_number"`
OrderID        uint          `gorm:"uniqueIndex;not null" json:"order_id"`
Order          Order         `gorm:"foreignKey:OrderID" json:"-"`
CustomerName   string        `json:"customer_name"`
CustomerPhone  string        `json:"customer_phone"`
ItemsAmount    float64       `json:"items_amount"`
DeliveryCharge float64       `json:"delivery_charge"`
WalletUsed     float64       `json:"wallet_amount_used"`
TotalAmount    float64       `json:"total_amount"`
PaymentMethod  string        `json:"payment_method"`
Items          []InvoiceItem `gorm:"foreignKey:InvoiceID" json:"items,omitempty"`
GeneratedAt    time.Time     `json:"generated_at"`
CreatedAt      time.Time     `json:"created_at"`
}

// InvoiceItem is a line-item snapshot - product name and price are copied
// at generation time so the invoice stays accurate even if the product is
// later renamed, repriced, or deleted.
type InvoiceItem struct {
ID          uint    `gorm:"primaryKey" json:"id"`
InvoiceID   uint    `gorm:"not null;index" json:"invoice_id"`
ProductID   uint    `json:"product_id"`
ProductName string  `json:"product_name"`
Quantity    int     `json:"quantity"`
Price       float64 `json:"price"`
}
