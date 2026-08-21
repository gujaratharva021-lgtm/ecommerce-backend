package handlers

import (
"encoding/json"
"fmt"
"math"
"net/http"
"time"

"github.com/gin-gonic/gin"
"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/database"
"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/models"
"github.com/xuri/excelize/v2"
)

func round2(v float64) float64 { return math.Round(v*100) / 100 }

func weekBounds(dateStr string) (time.Time, time.Time) {
var t time.Time
if dateStr != "" {
parsed, err := time.Parse("2006-01-02", dateStr)
if err == nil {
t = parsed
}
}
if t.IsZero() {
t = time.Now()
}
weekday := int(t.Weekday())
if weekday == 0 {
weekday = 7
}
monday := t.AddDate(0, 0, -(weekday - 1))
monday = time.Date(monday.Year(), monday.Month(), monday.Day(), 0, 0, 0, 0, time.UTC)
sunday := monday.AddDate(0, 0, 6)
sunday = time.Date(sunday.Year(), sunday.Month(), sunday.Day(), 23, 59, 59, 0, time.UTC)
return monday, sunday
}

func currentUserID(c *gin.Context) *uint {
v, ok := c.Get("user_id")
if !ok {
return nil
}
switch val := v.(type) {
case uint:
return &val
case int:
u := uint(val)
return &u
case float64:
u := uint(val)
return &u
}
return nil
}

type misRow struct {
Label     string  `json:"label"`
Current   float64 `json:"current"`
Previous  float64 `json:"previous"`
GrowthPct float64 `json:"growth_pct"`
}

func computeRevenueMIS(start, end, prevStart, prevEnd time.Time) []misRow {
calc := func(s, e time.Time) map[string]float64 {
m := map[string]float64{}
database.DB.Raw(`SELECT COALESCE(SUM(items_amount),0) FROM orders WHERE created_at BETWEEN ? AND ? AND status <> 'cancelled'`, s, e).Scan(&[]float64{})
var v float64
database.DB.Raw(`SELECT COALESCE(SUM(items_amount),0) FROM orders WHERE created_at BETWEEN ? AND ? AND status <> 'cancelled'`, s, e).Row().Scan(&v)
m["product_sales"] = v
database.DB.Raw(`SELECT COALESCE(SUM(items_amount),0) FROM orders WHERE created_at BETWEEN ? AND ?`, s, e).Row().Scan(&v)
m["gmv"] = v
database.DB.Raw(`SELECT COALESCE(SUM(delivery_charge),0) FROM orders WHERE created_at BETWEEN ? AND ? AND status <> 'cancelled'`, s, e).Row().Scan(&v)
m["delivery"] = v
database.DB.Raw(`SELECT COALESCE(SUM(platform_fee),0) FROM orders WHERE created_at BETWEEN ? AND ? AND status <> 'cancelled'`, s, e).Row().Scan(&v)
m["platform_fee"] = v
database.DB.Raw(`SELECT COALESCE(SUM(oc.discount_amount),0) FROM order_coupons oc JOIN orders o ON o.id = oc.order_id WHERE o.created_at BETWEEN ? AND ?`, s, e).Row().Scan(&v)
m["discounts"] = v
database.DB.Raw(`SELECT COALESCE(SUM(total_amount),0) FROM credit_notes WHERE issued_at BETWEEN ? AND ?`, s, e).Row().Scan(&v)
m["refunds"] = v
database.DB.Raw(`SELECT COALESCE(SUM(items_amount),0) FROM orders WHERE created_at BETWEEN ? AND ? AND status = 'cancelled'`, s, e).Row().Scan(&v)
m["cancellations"] = v
database.DB.Raw(`SELECT COALESCE(SUM(oi.quantity * p.cost_price),0) FROM order_items oi JOIN orders o ON o.id = oi.order_id JOIN products p ON p.id = oi.product_id WHERE o.created_at BETWEEN ? AND ? AND o.status <> 'cancelled'`, s, e).Row().Scan(&v)
m["cogs"] = v

m["gross_revenue"] = m["product_sales"] + m["delivery"] + m["platform_fee"]
m["net_revenue"] = m["gross_revenue"] - m["discounts"] - m["refunds"] - m["cancellations"]
m["gross_profit"] = m["net_revenue"] - m["cogs"]
if m["net_revenue"] != 0 {
m["gross_margin_pct"] = m["gross_profit"] / m["net_revenue"] * 100
}
return m
}
cur := calc(start, end)
prev := calc(prevStart, prevEnd)

order := []struct{ key, label string }{
{"gmv", "Gross Merchandise Value (GMV)"},
{"product_sales", "Product Sales Revenue"},
{"delivery", "Delivery Charges"},
{"platform_fee", "Platform / Convenience Fee"},
{"gross_revenue", "Gross Revenue"},
{"discounts", "Less: Customer Discounts"},
{"refunds", "Less: Refunds / Returns"},
{"cancellations", "Less: Cancellations"},
{"net_revenue", "Net Revenue"},
{"cogs", "COGS"},
{"gross_profit", "Gross Profit"},
{"gross_margin_pct", "Gross Margin %"},
}
out := []misRow{}
for _, r := range order {
c := cur[r.key]
p := prev[r.key]
growth := 0.0
if p != 0 {
growth = (c - p) / p * 100
}
out = append(out, misRow{Label: r.label, Current: round2(c), Previous: round2(p), GrowthPct: round2(growth)})
}
return out
}

