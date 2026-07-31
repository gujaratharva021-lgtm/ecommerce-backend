package handlers

import (
"net/http"
"strconv"

"github.com/gin-gonic/gin"
"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/database"
"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/models"
"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/utils"
"gorm.io/gorm"
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