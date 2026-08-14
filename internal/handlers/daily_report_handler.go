package handlers

import (
"fmt"
"net/http"
"time"

"github.com/gin-gonic/gin"
"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/database"
"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/models"
"github.com/xuri/excelize/v2"
)

type DailySalesSummary struct {
Date             string  `json:"date"`
TotalOrders      int64   `json:"total_orders"`
DeliveredOrders  int64   `json:"delivered_orders"`
CancelledOrders  int64   `json:"cancelled_orders"`
PendingOrders    int64   `json:"pending_orders"`
TotalRevenue     float64 `json:"total_revenue"`
CODRevenue       float64 `json:"cod_revenue"`
OnlineRevenue    float64 `json:"online_revenue"`
CODOrders        int64   `json:"cod_orders"`
OnlineOrders     int64   `json:"online_orders"`
TotalDeliveryFee float64 `json:"total_delivery_charge"`
TotalWalletUsed  float64 `json:"total_wallet_used"`
AvgOrderValue    float64 `json:"avg_order_value"`
}

func parseReportDate(c *gin.Context) (start, end time.Time, label string, err error) {
dateStr := c.Query("date")
var day time.Time
if dateStr == "" {
now := time.Now()
day = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
} else {
day, err = time.Parse("2006-01-02", dateStr)
if err != nil {
return
}
}
start = day
end = day.AddDate(0, 0, 1)
label = day.Format("2006-01-02")
return
}

func buildDailySalesSummary(start, end time.Time, label string) DailySalesSummary {
var s DailySalesSummary
s.Date = label
f := "created_at >= ? AND created_at < ?"

database.DB.Model(&models.Order{}).Where(f, start, end).Count(&s.TotalOrders)
database.DB.Model(&models.Order{}).Where(f+" AND status = ?", start, end, "delivered").Count(&s.DeliveredOrders)
database.DB.Model(&models.Order{}).Where(f+" AND status = ?", start, end, "cancelled").Count(&s.CancelledOrders)
database.DB.Model(&models.Order{}).Where(f+" AND status = ?", start, end, "pending").Count(&s.PendingOrders)

database.DB.Model(&models.Order{}).
Where(f+" AND status != ?", start, end, "cancelled").
Select("COALESCE(SUM(total_amount), 0)").Scan(&s.TotalRevenue)

database.DB.Model(&models.Order{}).
Where(f+" AND status != ? AND payment_method = ?", start, end, "cancelled", "cod").
Select("COALESCE(SUM(total_amount), 0)").Scan(&s.CODRevenue)

database.DB.Model(&models.Order{}).
Where(f+" AND status != ? AND payment_method = ?", start, end, "cancelled", "online").
Select("COALESCE(SUM(total_amount), 0)").Scan(&s.OnlineRevenue)

database.DB.Model(&models.Order{}).Where(f+" AND payment_method = ?", start, end, "cod").Count(&s.CODOrders)
database.DB.Model(&models.Order{}).Where(f+" AND payment_method = ?", start, end, "online").Count(&s.OnlineOrders)

database.DB.Model(&models.Order{}).
Where(f+" AND status != ?", start, end, "cancelled").
Select("COALESCE(SUM(delivery_charge), 0)").Scan(&s.TotalDeliveryFee)

database.DB.Model(&models.Order{}).
Where(f+" AND status != ?", start, end, "cancelled").
Select("COALESCE(SUM(wallet_amount_used), 0)").Scan(&s.TotalWalletUsed)

nonCancelled := s.TotalOrders - s.CancelledOrders
if nonCancelled > 0 {
s.AvgOrderValue = s.TotalRevenue / float64(nonCancelled)
}
return s
}

func GetDailySalesReport(c *gin.Context) {
start, end, label, err := parseReportDate(c)
if err != nil {
c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date format, use YYYY-MM-DD"})
return
}
c.JSON(http.StatusOK, gin.H{"summary": buildDailySalesSummary(start, end, label)})
}

func ExportDailySalesReport(c *gin.Context) {
start, end, label, err := parseReportDate(c)
if err != nil {
c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date format, use YYYY-MM-DD"})
return
}

summary := buildDailySalesSummary(start, end, label)

var orders []models.Order
if err := database.DB.
Where("created_at >= ? AND created_at < ?", start, end).
Order("created_at ASC").
Preload("User").
Find(&orders).Error; err != nil {
c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load orders for report"})
return
}

xf := excelize.NewFile()
defer xf.Close()

summarySheet := "Summary"
xf.SetSheetName("Sheet1", summarySheet)
rows := [][]interface{}{
{"Daily Sales Report", label},
{},
{"Total Orders", summary.TotalOrders},
{"Delivered Orders", summary.DeliveredOrders},
{"Cancelled Orders", summary.CancelledOrders},
{"Pending Orders", summary.PendingOrders},
{},
{"Total Revenue (excl. cancelled)", summary.TotalRevenue},
{"COD Revenue", summary.CODRevenue},
{"Online Revenue", summary.OnlineRevenue},
{"COD Orders", summary.CODOrders},
{"Online Orders", summary.OnlineOrders},
{},
{"Total Delivery Charges Collected", summary.TotalDeliveryFee},
{"Total Wallet Amount Used", summary.TotalWalletUsed},
{"Average Order Value", summary.AvgOrderValue},
}
for i, row := range rows {
cell, _ := excelize.CoordinatesToCellName(1, i+1)
r := row
xf.SetSheetRow(summarySheet, cell, &r)
}
xf.SetColWidth(summarySheet, "A", "A", 32)
xf.SetColWidth(summarySheet, "B", "B", 20)

ordersSheet := "Orders"
xf.NewSheet(ordersSheet)
header := []interface{}{"Order ID", "Time", "Customer Name", "Customer Phone", "Items Amount", "Delivery Charge", "Wallet Used", "Total Amount", "Payment Method", "Payment Status", "Order Status"}
xf.SetSheetRow(ordersSheet, "A1", &header)

for i, o := range orders {
row := []interface{}{
o.ID,
o.CreatedAt.Format("15:04:05"),
o.User.Name,
o.User.Phone,
o.ItemsAmount,
o.DeliveryCharge,
o.WalletAmountUsed,
o.TotalAmount,
o.PaymentMethod,
o.PaymentStatus,
o.Status,
}
r := row
cell, _ := excelize.CoordinatesToCellName(1, i+2)
xf.SetSheetRow(ordersSheet, cell, &r)
}
for _, col := range []string{"A", "B", "C", "D", "E", "F", "G", "H", "I", "J", "K"} {
xf.SetColWidth(ordersSheet, col, col, 16)
}
xf.SetActiveSheet(0)

filename := fmt.Sprintf("daily-sales-%s.xlsx", label)
c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
if err := xf.Write(c.Writer); err != nil {
c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate Excel file"})
}
}