func computeVendorExpenseMIS(start, end, prevStart, prevEnd time.Time, netRevCur, netRevPrev float64) []misRow {
labelMap := map[string]string{
"freight": "Freight / Inward Transport", "loading": "Loading / Unloading",
"warehousing": "Warehousing / Storage", "handling": "Handling Charges",
"delivery": "Last-Mile Delivery", "reverse_logistics": "Reverse Logistics",
"payment_gateway": "Payment Gateway Charges", "commission": "Vendor Commission / Platform Charges",
"marketing": "Vendor-funded Promotion / Discount", "wastage": "Wastage / Spoilage",
"expiry": "Expiry Loss", "damage": "Damaged Goods", "shrinkage": "Shrinkage / Stock Loss",
"returns": "Customer Returns Cost",
}
orderedKeys := []string{"cogs", "freight", "loading", "warehousing", "handling", "delivery", "reverse_logistics", "payment_gateway", "commission", "marketing", "wastage", "expiry", "damage", "shrinkage", "returns", "other"}

sumByCategory := func(s, e time.Time) map[string]float64 {
rows, _ := database.DB.Raw(`SELECT LOWER(TRIM(category)) as cat, COALESCE(SUM(amount),0) as total FROM expenses WHERE expense_date BETWEEN ? AND ? AND approval_status <> 'rejected' GROUP BY LOWER(TRIM(category))`, s, e).Rows()
m := map[string]float64{}
if rows != nil {
defer rows.Close()
for rows.Next() {
var cat string
var total float64
rows.Scan(&cat, &total)
if _, ok := labelMap[cat]; !ok {
cat = "other"
}
m[cat] += total
}
}
var cogs float64
database.DB.Raw(`SELECT COALESCE(SUM(oi.quantity * p.cost_price),0) FROM order_items oi JOIN orders o ON o.id = oi.order_id JOIN products p ON p.id = oi.product_id WHERE o.created_at BETWEEN ? AND ? AND o.status <> 'cancelled'`, s, e).Row().Scan(&cogs)
m["cogs"] = cogs
return m
}

cur := sumByCategory(start, end)
prev := sumByCategory(prevStart, prevEnd)

var totalCur, totalPrev float64
out := []misRow{}
for _, k := range orderedKeys {
label := labelMap[k]
if k == "cogs" {
label = "Purchase of Goods / COGS"
}
if k == "other" {
label = "Other Vendor Costs"
}
c := cur[k]
p := prev[k]
totalCur += c
totalPrev += p
growth := 0.0
if p != 0 {
growth = (c - p) / p * 100
}
out = append(out, misRow{Label: label, Current: round2(c), Previous: round2(p), GrowthPct: round2(growth)})
}
totalGrowth := 0.0
if totalPrev != 0 {
totalGrowth = (totalCur - totalPrev) / totalPrev * 100
}
out = append(out, misRow{Label: "Total Vendor Expenses", Current: round2(totalCur), Previous: round2(totalPrev), GrowthPct: round2(totalGrowth)})
out = append(out, misRow{Label: "Net Revenue", Current: round2(netRevCur), Previous: round2(netRevPrev), GrowthPct: 0})
pctCur, pctPrev := 0.0, 0.0
if netRevCur != 0 {
pctCur = totalCur / netRevCur * 100
}
if netRevPrev != 0 {
pctPrev = totalPrev / netRevPrev * 100
}
out = append(out, misRow{Label: "Vendor Expense % of Net Revenue", Current: round2(pctCur), Previous: round2(pctPrev), GrowthPct: 0})
return out
}

