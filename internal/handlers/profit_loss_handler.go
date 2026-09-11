package handlers

import (
"net/http"

"github.com/gin-gonic/gin"
"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/database"
)

// profitLossOrdersFilter is the same "did this order actually collect
// money" condition used by GetRevenueSummary, but pre-qualified with
// "orders." so it's safe to reuse inside joined queries (order_items JOIN
// orders JOIN products) without ambiguous-column errors.
const profitLossOrdersFilter = "(orders.payment_method = 'online' AND orders.payment_status = 'paid' AND orders.status NOT IN ('cancelled','returned')) OR (orders.payment_method = 'cod' AND orders.status = 'delivered')"

// GetProfitLoss godoc
// GET /api/v1/admin/finance/profit-loss?from=YYYY-MM-DD&to=YYYY-MM-DD
// Builds a P&L statement: Revenue - COGS = Gross Profit; Gross Profit -
// Operating Expenses = EBITDA (used here as a proxy for Net Profit, since
// this business has no depreciation/interest/tax line items tracked yet).
//
// COGS is computed from each order item's product cost_price at query time
// (not a stored snapshot), so it reflects the product's current cost even
// for old orders - acceptable for a small business dashboard, but flagged
// via cost_price_coverage so finance staff know if the number is trustworthy.
func GetProfitLoss(c *gin.Context) {
from, to := revenueDateRange(c)

var grossRevenue float64
database.DB.Table("orders").
Where("created_at >= ? AND created_at < ?", from, to).
Where("(payment_method = 'online' AND payment_status = 'paid' AND status NOT IN ('cancelled','returned')) OR (payment_method = 'cod' AND status = 'delivered')").
Select("COALESCE(SUM(items_amount),0)").
Scan(&grossRevenue)

// Discounts: coupon discounts applied to revenue-qualifying orders placed
// in this period (Defect #27a). Without this, a heavily-discounted order
// counted its full undiscounted items_amount as revenue, overstating both
// revenue and gross profit - the actual cash/wallet collected was always
// net of the coupon.
var discounts float64
database.DB.Table("order_coupons").
Joins("JOIN orders ON orders.id = order_coupons.order_id").
Where("orders.created_at >= ? AND orders.created_at < ?", from, to).
Where(profitLossOrdersFilter).
Select("COALESCE(SUM(order_coupons.discount_amount),0)").
Scan(&discounts)

// Refunds: GST credit notes (returns + substitutions) issued in this
// period (Defect #27b). A partial or full return that already shipped
// still counted its full items_amount as revenue above via the
// payment/delivery-status filter, which has no way to know a return
// happened after the fact - the credit note is the only record of money
// actually given back. Both tables are summed since a substitution
// credit note is issued separately from a return credit note but both
// represent revenue that must be backed out.
var creditNoteRefunds float64
database.DB.Table("credit_notes").
Where("issued_at >= ? AND issued_at < ?", from, to).
Select("COALESCE(SUM(total_amount),0)").
Scan(&creditNoteRefunds)

var substitutionRefunds float64
database.DB.Table("substitution_credit_notes").
Where("created_at >= ? AND created_at < ?", from, to).
Select("COALESCE(SUM(total_amount),0)").
Scan(&substitutionRefunds)
creditNoteRefunds += substitutionRefunds

netRevenue := grossRevenue - discounts - creditNoteRefunds

// COGS: sum of (order_item.quantity * product.cost_price) for items on
// revenue-qualifying orders in range. Only counts items whose product
// currently has cost_price > 0, so a partially-costed catalog doesn't
// silently understate COGS to zero.
var cogs float64
database.DB.Table("order_items").
Joins("JOIN orders ON orders.id = order_items.order_id").
Joins("JOIN products ON products.id = order_items.product_id").
Where("orders.created_at >= ? AND orders.created_at < ?", from, to).
Where(profitLossOrdersFilter).
Where("products.cost_price > 0").
Select("COALESCE(SUM(order_items.quantity * products.cost_price),0)").
Scan(&cogs)

// Coverage: what fraction of sold *items* (by quantity) had a cost_price
// set, so the UI can warn "COGS is incomplete" rather than imply it's exact.
var itemsWithCost, itemsTotal int64
database.DB.Table("order_items").
Joins("JOIN orders ON orders.id = order_items.order_id").
Joins("JOIN products ON products.id = order_items.product_id").
Where("orders.created_at >= ? AND orders.created_at < ?", from, to).
Where(profitLossOrdersFilter).
Where("products.cost_price > 0").
Select("COALESCE(SUM(order_items.quantity),0)").
Scan(&itemsWithCost)

database.DB.Table("order_items").
Joins("JOIN orders ON orders.id = order_items.order_id").
Where("orders.created_at >= ? AND orders.created_at < ?", from, to).
Where(profitLossOrdersFilter).
Select("COALESCE(SUM(order_items.quantity),0)").
Scan(&itemsTotal)

costPriceCoverage := 0.0
if itemsTotal > 0 {
costPriceCoverage = float64(itemsWithCost) / float64(itemsTotal) * 100
}

grossProfit := netRevenue - cogs

var operatingExpenses float64
database.DB.Table("expenses").
Where("expense_date >= ? AND expense_date < ?", from, to).
Where("approval_status IN ?", []string{"approved", "paid"}).
Select("COALESCE(SUM(amount),0)").
Scan(&operatingExpenses)

// Rider payout costs (Defect #27c) - these never touch the expenses
// table (see CreatePayroll/PayRiderPayout: they post straight to
// ledger_entries via PostRiderPayoutAccrualLedgerEntry), so without this
// the entire delivery-partner cost line was invisible to the P&L,
// overstating EBITDA/net profit by the full amount paid out to riders
// each period. Accrual date (when the payout was approved/created), not
// PaidAt, is used as the cost period - matches standard accrual
// accounting: the cost belongs to the period the work was done/recognized,
// not whenever finance got around to actually transferring the money.
var riderPayoutCosts float64
database.DB.Table("ledger_entries").
Where("reference_type = ? AND entry_date >= ? AND entry_date < ? AND type = ?",
"rider_payout_accrual", from, to, "debit").
Select("COALESCE(SUM(amount),0)").
Scan(&riderPayoutCosts)
operatingExpenses += riderPayoutCosts

ebitda := grossProfit - operatingExpenses

c.JSON(http.StatusOK, gin.H{
"from":                 from.Format("2006-01-02"),
"to":                   to.AddDate(0, 0, -1).Format("2006-01-02"),
"gross_revenue":        grossRevenue,
"discounts":            discounts,
"refunds":              creditNoteRefunds,
"revenue":              netRevenue,
"cogs":                 cogs,
"cost_price_coverage":  costPriceCoverage,
"gross_profit":         grossProfit,
"operating_expenses":   operatingExpenses,
"rider_payout_costs":   riderPayoutCosts,
"ebitda":               ebitda,
"net_profit":           ebitda,
})
}

