package handlers

import (
"errors"
"net/http"

"github.com/gin-gonic/gin"
"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/database"
"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/models"
"gorm.io/gorm"
"gorm.io/gorm/clause"
)

// ---------------------------------------------------------------------------
// Stock Transfers - Warehouse Staff side
// ---------------------------------------------------------------------------

// RequestStockTransfer godoc
// POST /api/v1/warehouse/stock-transfers (warehouse staff only)
// The requesting staff member's own warehouse is always the source.
func RequestStockTransfer(c *gin.Context) {
staffID := c.MustGet("user_id").(uint)

var staff models.WarehouseStaff
if err := database.DB.First(&staff, staffID).Error; err != nil {
c.JSON(http.StatusNotFound, gin.H{"error": "Warehouse staff not found"})
return
}

var req models.StockTransferRequest
if err := c.ShouldBindJSON(&req); err != nil {
c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
return
}

if req.ToWarehouseID == staff.WarehouseID {
c.JSON(http.StatusBadRequest, gin.H{"error": "Source and destination warehouse cannot be the same"})
return
}

var destWarehouse models.Warehouse
if err := database.DB.First(&destWarehouse, req.ToWarehouseID).Error; err != nil {
c.JSON(http.StatusNotFound, gin.H{"error": "Destination warehouse not found"})
return
}

var product models.Product
if err := database.DB.First(&product, req.ProductID).Error; err != nil {
c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
return
}

transfer := models.StockTransfer{
ProductID:       req.ProductID,
FromWarehouseID: staff.WarehouseID,
ToWarehouseID:   req.ToWarehouseID,
Quantity:        req.Quantity,
Status:          models.StockTransferPending,
RequestedBy:     staffID,
}

if err := database.DB.Create(&transfer).Error; err != nil {
c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create stock transfer request"})
return
}

c.JSON(http.StatusCreated, gin.H{"stock_transfer": transfer})
}

// GetMyStockTransfers godoc
// GET /api/v1/warehouse/stock-transfers (warehouse staff only)
// Returns transfers where the staff's warehouse is either the source or destination.
func GetMyStockTransfers(c *gin.Context) {
staffID := c.MustGet("user_id").(uint)

var staff models.WarehouseStaff
if err := database.DB.First(&staff, staffID).Error; err != nil {
c.JSON(http.StatusNotFound, gin.H{"error": "Warehouse staff not found"})
return
}

var transfers []models.StockTransfer
if err := database.DB.
Preload("Product").Preload("FromWarehouse").Preload("ToWarehouse").
Where("from_warehouse_id = ? OR to_warehouse_id = ?", staff.WarehouseID, staff.WarehouseID).
Order("created_at DESC").
Find(&transfers).Error; err != nil {
c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load stock transfers"})
return
}

c.JSON(http.StatusOK, gin.H{"stock_transfers": transfers})
}

// ReceiveStockTransfer godoc
// PUT /api/v1/warehouse/stock-transfers/:id/receive (warehouse staff only)
// Only staff at the DESTINATION warehouse may confirm receipt. Adds the
// quantity to the destination's inventory and marks the transfer received.
func ReceiveStockTransfer(c *gin.Context) {
staffID := c.MustGet("user_id").(uint)
id := c.Param("id")

var staff models.WarehouseStaff
if err := database.DB.First(&staff, staffID).Error; err != nil {
c.JSON(http.StatusNotFound, gin.H{"error": "Warehouse staff not found"})
return
}

var transfer models.StockTransfer
if err := database.DB.First(&transfer, id).Error; err != nil {
c.JSON(http.StatusNotFound, gin.H{"error": "Stock transfer not found"})
return
}

if transfer.ToWarehouseID != staff.WarehouseID {
c.JSON(http.StatusForbidden, gin.H{"error": "Only the destination warehouse can receive this transfer"})
return
}

if transfer.Status != models.StockTransferInTransit {
c.JSON(http.StatusBadRequest, gin.H{"error": "Only in-transit transfers can be received"})
return
}

txErr := database.DB.Transaction(func(tx *gorm.DB) error {
var inventory models.Inventory
err := tx.Where("product_id = ? AND warehouse_id = ?", transfer.ProductID, transfer.ToWarehouseID).
First(&inventory).Error
if err != nil {
inventory = models.Inventory{
ProductID:   transfer.ProductID,
WarehouseID: transfer.ToWarehouseID,
}
}
inventory.Stock += transfer.Quantity
inventory.InStock = inventory.Stock > 0
if err := tx.Save(&inventory).Error; err != nil {
return err
}

transfer.Status = models.StockTransferReceived
return tx.Save(&transfer).Error
})

if txErr != nil {
c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to receive stock transfer"})
return
}

c.JSON(http.StatusOK, gin.H{"stock_transfer": transfer})
}

