package handlers

import (
"fmt"
"net/http"
"strconv"
"time"

"github.com/gin-gonic/gin"
"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/database"
"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/models"
"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/services"
)

// GetWarehouseExceptions godoc
// GET /api/v1/warehouse/exceptions?status=&priority=&type=&order_id=&page=&limit= (warehouse staff only)
func GetWarehouseExceptions(c *gin.Context) {
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

db := database.DB.Model(&models.WarehouseException{}).Where("warehouse_id = ?", warehouseID)
if status := c.Query("status"); status != "" {
db = db.Where("status = ?", status)
}
if priority := c.Query("priority"); priority != "" {
db = db.Where("priority = ?", priority)
}
if excType := c.Query("type"); excType != "" {
db = db.Where("type = ?", excType)
}
if orderID := c.Query("order_id"); orderID != "" {
db = db.Where("order_id = ?", orderID)
}

var total int64
db.Count(&total)

var exceptions []models.WarehouseException
offset := (page - 1) * limit
if err := db.Preload("Order").Preload("Product").
Order("CASE priority WHEN 'high' THEN 1 WHEN 'medium' THEN 2 ELSE 3 END, created_at DESC").
Offset(offset).Limit(limit).Find(&exceptions).Error; err != nil {
c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch exceptions"})
return
}

c.JSON(http.StatusOK, gin.H{
"exceptions":  exceptions,
"page":        page,
"limit":       limit,
"total":       total,
"total_pages": int((total + int64(limit) - 1) / int64(limit)),
})
}

// GetWarehouseException godoc
// GET /api/v1/warehouse/exceptions/:id (warehouse staff only)
func GetWarehouseException(c *gin.Context) {
warehouseID := c.MustGet("warehouse_id").(uint)
id := c.Param("id")

var exception models.WarehouseException
if err := database.DB.Where("id = ? AND warehouse_id = ?", id, warehouseID).
Preload("Order").Preload("Product").First(&exception).Error; err != nil {
c.JSON(http.StatusNotFound, gin.H{"error": "Exception not found for your warehouse"})
return
}
c.JSON(http.StatusOK, exception)
}

// UpdateWarehouseException godoc
// PUT /api/v1/warehouse/exceptions/:id (warehouse staff only)
// Moves an exception through open -> investigating -> resolved/closed.
// resolved/closed require a resolution note and stamp resolved_by/resolved_at.
func UpdateWarehouseException(c *gin.Context) {
warehouseID := c.MustGet("warehouse_id").(uint)
staffID := c.MustGet("staff_id").(uint)
id := c.Param("id")

var req models.UpdateExceptionRequest
if err := c.ShouldBindJSON(&req); err != nil {
c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
return
}

var exception models.WarehouseException
if err := database.DB.Where("id = ? AND warehouse_id = ?", id, warehouseID).First(&exception).Error; err != nil {
c.JSON(http.StatusNotFound, gin.H{"error": "Exception not found for your warehouse"})
return
}

if (req.Status == models.ExceptionStatusResolved || req.Status == models.ExceptionStatusClosed) && req.Resolution == "" {
c.JSON(http.StatusBadRequest, gin.H{"error": "resolution is required when marking an exception resolved or closed"})
return
}

previousStatus := exception.Status
exception.Status = req.Status
if req.Resolution != "" {
exception.Resolution = req.Resolution
}
if req.Status == models.ExceptionStatusResolved || req.Status == models.ExceptionStatusClosed {
now := time.Now()
staffIDCopy := staffID
exception.ResolvedByID = &staffIDCopy
exception.ResolvedAt = &now
}

if err := database.DB.Save(&exception).Error; err != nil {
c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update exception"})
return
}

staffName, _ := c.Get("staff_name")
services.LogWarehouseAction(warehouseID, staffID, fmt.Sprint(staffName), "update_exception", "warehouse_exception",
strconv.Itoa(int(exception.ID)), "status="+previousStatus, "status="+exception.Status)

c.JSON(http.StatusOK, exception)
}
