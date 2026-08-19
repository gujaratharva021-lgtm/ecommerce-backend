package models

import "time"

// Expense records a business cost (rent, utilities, salaries, packaging,
// marketing, maintenance, or miscellaneous) so finance staff can track
// outgoings alongside revenue. WarehouseID is nullable since some expenses
// (e.g. company-wide marketing) aren't tied to a single warehouse.
type Expense struct {
ID          uint      `gorm:"primaryKey" json:"id"`
Amount      float64   `gorm:"not null" json:"amount"`
Category    string    `gorm:"not null;index" json:"category"`
ExpenseDate time.Time `gorm:"not null;index" json:"expense_date"`
WarehouseID *uint     `json:"warehouse_id"`
Warehouse   *Warehouse `gorm:"foreignKey:WarehouseID" json:"warehouse,omitempty"`
Note        string    `json:"note,omitempty"`
ReceiptURL  string    `json:"receipt_url,omitempty"`
AddedByID   uint      `gorm:"not null" json:"added_by_id"`
CreatedAt   time.Time `json:"created_at"`
UpdatedAt   time.Time `json:"updated_at"`
}

// ValidExpenseCategories restricts the category field to a known set, same
// pattern as ValidWarehouseStaffRoles.
var ValidExpenseCategories = map[string]bool{
"rent":        true,
"utilities":   true,
"salaries":    true,
"packaging":   true,
"marketing":   true,
"maintenance": true,
"misc":        true,
}

// ExpenseRequest is the body for POST/PUT /admin/finance/expenses
type ExpenseRequest struct {
Amount      float64 `json:"amount" binding:"required,gt=0"`
Category    string  `json:"category" binding:"required"`
ExpenseDate string  `json:"expense_date" binding:"required"`
WarehouseID *uint   `json:"warehouse_id"`
Note        string  `json:"note"`
ReceiptURL  string  `json:"receipt_url"`
}