// ApproveStockTransferByWarehouseStaff godoc
// PUT /api/v1/warehouse/stock-transfers/:id/approve (warehouse staff only)
// Only staff at the DESTINATION warehouse may approve an incoming transfer.
// Same effect as the admin ApproveStockTransfer: deducts stock from the
// source warehouse immediately and marks the transfer in_transit.
func ApproveStockTransferByWarehouseStaff(c *gin.Context) {
staffID := c.MustGet("user_id").(uint)
id := c.Param("id")

var staff models.WarehouseStaff
if err := database.DB.First(&staff, staffID).Error; err != nil {
c.JSON(http.StatusNotFound, gin.H{"error": "Warehouse staff not found"})
return
}

var transfer models.StockTransfer
if err := database.DB.First(&transfer, id).Error; err != nil {
c.JSON(http.StatusNotFound, gin.H{"error": "Stock transfer not found"})
return
}

if transfer.ToWarehouseID != staff.WarehouseID {
c.JSON(http.StatusForbidden, gin.H{"error": "Only the destination warehouse can approve this transfer"})
return
}

if transfer.Status != models.StockTransferPending {
c.JSON(http.StatusBadRequest, gin.H{"error": "Only pending transfers can be approved"})
return
}

txErr := database.DB.Transaction(func(tx *gorm.DB) error {
var inventory models.Inventory
if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
Where("product_id = ? AND warehouse_id = ?", transfer.ProductID, transfer.FromWarehouseID).
First(&inventory).Error; err != nil {
return errors.New("source warehouse has no inventory record for this product")
}

if inventory.Stock < transfer.Quantity {
return errors.New("insufficient stock at source warehouse to approve this transfer")
}

inventory.Stock -= transfer.Quantity
if inventory.Stock <= 0 {
inventory.InStock = false
}
if err := tx.Save(&inventory).Error; err != nil {
return err
}

transfer.Status = models.StockTransferInTransit
transfer.ApprovedBy = &staffID
return tx.Save(&transfer).Error
})

if txErr != nil {
c.JSON(http.StatusBadRequest, gin.H{"error": txErr.Error()})
return
}

c.JSON(http.StatusOK, gin.H{"stock_transfer": transfer})
}

// RejectStockTransferByWarehouseStaff godoc
// PUT /api/v1/warehouse/stock-transfers/:id/reject (warehouse staff only)
// Only staff at the DESTINATION warehouse may reject an incoming transfer.
func RejectStockTransferByWarehouseStaff(c *gin.Context) {
staffID := c.MustGet("user_id").(uint)
id := c.Param("id")

var staff models.WarehouseStaff
if err := database.DB.First(&staff, staffID).Error; err != nil {
c.JSON(http.StatusNotFound, gin.H{"error": "Warehouse staff not found"})
return
}

var transfer models.StockTransfer
if err := database.DB.First(&transfer, id).Error; err != nil {
c.JSON(http.StatusNotFound, gin.H{"error": "Stock transfer not found"})
return
}

if transfer.ToWarehouseID != staff.WarehouseID {
c.JSON(http.StatusForbidden, gin.H{"error": "Only the destination warehouse can reject this transfer"})
return
}

if transfer.Status != models.StockTransferPending {
c.JSON(http.StatusBadRequest, gin.H{"error": "Only pending transfers can be rejected"})
return
}

transfer.Status = models.StockTransferRejected
transfer.ApprovedBy = &staffID
if err := database.DB.Save(&transfer).Error; err != nil {
c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reject stock transfer"})
return
}

c.JSON(http.StatusOK, gin.H{"stock_transfer": transfer})
}

