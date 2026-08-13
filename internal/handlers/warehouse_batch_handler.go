package handlers

import (
"errors"
"fmt"
"net/http"
"strconv"
"time"

"github.com/gin-gonic/gin"
"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/database"
"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/models"
"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/services"
"gorm.io/gorm"
"gorm.io/gorm/clause"
)

// CreateBatch godoc
// POST /api/v1/warehouse/batches (warehouse staff only)
func CreateBatch(c *gin.Context) {
warehouseID := c.MustGet("warehouse_id").(uint)
staffID := c.MustGet("staff_id").(uint)

var req models.CreateBatchRequest
if err := c.ShouldBindJSON(&req); err != nil {
c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
return
}

var product models.Product
if err := database.DB.First(&product, req.ProductID).Error; err != nil {
c.JSON(http.StatusBadRequest, gin.H{"error": "Product not found"})
return
}

if !req.ExpiryDate.After(time.Now()) {
c.JSON(http.StatusBadRequest, gin.H{"error": "expiry_date must be in the future"})
return
}

batch := models.Batch{
ProductID:        req.ProductID,
WarehouseID:      warehouseID,
BatchNumber:      req.BatchNumber,
ManufactureDate:  req.ManufactureDate,
ExpiryDate:       req.ExpiryDate,
Quantity:         req.Quantity,
BinID:            req.BinID,
CreatedByStaffID: staffID,
}
if err := database.DB.Create(&batch).Error; err != nil {
c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create batch"})
return
}

staffName, _ := c.Get("staff_name")
services.LogWarehouseAction(warehouseID, staffID, fmt.Sprint(staffName), "batch_created", "batch", strconv.Itoa(int(batch.ID)),
"", fmt.Sprintf("product=%s qty=%d expiry=%s", product.Name, req.Quantity, req.ExpiryDate.Format("2006-01-02")))

c.JSON(http.StatusCreated, batch)
}

// GetWarehouseBatches godoc
// GET /api/v1/warehouse/batches?product_id=&expiring_within_days=&page=&limit= (warehouse staff only)
// Default order is FEFO - earliest expiry first - so this listing doubles
// as the pick-priority reference for batch-tracked products.
func GetWarehouseBatches(c *gin.Context) {
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

db := database.DB.Model(&models.Batch{}).Where("warehouse_id = ? AND quantity > 0", warehouseID)
if productID := c.Query("product_id"); productID != "" {
db = db.Where("product_id = ?", productID)
}
if days := c.Query("expiring_within_days"); days != "" {
if v, err := strconv.Atoi(days); err == nil {
cutoff := time.Now().AddDate(0, 0, v)
db = db.Where("expiry_date <= ?", cutoff)
}
}

var total int64
db.Count(&total)

var batches []models.Batch
offset := (page - 1) * limit
if err := db.Preload("Product").Preload("Bin.Rack.Zone").
Order("expiry_date ASC").Offset(offset).Limit(limit).Find(&batches).Error; err != nil {
c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch batches"})
return
}

c.JSON(http.StatusOK, gin.H{
"batches":     batches,
"page":        page,
"limit":       limit,
"total":       total,
"total_pages": int((total + int64(limit) - 1) / int64(limit)),
})
}

// GetExpiringBatches godoc
// GET /api/v1/warehouse/batches/expiring?days=7 (warehouse staff only)
// Convenience endpoint for dashboard/notification alerts - batches with
// quantity remaining that expire within the given window (default 7 days).
func GetExpiringBatches(c *gin.Context) {
warehouseID := c.MustGet("warehouse_id").(uint)

days := 7
if d := c.Query("days"); d != "" {
if v, err := strconv.Atoi(d); err == nil && v > 0 {
days = v
}
}
cutoff := time.Now().AddDate(0, 0, days)

var batches []models.Batch
if err := database.DB.Where("warehouse_id = ? AND quantity > 0 AND expiry_date <= ?", warehouseID, cutoff).
Preload("Product").Order("expiry_date ASC").Find(&batches).Error; err != nil {
c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch expiring batches"})
return
}

c.JSON(http.StatusOK, gin.H{"batches": batches, "days": days, "count": len(batches)})
}

// AdjustBatchQuantity godoc
// PUT /api/v1/warehouse/batches/:id/quantity (warehouse staff only)
// Corrects or consumes a batch's remaining quantity (e.g. after a FEFO
// pick, or a recount). Writes an audit log entry - does not touch the
// product's aggregate Inventory.Stock, which is adjusted separately via
// the normal stock-adjustment/pick flow.
func AdjustBatchQuantity(c *gin.Context) {
warehouseID := c.MustGet("warehouse_id").(uint)
staffID := c.MustGet("staff_id").(uint)
id := c.Param("id")

var req models.AdjustBatchQuantityRequest
if err := c.ShouldBindJSON(&req); err != nil {
c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
return
}

var batch models.Batch
var previousQty int
statusCode := http.StatusInternalServerError
txErr := database.DB.Transaction(func(tx *gorm.DB) error {
if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
Where("id = ? AND warehouse_id = ?", id, warehouseID).First(&batch).Error; err != nil {
statusCode = http.StatusNotFound
return errors.New("batch not found for your warehouse")
}
previousQty = batch.Quantity
batch.Quantity = req.Quantity
return tx.Save(&batch).Error
})
if txErr != nil {
c.JSON(statusCode, gin.H{"error": txErr.Error()})
return
}

staffName, _ := c.Get("staff_name")
services.LogWarehouseAction(warehouseID, staffID, fmt.Sprint(staffName), "batch_quantity_adjusted", "batch", id,
fmt.Sprintf("qty=%d", previousQty), fmt.Sprintf("qty=%d reason=%s", req.Quantity, req.Reason))

c.JSON(http.StatusOK, batch)
}

// DeleteBatch godoc
// DELETE /api/v1/warehouse/batches/:id (warehouse staff only)
// Only allowed once quantity has reached zero (fully consumed/expired out).
func DeleteBatch(c *gin.Context) {
warehouseID := c.MustGet("warehouse_id").(uint)
staffID := c.MustGet("staff_id").(uint)
id := c.Param("id")

var batch models.Batch
if err := database.DB.Where("id = ? AND warehouse_id = ?", id, warehouseID).First(&batch).Error; err != nil {
c.JSON(http.StatusNotFound, gin.H{"error": "Batch not found for your warehouse"})
return
}
if batch.Quantity > 0 {
c.JSON(http.StatusConflict, gin.H{"error": "Cannot delete a batch that still has quantity remaining - adjust it to 0 first"})
return
}

database.DB.Delete(&batch)

staffName, _ := c.Get("staff_name")
services.LogWarehouseAction(warehouseID, staffID, fmt.Sprint(staffName), "batch_deleted", "batch", id, batch.BatchNumber, "")

c.JSON(http.StatusOK, gin.H{"success": true})
}
