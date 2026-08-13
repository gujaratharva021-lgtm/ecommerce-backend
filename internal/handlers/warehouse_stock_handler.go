package handlers

import (
"errors"
"fmt"
"net/http"
"strconv"

"github.com/gin-gonic/gin"
"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/database"
"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/models"
"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/services"
"gorm.io/gorm"
"gorm.io/gorm/clause"
)

// AdjustStock godoc
// POST /api/v1/warehouse/inventory/:product_id/adjust (warehouse staff only)
// Sets a product's stock to an absolute new value at the caller's warehouse,
// requiring a reason, and writes a StockMovement audit record. This is the
// ONLY sanctioned way for warehouse staff to change stock outside of the
// order/picking flow - never allow raw stock writes from the frontend.
func AdjustStock(c *gin.Context) {
warehouseID := c.MustGet("warehouse_id").(uint)
staffID := c.MustGet("staff_id").(uint)
staffName, _ := c.Get("staff_name")
productID := c.Param("product_id")

var req models.StockAdjustmentRequest
if err := c.ShouldBindJSON(&req); err != nil {
c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
return
}

productIDUint64, _ := strconv.ParseUint(productID, 10, 64)
staffIDCopy := staffID

var inv models.Inventory
var previousQty int
statusCode := http.StatusInternalServerError

txErr := database.DB.Transaction(func(tx *gorm.DB) error {
if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
Where("product_id = ? AND warehouse_id = ?", productID, warehouseID).
First(&inv).Error; err != nil {
statusCode = http.StatusNotFound
return errors.New("No inventory record for this product in your warehouse")
}

previousQty = inv.Stock
change := req.NewQuantity - previousQty
if change == 0 {
statusCode = http.StatusBadRequest
return errors.New("New quantity is the same as current stock - nothing to adjust")
}

inv.Stock = req.NewQuantity
inv.InStock = req.NewQuantity > 0
if err := tx.Save(&inv).Error; err != nil {
return err
}
movement := models.StockMovement{
ProductID:    uint(productIDUint64),
WarehouseID:  warehouseID,
PreviousQty:  previousQty,
Change:       change,
NewQty:       req.NewQuantity,
MovementType: models.MovementAdjustment,
Reason:       req.Reason,
StaffID:      &staffIDCopy,
Notes:        req.Notes,
}
return tx.Create(&movement).Error
})

if txErr != nil {
c.JSON(statusCode, gin.H{"error": txErr.Error()})
return
}

services.LogWarehouseAction(warehouseID, staffID, fmt.Sprint(staffName), "stock_adjustment", "inventory", productID,
fmt.Sprintf("stock=%d", previousQty), fmt.Sprintf("stock=%d reason=%s", req.NewQuantity, req.Reason))

if inv.Stock <= 0 {
services.NotifyWarehouse(warehouseID, models.WhNotifyOutOfStock,
"Product out of stock", fmt.Sprintf("Product #%d is now out of stock at your warehouse.", productIDUint64),
nil, ptrUint(uint(productIDUint64)))
} else if inv.Stock < lowStockThreshold {
services.NotifyWarehouse(warehouseID, models.WhNotifyLowStock,
"Low stock warning", fmt.Sprintf("Product #%d is running low (%d left).", productIDUint64, inv.Stock),
nil, ptrUint(uint(productIDUint64)))
}

c.JSON(http.StatusOK, inv)
}

func ptrUint(v uint) *uint { return &v }

// GetStockMovements godoc
// GET /api/v1/warehouse/stock-movements?product_id=&movement_type=&page=&limit= (warehouse staff only)
func GetStockMovements(c *gin.Context) {
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

db := database.DB.Model(&models.StockMovement{}).Where("warehouse_id = ?", warehouseID)
if productID := c.Query("product_id"); productID != "" {
db = db.Where("product_id = ?", productID)
}
if movementType := c.Query("movement_type"); movementType != "" {
db = db.Where("movement_type = ?", movementType)
}

var total int64
db.Count(&total)

var movements []models.StockMovement
offset := (page - 1) * limit
if err := db.Preload("Product").Order("created_at DESC").Offset(offset).Limit(limit).Find(&movements).Error; err != nil {
c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch stock movements"})
return
}

c.JSON(http.StatusOK, gin.H{
"movements":   movements,
"page":        page,
"limit":       limit,
"total":       total,
"total_pages": int((total + int64(limit) - 1) / int64(limit)),
})
}
