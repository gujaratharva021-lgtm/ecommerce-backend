package handlers

import (
"net/http"
"strconv"
"time"

"github.com/gin-gonic/gin"
"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/database"
"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/models"
)

// GetWarehouseNotifications godoc
// GET /api/v1/warehouse/notifications?is_read=&type=&page=&limit= (warehouse staff only)
func GetWarehouseNotifications(c *gin.Context) {
warehouseID := c.MustGet("warehouse_id").(uint)

page := 1
limit := 20
if p := c.Query("page"); p != "" {
if v, err := strconv.Atoi(p); err == nil && v > 0 {
page = v
}
}
if l := c.Query("limit"); l != "" {
if v, err := strconv.Atoi(l); err == nil && v > 0 && v <= 100 {
limit = v
}
}

db := database.DB.Model(&models.WarehouseNotification{}).Where("warehouse_id = ?", warehouseID)
if isRead := c.Query("is_read"); isRead != "" {
db = db.Where("is_read = ?", isRead == "true")
}
if notifType := c.Query("type"); notifType != "" {
db = db.Where("type = ?", notifType)
}

var total, unread int64
db.Count(&total)
database.DB.Model(&models.WarehouseNotification{}).Where("warehouse_id = ? AND is_read = ?", warehouseID, false).Count(&unread)

var notifications []models.WarehouseNotification
offset := (page - 1) * limit
if err := db.Order("created_at DESC").Offset(offset).Limit(limit).Find(&notifications).Error; err != nil {
c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch notifications"})
return
}

c.JSON(http.StatusOK, gin.H{
"notifications": notifications,
"unread_count":  unread,
"page":          page,
"limit":         limit,
"total":         total,
"total_pages":   int((total + int64(limit) - 1) / int64(limit)),
})
}

// MarkNotificationRead godoc
// PUT /api/v1/warehouse/notifications/:id/read (warehouse staff only)
func MarkNotificationRead(c *gin.Context) {
warehouseID := c.MustGet("warehouse_id").(uint)
id := c.Param("id")

result := database.DB.Model(&models.WarehouseNotification{}).
Where("id = ? AND warehouse_id = ?", id, warehouseID).
Updates(map[string]interface{}{"is_read": true, "read_at": time.Now()})

if result.Error != nil {
c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update notification"})
return
}
if result.RowsAffected == 0 {
c.JSON(http.StatusNotFound, gin.H{"error": "Notification not found for your warehouse"})
return
}
c.JSON(http.StatusOK, gin.H{"success": true})
}

// MarkAllNotificationsRead godoc
// PUT /api/v1/warehouse/notifications/read-all (warehouse staff only)
func MarkAllNotificationsRead(c *gin.Context) {
warehouseID := c.MustGet("warehouse_id").(uint)

if err := database.DB.Model(&models.WarehouseNotification{}).
Where("warehouse_id = ? AND is_read = ?", warehouseID, false).
Updates(map[string]interface{}{"is_read": true, "read_at": time.Now()}).Error; err != nil {
c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update notifications"})
return
}
c.JSON(http.StatusOK, gin.H{"success": true})
}
