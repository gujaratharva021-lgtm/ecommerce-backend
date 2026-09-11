package handlers

import (
"time"
"fmt"
"net/http"
"strings"

"github.com/gin-gonic/gin"
"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/cache"
"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/database"
"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/models"
"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/utils"
)

// GetCustomers godoc
// GET /api/v1/admin/customers (admin only)
func GetCustomers(c *gin.Context) {
var query models.CustomerListQuery
if err := c.ShouldBindQuery(&query); err != nil {
c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
return
}
if query.Page < 1 {
query.Page = 1
}
if query.Limit < 1 || query.Limit > 100 {
query.Limit = 20
}

db := database.DB.Model(&models.User{}).Where("role = ?", "customer")

if strings.TrimSpace(query.Search) != "" {
like := "%" + strings.TrimSpace(query.Search) + "%"
db = db.Where("name ILIKE ? OR phone ILIKE ?", like, like)
}
if query.Status == "blocked" {
db = db.Where("is_blocked = ?", true)
} else if query.Status == "active" {
db = db.Where("is_blocked = ?", false)
}

var total int64
if err := db.Count(&total).Error; err != nil {
c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count customers"})
return
}

switch query.Sort {
case "oldest":
db = db.Order("created_at ASC")
default:
db = db.Order("created_at DESC")
}

offset := (query.Page - 1) * query.Limit
var users []models.User
if err := db.Offset(offset).Limit(query.Limit).Find(&users).Error; err != nil {
c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch customers"})
return
}

// Aggregate order stats for every customer on this page in a single
// GROUP BY query instead of firing 3 queries per customer (Defect #09:
// N+1 that degraded badly as the customer table grew). TotalSpent
// excludes cancelled/returned orders - a refunded or cancelled order
// was never actually kept revenue, so counting it inflated the
// Lifetime Value / Total Spent figure shown on the admin dashboard.
userIDs := make([]uint, len(users))
for i, u := range users {
userIDs[i] = u.ID
}

type orderAgg struct {
UserID      uint
TotalOrders int64
TotalSpent  float64
LastOrderAt *time.Time
}
var aggs []orderAgg
if len(userIDs) > 0 {
database.DB.Model(&models.Order{}).
Select("user_id, COUNT(*) AS total_orders, "+
"COALESCE(SUM(CASE WHEN payment_status = 'paid' AND status NOT IN ('cancelled','returned') THEN total_amount ELSE 0 END), 0) AS total_spent, "+
"MAX(created_at) AS last_order_at").
Where("user_id IN ?", userIDs).
Group("user_id").
Scan(&aggs)
}
aggByUserID := make(map[uint]orderAgg, len(aggs))
for _, a := range aggs {
aggByUserID[a.UserID] = a
}

summaries := make([]models.CustomerSummary, 0, len(users))
for _, u := range users {
agg := aggByUserID[u.ID]
summary := models.CustomerSummary{
ID:          u.ID,
Name:        u.Name,
Phone:       u.Phone,
IsBlocked:   u.IsBlocked,
CreatedAt:   u.CreatedAt,
TotalOrders: agg.TotalOrders,
TotalSpent:  agg.TotalSpent,
}
if agg.LastOrderAt != nil {
summary.LastOrderAt = agg.LastOrderAt
}
summaries = append(summaries, summary)
}

c.JSON(http.StatusOK, models.CustomerListResponse{
Customers:  summaries,
Page:       query.Page,
Limit:      query.Limit,
Total:      total,
TotalPages: int((total + int64(query.Limit) - 1) / int64(query.Limit)),
})
}

// GetCustomerByID godoc
// GET /api/v1/admin/customers/:id (admin only)
func GetCustomerByID(c *gin.Context) {
id := c.Param("id")

var user models.User
if err := database.DB.Where("role = ?", "customer").First(&user, id).Error; err != nil {
c.JSON(http.StatusNotFound, gin.H{"error": "Customer not found"})
return
}

var orders []models.Order
database.DB.Preload("Items.Product").Preload("Address").Where("user_id = ?", user.ID).Order("created_at DESC").Find(&orders)

var addresses []models.Address
database.DB.Where("user_id = ?", user.ID).Find(&addresses)

var wallet models.Wallet
var walletPtr *models.Wallet
if err := database.DB.Where("user_id = ?", user.ID).First(&wallet).Error; err == nil {
walletPtr = &wallet
}

var transactions []models.WalletTransaction
if walletPtr != nil {
database.DB.Where("wallet_id = ?", walletPtr.ID).Order("created_at DESC").Limit(50).Find(&transactions)
}

// Computed from the orders slice already loaded above instead of a
// separate query (Defect #09) - also excludes cancelled/returned orders
// from TotalSpent, since a cancelled or returned order was never
// actually kept revenue and shouldn't inflate this customer's Total
// Spent / Lifetime Value figure.
var totalOrders int64 = int64(len(orders))
var totalSpent float64
for _, o := range orders {
if o.PaymentStatus == models.OrderPaymentStatusPaid && o.Status != models.OrderStatusCancelled && o.Status != models.OrderStatusReturned {
totalSpent += o.TotalAmount
}
}

c.JSON(http.StatusOK, models.CustomerDetail{
ID:           user.ID,
Name:         user.Name,
Phone:        user.Phone,
IsBlocked:    user.IsBlocked,
CreatedAt:    user.CreatedAt,
TotalOrders:  totalOrders,
TotalSpent:   totalSpent,
Orders:       orders,
Addresses:    addresses,
Wallet:       walletPtr,
Transactions: transactions,
})
}

// BlockCustomer godoc
// PUT /api/v1/admin/customers/:id/block (admin only)
func BlockCustomer(c *gin.Context) {
id := c.Param("id")
var user models.User
if err := database.DB.Where("role = ?", "customer").First(&user, id).Error; err != nil {
c.JSON(http.StatusNotFound, gin.H{"error": "Customer not found"})
return
}
user.IsBlocked = true
if err := database.DB.Save(&user).Error; err != nil {
c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to block customer"})
return
}
// Invalidate the AuthMiddleware block-status cache so this takes effect
// on the user's very next request, not after the cache TTL expires
// (Bug#20).
_ = cache.Delete(c.Request.Context(), fmt.Sprintf("auth:blocked:%d", user.ID))
adminID := c.MustGet("user_id").(uint)
adminPhone := c.MustGet("phone").(string)
utils.LogAudit(adminID, adminPhone, "block_customer", "customer", id, "blocked")
c.JSON(http.StatusOK, gin.H{"success": true, "is_blocked": true})
}

// UnblockCustomer godoc
// PUT /api/v1/admin/customers/:id/unblock (admin only)
func UnblockCustomer(c *gin.Context) {
id := c.Param("id")
var user models.User
if err := database.DB.Where("role = ?", "customer").First(&user, id).Error; err != nil {
c.JSON(http.StatusNotFound, gin.H{"error": "Customer not found"})
return
}
user.IsBlocked = false
if err := database.DB.Save(&user).Error; err != nil {
c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to unblock customer"})
return
}
_ = cache.Delete(c.Request.Context(), fmt.Sprintf("auth:blocked:%d", user.ID))
adminID := c.MustGet("user_id").(uint)
adminPhone := c.MustGet("phone").(string)
utils.LogAudit(adminID, adminPhone, "unblock_customer", "customer", id, "unblocked")
c.JSON(http.StatusOK, gin.H{"success": true, "is_blocked": false})
}
