package handlers

import (
"net/http"

"github.com/gin-gonic/gin"
"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/database"
"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/models"
)

// ---------------------------------------------------------------------------
// Warehouse Staff (admin only)
// ---------------------------------------------------------------------------

// CreateWarehouseStaff godoc
// POST /api/v1/admin/warehouse-staff (admin only)
func CreateWarehouseStaff(c *gin.Context) {
var req models.WarehouseStaffRequest
if err := c.ShouldBindJSON(&req); err != nil {
c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
return
}

var warehouse models.Warehouse
if err := database.DB.First(&warehouse, req.WarehouseID).Error; err != nil {
c.JSON(http.StatusNotFound, gin.H{"error": "Warehouse not found"})
return
}

role := req.Role
if role == "" {
role = "picker"
}
if !models.ValidWarehouseStaffRoles[role] {
c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid role"})
return
}

staff := models.WarehouseStaff{
Name:        req.Name,
Phone:       req.Phone,
WarehouseID: req.WarehouseID,
Role:        role,
IsActive:    true,
}
if req.IsActive != nil {
staff.IsActive = *req.IsActive
}

if err := database.DB.Create(&staff).Error; err != nil {
c.JSON(http.StatusConflict, gin.H{"error": "Warehouse staff already exists or could not be created"})
return
}

c.JSON(http.StatusCreated, gin.H{"warehouse_staff": staff})
}

// GetWarehouseStaff godoc
// GET /api/v1/admin/warehouse-staff (admin only)
// Optional ?warehouse_id= query param to filter by warehouse.
func GetWarehouseStaff(c *gin.Context) {
var staff []models.WarehouseStaff
query := database.DB.Preload("Warehouse").Order("created_at DESC")

if warehouseID := c.Query("warehouse_id"); warehouseID != "" {
query = query.Where("warehouse_id = ?", warehouseID)
}

if err := query.Find(&staff).Error; err != nil {
c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load warehouse staff"})
return
}
c.JSON(http.StatusOK, gin.H{"warehouse_staff": staff})
}

// UpdateWarehouseStaff godoc
// PUT /api/v1/admin/warehouse-staff/:id (admin only)
func UpdateWarehouseStaff(c *gin.Context) {
id := c.Param("id")

var staff models.WarehouseStaff
if err := database.DB.First(&staff, id).Error; err != nil {
c.JSON(http.StatusNotFound, gin.H{"error": "Warehouse staff not found"})
return
}

var req models.WarehouseStaffRequest
if err := c.ShouldBindJSON(&req); err != nil {
c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
return
}

var warehouse models.Warehouse
if err := database.DB.First(&warehouse, req.WarehouseID).Error; err != nil {
c.JSON(http.StatusNotFound, gin.H{"error": "Warehouse not found"})
return
}

role := req.Role
if role == "" {
role = staff.Role
}
if !models.ValidWarehouseStaffRoles[role] {
c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid role"})
return
}

staff.Name = req.Name
staff.Phone = req.Phone
staff.WarehouseID = req.WarehouseID
staff.Role = role
if req.IsActive != nil {
staff.IsActive = *req.IsActive
}

if err := database.DB.Save(&staff).Error; err != nil {
c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update warehouse staff"})
return
}

c.JSON(http.StatusOK, gin.H{"warehouse_staff": staff})
}

// DeleteWarehouseStaff godoc
// DELETE /api/v1/admin/warehouse-staff/:id (admin only)
func DeleteWarehouseStaff(c *gin.Context) {
id := c.Param("id")

result := database.DB.Delete(&models.WarehouseStaff{}, id)
if result.Error != nil {
c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete warehouse staff"})
return
}
if result.RowsAffected == 0 {
c.JSON(http.StatusNotFound, gin.H{"error": "Warehouse staff not found"})
return
}

c.JSON(http.StatusOK, gin.H{"message": "Warehouse staff deleted"})
}
