package handlers

import (
"log"
"fmt"
"net/http"
"strconv"

"github.com/gin-gonic/gin"
"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/config"
"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/database"
"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/models"
"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/utils"
"gorm.io/gorm"
"gorm.io/gorm/clause"
"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/services"
)

// ---------------------------------------------------------------------------
// Wallet - customer side
// ---------------------------------------------------------------------------

// GetWallet godoc
// GET /api/v1/wallet (protected)
// Returns the user's current balance and recent transaction history.
func GetWallet(c *gin.Context) {
userID := c.MustGet("user_id").(uint)

var wallet models.Wallet
err := database.DB.Where("user_id = ?", userID).First(&wallet).Error
if err != nil {
// No wallet yet — user has never had a credit/debit. Report a zero
// balance instead of a 404, since "no wallet" == "empty wallet" from
// the client's point of view.
c.JSON(http.StatusOK, models.WalletResponse{
Balance:      0,
Transactions: []models.WalletTransaction{},
})
return
}

var transactions []models.WalletTransaction
database.DB.Where("wallet_id = ?", wallet.ID).
Order("created_at DESC").
Limit(50).
Find(&transactions)

c.JSON(http.StatusOK, models.WalletResponse{
Balance:      wallet.Balance,
Transactions: transactions,
})
}

// InitiateWalletTopup godoc
// POST /api/v1/wallet/add-money (protected)
// Creates a Razorpay order for the requested amount and a pending
// WalletTopup row - the wallet is NOT credited here. Replaces the old
// AddMoneyToWallet, which trusted the client-supplied amount directly and
// credited the wallet with no payment verification at all (Bug#19: any
// authenticated user could mint arbitrary free store credit). Crediting
// now only happens in VerifyWalletTopup, after the Razorpay signature is
// verified against this exact order.
func InitiateWalletTopup(c *gin.Context) {
userID := c.MustGet("user_id").(uint)

var req models.AddMoneyRequest
if err := c.ShouldBindJSON(&req); err != nil {
c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
return
}

rzpOrderID, err := utils.CreateRazorpayOrder(req.Amount, fmt.Sprintf("wallet_topup_user_%d", userID))
if err != nil {
c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
return
}

topup := models.WalletTopup{
UserID:          userID,
Amount:          req.Amount,
RazorpayOrderID: rzpOrderID,
Status:          "created",
}
if err := database.DB.Create(&topup).Error; err != nil {
c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to record wallet top-up"})
return
}

c.JSON(http.StatusOK, models.InitiateWalletTopupResponse{
RazorpayOrderID: rzpOrderID,
Amount:          int64(req.Amount*100 + 0.5),
Currency:        "INR",
KeyID:           config.AppConfig.RazorpayKeyID,
TopupID:         topup.ID,
})
}

// VerifyWalletTopup godoc
// POST /api/v1/wallet/add-money/verify (protected)
// Verifies the Razorpay signature for a wallet top-up and, only on
// success, credits the wallet for the exact amount recorded when the
// order was created (never a client-supplied amount at this step). Locks
// and rechecks the WalletTopup row's status inside the transaction so a
// replayed verify call cannot credit the wallet twice, mirroring the
// idempotency fix applied to order payment verification (Bug#2).
func VerifyWalletTopup(c *gin.Context) {
userID := c.MustGet("user_id").(uint)

var req models.VerifyWalletTopupRequest
if err := c.ShouldBindJSON(&req); err != nil {
c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
return
}

var topup models.WalletTopup
if err := database.DB.Where("razorpay_order_id = ?", req.RazorpayOrderID).First(&topup).Error; err != nil {
c.JSON(http.StatusBadRequest, gin.H{"error": "No wallet top-up found for this Razorpay order"})
return
}
if topup.UserID != userID {
c.JSON(http.StatusForbidden, gin.H{"error": "You do not have access to this top-up"})
return
}

if !utils.VerifyRazorpaySignature(req.RazorpayOrderID, req.RazorpayPaymentID, req.RazorpaySignature) {
database.DB.Model(&topup).Update("status", "failed")
c.JSON(http.StatusBadRequest, gin.H{"error": "Payment signature verification failed"})
return
}

var alreadyProcessed bool
txErr := database.DB.Transaction(func(tx *gorm.DB) error {
if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&topup, topup.ID).Error; err != nil {
return err
}
if topup.Status == "paid" {
alreadyProcessed = true
return nil
}
topup.RazorpayPaymentID = req.RazorpayPaymentID
topup.Status = "paid"
if err := tx.Save(&topup).Error; err != nil {
return err
}
topupID := topup.ID
return utils.CreditWallet(tx, userID, topup.Amount, models.WalletReasonAddMoney, "wallet_topup", &topupID, "Added via app (Razorpay-verified)")
})
if txErr != nil {
c.JSON(http.StatusInternalServerError, gin.H{"error": "Payment was verified but failed to credit the wallet - contact support"})
return
}

var wallet models.Wallet
database.DB.Where("user_id = ?", userID).First(&wallet)
if alreadyProcessed {
c.JSON(http.StatusOK, gin.H{"message": "Top-up already verified", "wallet": wallet})
return
}

// Post the Bank/Wallet-Liability ledger entry for this top-up (Defect #03).
// Kept outside the main transaction, same pattern as PostSalesLedgerEntry -
// non-fatal on error (logged, not surfaced to the customer) since the
// wallet credit above already committed and is the source of truth for
// what the customer can spend; PostWalletTopupLedgerEntry is itself
// idempotent as a safety net against retries.
if err := services.PostWalletTopupLedgerEntry(topup.ID); err != nil {
log.Printf("failed to post wallet topup ledger entry for topup %d: %v", topup.ID, err)
}

c.JSON(http.StatusOK, gin.H{"wallet": wallet})
}

// ---------------------------------------------------------------------------
// Wallet - admin side
// ---------------------------------------------------------------------------

// AdminCreditWallet godoc
// POST /api/v1/admin/wallet/credit/:user_id (admin only)
// Manually credits a user's wallet — used for promotional cashback,
// goodwill credits, or correcting a support issue.
func AdminCreditWallet(c *gin.Context) {
adminID := c.MustGet("user_id").(uint)
userIDParam := c.Param("user_id")
userID64, err := strconv.ParseUint(userIDParam, 10, 64)
if err != nil {
c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user_id"})
return
}
userID := uint(userID64)

var user models.User
if err := database.DB.First(&user, userID).Error; err != nil {
c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
return
}

var req models.AdminWalletCreditRequest
if err := c.ShouldBindJSON(&req); err != nil {
c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
return
}

var wallet models.Wallet
txErr := database.DB.Transaction(func(tx *gorm.DB) error {
return utils.CreditWallet(tx, userID, req.Amount, models.WalletReasonAdminCredit, "admin", &adminID, req.Note)
})
if txErr != nil {
c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to credit wallet"})
return
}

database.DB.Where("user_id = ?", userID).First(&wallet)
c.JSON(http.StatusOK, gin.H{"wallet": wallet})
}