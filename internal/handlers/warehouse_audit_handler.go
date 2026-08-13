package handlers

import (
"net/http"
"strconv"

"github.com/gin-gonic/gin"
"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/database"
"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/models"
)

// GetWarehouseAuditLogs godoc
// GET /api/v1/warehouse/audit-logs?action=&entity_type=&staff_id=&page=&limit= (warehouse staff only)
func GetWarehouseAuditLogs(c *gin.Context) {
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

db := database.DB.Model(&models.WarehouseAuditLog{}).Where("warehouse_id = ?", warehouseID)
if action := c.Query("action"); action != "" {
db = db.Where("action = ?", action)
}
if entityType := c.Query("entity_type"); entityType != "" {
db = db.Where("entity_type = ?", entityType)
}
if staffID := c.Query("staff_id"); staffID != "" {
db = db.Where("staff_id = ?", staffID)
}

var total int64
db.Count(&total)

var logs []models.WarehouseAuditLog
offset := (page - 1) * limit
if err := db.Order("created_at DESC").Offset(offset).Limit(limit).Find(&logs).Error; err != nil {
c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch audit logs"})
return
}

c.JSON(http.StatusOK, gin.H{
"audit_logs":  logs,
"page":        page,
"limit":       limit,
"total":       total,
"total_pages": int((total + int64(limit) - 1) / int64(limit)),
})
}
