package handlers

import (
"net/http"

"github.com/gin-gonic/gin"
"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/database"
)

// TEMPORARY DEBUG ENDPOINT - remove after use.
// Merges legacy duplicate accounts (1000 Cash, 4000 Sales Revenue) created
// by ad-hoc testing before seedChartOfAccounts existed, into the proper
// seeded accounts (1001 Cash, 4001 Product Sales), then deletes the
// duplicates so Balance Sheet/Trial Balance stop reporting a mismatch.
// POST /api/v1/admin/debug/merge-duplicate-accounts
func DebugMergeDuplicateAccounts(c *gin.Context) {
r1 := database.DB.Exec(`
UPDATE ledger_entries SET account_id = (SELECT id FROM accounts WHERE code = '1001')
WHERE account_id = (SELECT id FROM accounts WHERE code = '1000')`)
r2 := database.DB.Exec(`
UPDATE ledger_entries SET account_id = (SELECT id FROM accounts WHERE code = '4001')
WHERE account_id = (SELECT id FROM accounts WHERE code = '4000')`)
r3 := database.DB.Exec(`DELETE FROM accounts WHERE code = '1000'`)
r4 := database.DB.Exec(`DELETE FROM accounts WHERE code = '4000'`)

c.JSON(http.StatusOK, gin.H{
"cash_entries_moved":    r1.RowsAffected,
"sales_entries_moved":   r2.RowsAffected,
"account_1000_deleted":  r3.RowsAffected,
"account_4000_deleted":  r4.RowsAffected,
"errors": gin.H{"r1": r1.Error, "r2": r2.Error, "r3": r3.Error, "r4": r4.Error},
})
}