type vendorSettlementRow struct {
VendorID         uint    `json:"vendor_id"`
VendorName       string  `json:"vendor_name"`
GrossSales       float64 `json:"gross_sales"`
Commission       float64 `json:"commission"`
Discount         float64 `json:"discount"`
Returns          float64 `json:"returns"`
DeliveryRecovery float64 `json:"delivery_recovery"`
OtherCharges     float64 `json:"other_charges"`
NetPayable       float64 `json:"net_payable"`
AmountPaid       float64 `json:"amount_paid"`
Balance          float64 `json:"balance"`
Status           string  `json:"status"`
}

func computeVendorSettlement(start, end time.Time) []vendorSettlementRow {
rows := []vendorSettlementRow{}
database.DB.Raw(`
SELECT v.id as vendor_id, v.name as vendor_name,
COALESCE(SUM(vb.amount),0) as gross_sales,
0 as commission, 0 as discount, 0 as returns, 0 as delivery_recovery,
COALESCE(SUM(vb.gst_amount),0) as other_charges,
COALESCE(SUM(vb.amount + vb.gst_amount),0) as net_payable,
COALESCE(SUM(vb.amount_paid),0) as amount_paid,
COALESCE(SUM(vb.amount + vb.gst_amount - vb.amount_paid),0) as balance,
CASE WHEN COALESCE(SUM(vb.amount + vb.gst_amount - vb.amount_paid),0) <= 0 THEN 'Paid'
WHEN COALESCE(SUM(vb.amount_paid),0) > 0 THEN 'Partial'
ELSE 'Pending' END as status
FROM vendors v
JOIN vendor_bills vb ON vb.vendor_id = v.id AND vb.bill_date BETWEEN ? AND ? AND vb.voided_at IS NULL
GROUP BY v.id, v.name
ORDER BY net_payable DESC
`, start, end).Scan(&rows)
return rows
}

type manualEntryOut struct {
ID     uint                   `json:"id"`
RowKey string                 `json:"row_key"`
Data   map[string]interface{} `json:"data"`
}

func getManualEntries(sheet, weekStart string) []manualEntryOut {
var entries []models.MISManualEntry
database.DB.Where("sheet = ? AND week_start = ?", sheet, weekStart).Order("id").Find(&entries)
out := []manualEntryOut{}
for _, e := range entries {
var d map[string]interface{}
json.Unmarshal([]byte(e.Data), &d)
out = append(out, manualEntryOut{ID: e.ID, RowKey: e.RowKey, Data: d})
}
return out
}

