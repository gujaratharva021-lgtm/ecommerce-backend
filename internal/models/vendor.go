package models

import "time"

// Vendor is a supplier the business buys from (packaging, produce, etc).
// Kept separate from User/DeliveryPartner since vendors are a B2B
// relationship with no login of their own - finance staff manage them
// directly, unlike customer or delivery partner accounts.
type Vendor struct {
ID          uint      `gorm:"primaryKey" json:"id"`
Name        string    `gorm:"not null" json:"name"`
ContactName string    `json:"contact_name,omitempty"`
Phone       string    `json:"phone,omitempty"`
Email       string    `json:"email,omitempty"`
GSTIN       string    `json:"gstin,omitempty"`
Address     string    `json:"address,omitempty"`
IsActive    bool      `gorm:"default:true" json:"is_active"`
CreatedAt   time.Time `json:"created_at"`
UpdatedAt   time.Time `json:"updated_at"`
}

// VendorRequest is the body for POST/PUT /admin/finance/vendors
type VendorRequest struct {
Name        string `json:"name" binding:"required"`
ContactName string `json:"contact_name"`
Phone       string `json:"phone"`
Email       string `json:"email"`
GSTIN       string `json:"gstin"`
Address     string `json:"address"`
IsActive    *bool  `json:"is_active"`
}

// VendorBill is a payable - an invoice/bill the vendor has raised on us,
// tracked until paid. AmountPaid lets a bill be partially paid over time;
// Status is derived (unpaid/partially_paid/paid) rather than stored
// independently, so it can never drift out of sync with AmountPaid.
type VendorBill struct {
ID          uint       `gorm:"primaryKey" json:"id"`
VendorID    uint       `gorm:"not null;index" json:"vendor_id"`
Vendor      Vendor     `gorm:"foreignKey:VendorID" json:"vendor,omitempty"`
BillNumber  string     `json:"bill_number,omitempty"`
Amount      float64    `gorm:"not null" json:"amount"`
AmountPaid  float64    `gorm:"not null;default:0" json:"amount_paid"`
BillDate    time.Time  `gorm:"not null" json:"bill_date"`
DueDate     *time.Time `json:"due_date,omitempty"`
Note        string     `json:"note,omitempty"`
CreatedByID uint       `gorm:"not null" json:"created_by_id"`
CreatedAt   time.Time  `json:"created_at"`
UpdatedAt   time.Time  `json:"updated_at"`
}

// VendorBillStatus computes the derived status from Amount vs AmountPaid.
func VendorBillStatus(b VendorBill) string {
if b.AmountPaid <= 0 {
return "unpaid"
}
if b.AmountPaid >= b.Amount {
return "paid"
}
return "partially_paid"
}

// VendorBillRequest is the body for POST/PUT /admin/finance/vendor-bills
type VendorBillRequest struct {
VendorID   uint    `json:"vendor_id" binding:"required"`
BillNumber string  `json:"bill_number"`
Amount     float64 `json:"amount" binding:"required,gt=0"`
BillDate   string  `json:"bill_date" binding:"required"`
DueDate    string  `json:"due_date"`
Note       string  `json:"note"`
}

// VendorBillPaymentRequest is the body for POST /admin/finance/vendor-bills/:id/pay
type VendorBillPaymentRequest struct {
Amount float64 `json:"amount" binding:"required,gt=0"`
}
