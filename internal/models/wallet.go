package models

import "time"

// Wallet transaction types
const (
WalletTxnCredit = "credit"
WalletTxnDebit  = "debit"
)

// Wallet transaction reasons — used for filtering/reporting and to keep
// the ledger self-explanatory without joining back to other tables.
const (
WalletReasonCashback     = "cashback"
WalletReasonRefund       = "refund"
WalletReasonCheckoutUse  = "checkout_use"
WalletReasonAdminCredit  = "admin_credit"
WalletReasonAdminDebit   = "admin_debit"
WalletReasonOrderRefund  = "order_cancelled_refund"
)

// Wallet holds a user's spendable balance. One wallet per user, created
// lazily on first credit/debit.
type Wallet struct {
ID        uint      `gorm:"primaryKey" json:"id"`
UserID    uint      `gorm:"not null;uniqueIndex" json:"user_id"`
Balance   float64   `gorm:"not null;default:0" json:"balance"`
CreatedAt time.Time `json:"created_at"`
UpdatedAt time.Time `json:"updated_at"`
}

// WalletTransaction is an immutable ledger entry. BalanceAfter is snapshotted
// at write time so the history is auditable even if Wallet.Balance is ever
// recalculated or disputed.
type WalletTransaction struct {
ID            uint      `gorm:"primaryKey" json:"id"`
WalletID      uint      `gorm:"not null;index" json:"wallet_id"`
Type          string    `gorm:"not null" json:"type"` // credit / debit
Amount        float64   `gorm:"not null" json:"amount"`
Reason        string    `gorm:"not null" json:"reason"`
ReferenceType string    `json:"reference_type,omitempty"` // e.g. "order"
ReferenceID   *uint     `json:"reference_id,omitempty"`
BalanceAfter  float64   `gorm:"not null" json:"balance_after"`
Note          string    `json:"note,omitempty"`
CreatedAt     time.Time `json:"created_at"`
}

// AdminWalletCreditRequest is the body for POST /admin/wallet/credit/:user_id
type AdminWalletCreditRequest struct {
Amount float64 `json:"amount" binding:"required,gt=0"`
Note   string  `json:"note"`
}

// WalletResponse wraps balance + recent transactions for GET /wallet
type WalletResponse struct {
Balance      float64              `json:"balance"`
Transactions []WalletTransaction  `json:"transactions"`
}