func GetWeeklyMIS(c *gin.Context) {
start, end := weekBounds(c.Query("week_start"))
prevStart := start.AddDate(0, 0, -7)
prevEnd := end.AddDate(0, 0, -7)

revenueRows := computeRevenueMIS(start, end, prevStart, prevEnd)
var netRevCur, netRevPrev float64
for _, r := range revenueRows {
if r.Label == "Net Revenue" {
netRevCur = r.Current
netRevPrev = r.Previous
}
}
expenseRows := computeVendorExpenseMIS(start, end, prevStart, prevEnd, netRevCur, netRevPrev)
settlementRows := computeVendorSettlement(start, end)

weekKey := start.Format("2006-01-02")
var approvalConfig []models.MISExpenseApproval
database.DB.Order("id").Find(&approvalConfig)

c.JSON(http.StatusOK, gin.H{
"week_start":            start.Format("2006-01-02"),
"week_end":              end.Format("2006-01-02"),
"prev_week_start":       prevStart.Format("2006-01-02"),
"prev_week_end":         prevEnd.Format("2006-01-02"),
"revenue_mis":           revenueRows,
"vendor_expense_mis":    expenseRows,
"vendor_settlement":     settlementRows,
"revenue_by_vendor":     getManualEntries("revenue_by_vendor", weekKey),
"vendor_pl":             getManualEntries("vendor_pl", weekKey),
"vendor_reconciliation": getManualEntries("vendor_reconciliation", weekKey),
"expense_approval":      approvalConfig,
})
}

type upsertManualEntryReq struct {
ID        *uint                  `json:"id"`
Sheet     string                 `json:"sheet" binding:"required"`
WeekStart string                 `json:"week_start" binding:"required"`
RowKey    string                 `json:"row_key" binding:"required"`
Data      map[string]interface{} `json:"data"`
}

func UpsertMISManualEntry(c *gin.Context) {
var req upsertManualEntryReq
if err := c.ShouldBindJSON(&req); err != nil {
c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
return
}
dataBytes, _ := json.Marshal(req.Data)
uid := currentUserID(c)

if req.ID != nil && *req.ID > 0 {
database.DB.Model(&models.MISManualEntry{}).Where("id = ?", *req.ID).Updates(map[string]interface{}{
"row_key": req.RowKey, "data": string(dataBytes), "updated_by_id": uid, "updated_at": time.Now(),
})
c.JSON(http.StatusOK, gin.H{"id": *req.ID})
return
}
entry := models.MISManualEntry{
Sheet: req.Sheet, WeekStart: req.WeekStart, RowKey: req.RowKey,
Data: string(dataBytes), CreatedByID: uid, UpdatedByID: uid,
}
database.DB.Create(&entry)
c.JSON(http.StatusOK, gin.H{"id": entry.ID})
}

func DeleteMISManualEntry(c *gin.Context) {
id := c.Param("id")
database.DB.Where("id = ?", id).Delete(&models.MISManualEntry{})
c.JSON(http.StatusOK, gin.H{"deleted": true})
}

func UpdateMISExpenseApproval(c *gin.Context) {
id := c.Param("id")
var req models.MISExpenseApproval
if err := c.ShouldBindJSON(&req); err != nil {
c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
return
}
database.DB.Model(&models.MISExpenseApproval{}).Where("id = ?", id).Updates(map[string]interface{}{
"up_to_25k": req.UpTo25k, "range_25k_1l": req.Range25k1L, "range_1l_5l": req.Range1L5L,
"above_5l": req.Above5L, "required_documents": req.RequiredDocuments, "approver": req.Approver,
})
c.JSON(http.StatusOK, gin.H{"updated": true})
}

