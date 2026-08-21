package handlers

import (
"net/http"

"github.com/gin-gonic/gin"
"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/database"
)

// TEMPORARY DEBUG ENDPOINT - remove after use.
// Dumps every account in the Chart of Accounts with its type and computed
// balance as of now, so we can see the full trial balance and find where
// the Balance Sheet imbalance is coming from.
// GET /api/v1/admin/debug/trial-balance
func DebugTrialBalance(c *gin.Context) {
type row struct {
Code    string  `json:"code"`
Name    string  `json:"name"`
Type    string  `json:"type"`
Balance float64 `json:"balance"`
}
var rows []row
database.DB.Table("accounts").
Select(`accounts.code, accounts.name, accounts.type,
COALESCE(SUM(CASE WHEN ledger_entries.type = 'debit' THEN ledger_entries.amount ELSE -ledger_entries.amount END),0) as balance`).
Joins("LEFT JOIN ledger_entries ON ledger_entries.account_id = accounts.id").
Group("accounts.code, accounts.name, accounts.type").
Order("accounts.type ASC, accounts.code ASC").
Scan(&rows)

var totalDebitNature, totalCreditNature float64
byType := gin.H{}
for _, r := range rows {
if r.Type == "asset" || r.Type == "expense" {
totalDebitNature += r.Balance
} else {
totalCreditNature += r.Balance
}
}

c.JSON(http.StatusOK, gin.H{
"accounts":            rows,
"total_debit_nature":  totalDebitNature,
"total_credit_nature": totalCreditNature,
"note":                "for a fully balanced ledger, sum of all debit-natural balances (asset+expense) should equal sum of all credit-natural balances (liability+revenue) in absolute terms",
})
_ = byType
}
