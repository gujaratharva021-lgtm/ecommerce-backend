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
AddressLine1   string        `json:"address_line1"`
AddressLine2   string        `json:"address_line2,omitempty"`
AddressCity    string        `json:"address_city"`
AddressState   string        `json:"address_state"`
AddressPincode string        `json:"address_pincode"`
ItemsAmount    float64       `json:"items_amount"`
DiscountAmount float64       `json:"discount_amount"`
DeliveryCharge float64       `json:"delivery_charge"`
WalletUsed     float64       `gorm:"column:wallet_amount_used" json:"wallet_amount_used"`
TotalAmount    float64       `json:"total_amount"`
PaymentMethod  string        `json:"payment_method"`
// PaymentReference is the gateway transaction ID (e.g. Razorpay payment
// ID) for online payments. Blank for COD, since there's no gateway
// transaction to reference.
PaymentReference string        `json:"payment_reference,omitempty"`
Items            []InvoiceItem `gorm:"foreignKey:InvoiceID" json:"items,omitempty"`
GeneratedAt      time.Time     `json:"generated_at"`
CreatedAt        time.Time     `json:"created_at"`
}

// InvoiceItem is a line-item snapshot - product name and price are copied
// at generation time so the invoice stays accurate even if the product is
// later renamed, repriced, or deleted.
type InvoiceItem struct {
ID          uint   `gorm:"primaryKey" json:"id"`
InvoiceID   uint   `gorm:"not null;index" json:"invoice_id"`
ProductID   uint   `json:"product_id"`
ProductName string `json:"product_name"`
// SKU is the product's barcode at the time of the order, if one was
// assigned. Blank if the product never had a barcode generated - not
// invented, since there's no separate SKU concept in this catalog.
SKU      string  `json:"sku,omitempty"`
Quantity int     `json:"quantity"`
Price    float64 `json:"price"`
}
