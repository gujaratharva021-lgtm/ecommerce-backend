package models

import "time"

// Account is a single entry in the Chart of Accounts - the fixed list of
// buckets every ledger entry must debit or credit. Type constrains what
// side of the accounting equation an account sits on; Code is a short
// human-friendly reference (e.g. "1000" for Cash) used in the ledger UI
// instead of raw IDs.
type Account struct {
ID        uint      `gorm:"primaryKey" json:"id"`
Code      string    `gorm:"uniqueIndex;not null" json:"code"`
Name      string    `gorm:"not null" json:"name"`
Type      string    `gorm:"not null;index" json:"type"` // asset / liability / equity / revenue / expense
IsActive  bool      `gorm:"default:true" json:"is_active"`
CreatedAt time.Time `json:"created_at"`
UpdatedAt time.Time `json:"updated_at"`
}

// ValidAccountTypes restricts Account.Type to standard accounting categories.
var ValidAccountTypes = map[string]bool{
"asset":     true,
"liability": true,
"equity":    true,
"revenue":   true,
"expense":   true,
}

// AccountRequest is the body for POST/PUT /admin/finance/accounts
type AccountRequest struct {
Code     string `json:"code" binding:"required"`
Name     string `json:"name" binding:"required"`
Type     string `json:"type" binding:"required"`
IsActive *bool  `json:"is_active"`
}

// LedgerEntry is one leg of a double-entry transaction: every real-world
// event (a bill received, a payment made, a manual adjustment) produces at
// least one debit entry and one matching credit entry sharing the same
// TransactionRef, so debits always equal credits across a ref. This table
// stores individual legs rather than paired rows, which is the standard
// general-ledger shape and lets a single transaction touch more than two
// accounts if ever needed.
type LedgerEntry struct {
ID             uint      `gorm:"primaryKey" json:"id"`
TransactionRef string    `gorm:"not null;index" json:"transaction_ref"`
AccountID      uint      `gorm:"not null;index" json:"account_id"`
Account        Account   `gorm:"foreignKey:AccountID" json:"account,omitempty"`
Type           string    `gorm:"not null" json:"type"` // debit / credit
Amount         float64   `gorm:"not null" json:"amount"`
Description    string    `json:"description,omitempty"`
ReferenceType  string    `json:"reference_type,omitempty"` // e.g. "vendor_bill", "manual"
ReferenceID    *uint     `json:"reference_id,omitempty"`
EntryDate      time.Time `gorm:"not null;index" json:"entry_date"`
// CreatedByID is nil for system-generated entries (e.g. automatic sales
// ledger postings) rather than a real admin/staff action.
CreatedByID    *uint     `json:"created_by_id,omitempty"`
CreatedAt      time.Time `json:"created_at"`
}

// ValidLedgerEntryTypes restricts LedgerEntry.Type.
var ValidLedgerEntryTypes = map[string]bool{
"debit":  true,
"credit": true,
}

// LedgerEntryLine is one side of a manual journal entry submission.
type LedgerEntryLine struct {
AccountID   uint    `json:"account_id" binding:"required"`
Type        string  `json:"type" binding:"required"` // debit / credit
Amount      float64 `json:"amount" binding:"required,gt=0"`
Description string  `json:"description"`
}

// ManualJournalEntryRequest is the body for POST /admin/finance/ledger.
// Lines must balance (sum of debits == sum of credits) or the request is
// rejected - this is the one invariant that makes it a real ledger instead
// of a free-form transaction log.
type ManualJournalEntryRequest struct {
EntryDate string            `json:"entry_date" binding:"required"`
Lines     []LedgerEntryLine `json:"lines" binding:"required,min=2,dive"`
}

// PendingJournalEntry stages a manual journal entry for maker-checker
// approval (12.25) before it becomes real LedgerEntry rows. Lines are
// stored as JSON (same shape as ManualJournalEntryRequest.Lines) rather
// than a child table, since they are never queried independently of the
// parent - only re-parsed once, at approval time.
type PendingJournalEntry struct {
ID              uint       `gorm:"primaryKey" json:"id"`
EntryDate       string     `json:"entry_date"`
LinesJSON       string     `gorm:"type:text" json:"-"`
Lines           []LedgerEntryLine `gorm:"-" json:"lines,omitempty"`
Description     string     `json:"description,omitempty"`
TotalAmount     float64    `json:"total_amount"`
Status          string     `gorm:"not null;default:pending;index" json:"status"` // pending/approved/rejected
RequestedByID   uint       `gorm:"not null" json:"requested_by_id"`
ApprovedByID    *uint      `json:"approved_by_id,omitempty"`
ApprovedAt      *time.Time `json:"approved_at,omitempty"`
RejectionReason string     `json:"rejection_reason,omitempty"`
CreatedAt       time.Time  `json:"created_at"`
}

type PendingJournalEntryRejectRequest struct {
Reason string `json:"reason" binding:"required"`
}
