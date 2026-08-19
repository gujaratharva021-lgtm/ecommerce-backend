package models

import "time"

// Payroll records a salary payment (or pending payment) for a warehouse
// staff member, for a given month/year. Status starts "pending" and moves
// to "paid" once finance marks it settled - PaidByID/PaidAt capture who
// confirmed it and when, for accountability.
type Payroll struct {
ID            uint            `gorm:"primaryKey" json:"id"`
StaffID       uint            `gorm:"not null;index" json:"staff_id"`
Staff         WarehouseStaff  `gorm:"foreignKey:StaffID" json:"staff,omitempty"`
Amount        float64         `gorm:"not null" json:"amount"`
Month         int             `gorm:"not null" json:"month"`
Year          int             `gorm:"not null" json:"year"`
Status        string          `gorm:"not null;default:pending;index" json:"status"`
PaymentMethod string          `json:"payment_method,omitempty"`
Note          string          `json:"note,omitempty"`
PaidByID      *uint           `json:"paid_by_id,omitempty"`
PaidAt        *time.Time      `json:"paid_at,omitempty"`
CreatedAt     time.Time       `json:"created_at"`
UpdatedAt     time.Time       `json:"updated_at"`
}

// ValidPayrollStatuses restricts the status field.
var ValidPayrollStatuses = map[string]bool{
"pending": true,
"paid":    true,
}

// ValidPaymentMethods restricts the payment_method field.
var ValidPaymentMethods = map[string]bool{
"cash": true,
"bank": true,
"upi":  true,
}

// PayrollRequest is the body for POST/PUT /admin/finance/payroll
type PayrollRequest struct {
StaffID       uint    `json:"staff_id" binding:"required"`
Amount        float64 `json:"amount" binding:"required,gt=0"`
Month         int     `json:"month" binding:"required,min=1,max=12"`
Year          int     `json:"year" binding:"required"`
Status        string  `json:"status"`
PaymentMethod string  `json:"payment_method"`
Note          string  `json:"note"`
}