// ---------------------------------------------------------------------------
// Stock Transfers - Admin side
// ---------------------------------------------------------------------------

// GetStockTransfers godoc
// GET /api/v1/admin/stock-transfers (admin only)
// Optional ?status= query param to filter.
func GetStockTransfers(c *gin.Context) {
var transfers []models.StockTransfer
query := database.DB.
Preload("Product").Preload("FromWarehouse").Preload("ToWarehouse").
Order("created_at DESC")

if status := c.Query("status"); status != "" {
query = query.Where("status = ?", status)
}

if err := query.Find(&transfers).Error; err != nil {
c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load stock transfers"})
return
}

c.JSON(http.StatusOK, gin.H{"stock_transfers": transfers})
}

// ApproveStockTransfer godoc
// PUT /api/v1/admin/stock-transfers/:id/approve (admin only)
// Deducts stock from the source warehouse immediately and marks the
// transfer in_transit. The destination only gets the stock once it
// confirms receipt via ReceiveStockTransfer.
func ApproveStockTransfer(c *gin.Context) {
adminID := c.MustGet("user_id").(uint)
id := c.Param("id")

var transfer models.StockTransfer
if err := database.DB.First(&transfer, id).Error; err != nil {
c.JSON(http.StatusNotFound, gin.H{"error": "Stock transfer not found"})
return
}

if transfer.Status != models.StockTransferPending {
c.JSON(http.StatusBadRequest, gin.H{"error": "Only pending transfers can be approved"})
return
}

txErr := database.DB.Transaction(func(tx *gorm.DB) error {
var inventory models.Inventory
if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
Where("product_id = ? AND warehouse_id = ?", transfer.ProductID, transfer.FromWarehouseID).
First(&inventory).Error; err != nil {
return errors.New("source warehouse has no inventory record for this product")
}

if inventory.Stock < transfer.Quantity {
return errors.New("insufficient stock at source warehouse to approve this transfer")
}

inventory.Stock -= transfer.Quantity
if inventory.Stock <= 0 {
inventory.InStock = false
}
if err := tx.Save(&inventory).Error; err != nil {
return err
}

transfer.Status = models.StockTransferInTransit
transfer.ApprovedBy = &adminID
return tx.Save(&transfer).Error
})

if txErr != nil {
c.JSON(http.StatusBadRequest, gin.H{"error": txErr.Error()})
return
}

c.JSON(http.StatusOK, gin.H{"stock_transfer": transfer})
}

// RejectStockTransfer godoc
// PUT /api/v1/admin/stock-transfers/:id/reject (admin only)
func RejectStockTransfer(c *gin.Context) {
adminID := c.MustGet("user_id").(uint)
id := c.Param("id")

var transfer models.StockTransfer
if err := database.DB.First(&transfer, id).Error; err != nil {
c.JSON(http.StatusNotFound, gin.H{"error": "Stock transfer not found"})
return
}

if transfer.Status != models.StockTransferPending {
c.JSON(http.StatusBadRequest, gin.H{"error": "Only pending transfers can be rejected"})
return
}

transfer.Status = models.StockTransferRejected
transfer.ApprovedBy = &adminID
if err := database.DB.Save(&transfer).Error; err != nil {
c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reject stock transfer"})
return
}

c.JSON(http.StatusOK, gin.H{"stock_transfer": transfer})
}