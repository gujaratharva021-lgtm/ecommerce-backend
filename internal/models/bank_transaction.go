package models

import "time"

// BankTransaction is one line from a bank statement, entered manually
// (CSV/statement import can be added later without changing this shape).
// Reconciliation means matching each bank line to an internal record
// (a VendorBill payment, a settlement, etc) - MatchedType/MatchedID capture
// that link once a finance user confirms it; until then the transaction
// sits "unmatched" and is what a reconciliation screen surfaces to review.
type BankTransaction struct {
ID              uint       `gorm:"primaryKey" json:"id"`
TransactionDate time.Time  `gorm:"not null;index" json:"transaction_date"`
Description     string     `json:"description"`
Amount          float64    `gorm:"not null" json:"amount"` // positive = credit (money in), negative = debit (money out)
ReferenceNumber string     `json:"reference_number,omitempty"`
Status          string     `gorm:"not null;default:unmatched;index" json:"status"` // unmatched / matched / ignored
MatchedType     string     `json:"matched_type,omitempty"`                         // e.g. "vendor_bill_payment", "payout"
MatchedID       *uint      `json:"matched_id,omitempty"`
MatchedAt       *time.Time `json:"matched_at,omitempty"`
MatchedByID     *uint      `json:"matched_by_id,omitempty"`
Note            string     `json:"note,omitempty"`
CreatedByID     uint       `gorm:"not null" json:"created_by_id"`
CreatedAt       time.Time  `json:"created_at"`
UpdatedAt       time.Time  `json:"updated_at"`
}

// ValidBankTransactionStatuses restricts BankTransaction.Status.
var ValidBankTransactionStatuses = map[string]bool{
"unmatched": true,
"matched":   true,
"ignored":   true,
}

// BankTransactionRequest is the body for POST /admin/finance/bank-transactions
type BankTransactionRequest struct {
TransactionDate string  `json:"transaction_date" binding:"required"`
Description     string  `json:"description"`
Amount          float64 `json:"amount" binding:"required"`
ReferenceNumber string  `json:"reference_number"`
}

// BankTransactionMatchRequest is the body for POST /admin/finance/bank-transactions/:id/match
type BankTransactionMatchRequest struct {
MatchedType string `json:"matched_type" binding:"required"`
MatchedID   *uint  `json:"matched_id"`
Note        string `json:"note"`
}
