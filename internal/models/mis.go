package models

import "time"

type MISManualEntry struct {
ID          uint      `json:"id" gorm:"primaryKey"`
Sheet       string    `json:"sheet"`
WeekStart   string    `json:"week_start" gorm:"type:date"`
RowKey      string    `json:"row_key"`
Data        string    `json:"data" gorm:"type:text"`
CreatedByID *uint     `json:"created_by_id"`
UpdatedByID *uint     `json:"updated_by_id"`
CreatedAt   time.Time `json:"created_at"`
UpdatedAt   time.Time `json:"updated_at"`
}

type MISExpenseApproval struct {
ID                uint   `json:"id" gorm:"primaryKey"`
Category          string `json:"category"`
UpTo25k           string `json:"up_to_25k"`
Range25k1L        string `json:"range_25k_1l"`
Range1L5L         string `json:"range_1l_5l"`
Above5L           string `json:"above_5l"`
RequiredDocuments string `json:"required_documents"`
Approver          string `json:"approver"`
}

func (MISManualEntry) TableName() string     { return "mis_manual_entries" }
func (MISExpenseApproval) TableName() string { return "mis_expense_approval" }
