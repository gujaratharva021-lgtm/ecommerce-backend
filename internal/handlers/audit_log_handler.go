package handlers

import (
"net/http"
"strconv"

"github.com/gin-gonic/gin"
"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/database"
"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/models"
)

// GetAuditLogs godoc
// GET /api/v1/admin/audit-logs?page=&limit=&action=&entity_type= (admin only)
func GetAuditLogs(c *gin.Context) {
page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
if page < 1 {
page = 1
}
if limit < 1 || limit > 200 {
limit = 50
}

query := database.DB.Model(&models.AuditLog{})

if action := c.Query("action"); action != "" {
query = query.Where("action = ?", action)
}
if entityType := c.Query("entity_type"); entityType != "" {
query = query.Where("entity_type = ?", entityType)
}

var total int64
query.Count(&total)

var logs []models.AuditLog
offset := (page - 1) * limit
if err := query.Order("created_at desc").Offset(offset).Limit(limit).Find(&logs).Error; err != nil {
c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch audit logs"})
return
}

c.JSON(http.StatusOK, gin.H{
"logs":       logs,
"page":       page,
"limit":      limit,
"total":      total,
"total_pages": (total + int64(limit) - 1) / int64(limit),
})
}
