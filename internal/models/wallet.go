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
WalletReasonAddMoney     = "add_money"
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

// AddMoneyRequest is the body for POST /wallet/add-money
type AddMoneyRequest struct {
Amount float64 `json:"amount" binding:"required,gt=0"`
}

// WalletTopup tracks a gateway-verified wallet top-up from initiation
// through verification. Unlike AddMoneyRequest (used only to specify the
// amount up front), the wallet is never credited off this request alone -
// InitiateWalletTopup creates this row with Status "created" and a real
// Razorpay order for the exact amount, and VerifyWalletTopup only credits
// the wallet after the Razorpay signature is verified against that same
// order (Bug#19: previously the client-supplied amount was trusted
// directly with no payment verification at all).
type WalletTopup struct {
ID                uint      `gorm:"primaryKey" json:"id"`
UserID            uint      `gorm:"not null;index" json:"user_id"`
Amount            float64   `gorm:"not null" json:"amount"`
RazorpayOrderID   string    `gorm:"not null" json:"razorpay_order_id"`
RazorpayPaymentID string    `json:"razorpay_payment_id,omitempty"`
Status            string    `gorm:"default:created" json:"status"` // created/paid/failed
CreatedAt         time.Time `json:"created_at"`
UpdatedAt         time.Time `json:"updated_at"`
}

// InitiateWalletTopupResponse mirrors CreatePaymentOrderResponse's shape
// so the frontend Razorpay Checkout integration is consistent across both
// order payment and wallet top-up flows.
type InitiateWalletTopupResponse struct {
RazorpayOrderID string `json:"razorpay_order_id"`
Amount          int64  `json:"amount"` // paise
Currency        string `json:"currency"`
KeyID           string `json:"key_id"`
TopupID         uint   `json:"topup_id"`
}

// VerifyWalletTopupRequest is the body for POST /wallet/add-money/verify.
type VerifyWalletTopupRequest struct {
RazorpayOrderID   string `json:"razorpay_order_id" binding:"required"`
RazorpayPaymentID string `json:"razorpay_payment_id" binding:"required"`
RazorpaySignature string `json:"razorpay_signature" binding:"required"`
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