func ExportWeeklyMIS(c *gin.Context) {
start, end := weekBounds(c.Query("week_start"))
prevStart := start.AddDate(0, 0, -7)
prevEnd := end.AddDate(0, 0, -7)

revenueRows := computeRevenueMIS(start, end, prevStart, prevEnd)
var netRevCur, netRevPrev float64
for _, r := range revenueRows {
if r.Label == "Net Revenue" {
netRevCur = r.Current
netRevPrev = r.Previous
}
}
expenseRows := computeVendorExpenseMIS(start, end, prevStart, prevEnd, netRevCur, netRevPrev)
settlementRows := computeVendorSettlement(start, end)

xf := excelize.NewFile()
defer xf.Close()

revSheet := "Revenue MIS"
xf.SetSheetName("Sheet1", revSheet)
xf.SetSheetRow(revSheet, "A1", &[]interface{}{"Line Item", "This Week", "Last Week", "Growth %"})
for i, r := range revenueRows {
cell, _ := excelize.CoordinatesToCellName(1, i+2)
xf.SetSheetRow(revSheet, cell, &[]interface{}{r.Label, r.Current, r.Previous, r.GrowthPct})
}
for _, col := range []string{"A", "B", "C", "D"} {
xf.SetColWidth(revSheet, col, col, 28)
}

expSheet := "Vendor Expense MIS"
xf.NewSheet(expSheet)
xf.SetSheetRow(expSheet, "A1", &[]interface{}{"Line Item", "This Week", "Last Week", "Growth %"})
for i, r := range expenseRows {
cell, _ := excelize.CoordinatesToCellName(1, i+2)
xf.SetSheetRow(expSheet, cell, &[]interface{}{r.Label, r.Current, r.Previous, r.GrowthPct})
}
for _, col := range []string{"A", "B", "C", "D"} {
xf.SetColWidth(expSheet, col, col, 28)
}

setSheet := "Vendor Settlement"
xf.NewSheet(setSheet)
xf.SetSheetRow(setSheet, "A1", &[]interface{}{"Vendor", "Gross Sales", "Commission", "Discount", "Returns", "Delivery Recovery", "Other Charges", "Net Payable", "Amount Paid", "Balance", "Status"})
for i, r := range settlementRows {
cell, _ := excelize.CoordinatesToCellName(1, i+2)
xf.SetSheetRow(setSheet, cell, &[]interface{}{r.VendorName, r.GrossSales, r.Commission, r.Discount, r.Returns, r.DeliveryRecovery, r.OtherCharges, r.NetPayable, r.AmountPaid, r.Balance, r.Status})
}
for _, col := range []string{"A", "B", "C", "D", "E", "F", "G", "H", "I", "J", "K"} {
xf.SetColWidth(setSheet, col, col, 20)
}

weekKey := start.Format("2006-01-02")

revByVendorSheet := "Revenue by Vendor"
xf.NewSheet(revByVendorSheet)
writeManualSheet(xf, revByVendorSheet, getManualEntries("revenue_by_vendor", weekKey))

vendorPLSheet := "Vendor P&L"
xf.NewSheet(vendorPLSheet)
writeManualSheet(xf, vendorPLSheet, getManualEntries("vendor_pl", weekKey))

vendorReconSheet := "Vendor Reconciliation"
xf.NewSheet(vendorReconSheet)
writeManualSheet(xf, vendorReconSheet, getManualEntries("vendor_reconciliation", weekKey))

var approvalConfig []models.MISExpenseApproval
database.DB.Order("id").Find(&approvalConfig)
approvalSheet := "Expense Approval"
xf.NewSheet(approvalSheet)
xf.SetSheetRow(approvalSheet, "A1", &[]interface{}{"Category", "Up to 25k", "25k-1L", "1L-5L", "Above 5L", "Required Documents", "Approver"})
for i, a := range approvalConfig {
cell, _ := excelize.CoordinatesToCellName(1, i+2)
xf.SetSheetRow(approvalSheet, cell, &[]interface{}{a.Category, a.UpTo25k, a.Range25k1L, a.Range1L5L, a.Above5L, a.RequiredDocuments, a.Approver})
}
for _, col := range []string{"A", "B", "C", "D", "E", "F", "G"} {
xf.SetColWidth(approvalSheet, col, col, 22)
}

xf.SetActiveSheet(0)

filename := fmt.Sprintf("weekly-mis-%s.xlsx", weekKey)
c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
if err := xf.Write(c.Writer); err != nil {
c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate Excel file"})
}
}

func writeManualSheet(xf *excelize.File, sheet string, entries []manualEntryOut) {
xf.SetSheetRow(sheet, "A1", &[]interface{}{"Row Key", "Data (JSON)"})
for i, e := range entries {
b, _ := json.Marshal(e.Data)
cell, _ := excelize.CoordinatesToCellName(1, i+2)
xf.SetSheetRow(sheet, cell, &[]interface{}{e.RowKey, string(b)})
}
xf.SetColWidth(sheet, "A", "A", 20)
xf.SetColWidth(sheet, "B", "B", 60)
}
