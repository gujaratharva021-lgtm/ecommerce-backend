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

// CreateSubstitutionRequest godoc
// POST /api/v1/warehouse/substitutions (warehouse/store staff - picker/packer)
// A picker/packer requests to swap the original product for a substitute
// when the original is unavailable or short during picking/packing.
func CreateSubstitutionRequest(c *gin.Context) {
	warehouseID := c.MustGet("warehouse_id").(uint)
	staffID := c.MustGet("staff_id").(uint)

	var req models.CreateSubstitutionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Make sure the order actually belongs to this warehouse.
	var order models.Order
	if err := database.DB.Where("id = ? AND warehouse_id = ?", req.OrderID, warehouseID).First(&order).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Order not found for your warehouse"})
		return
	}

	sub := models.SubstitutionRequest{
		OrderID:             req.OrderID,
		PickingTaskItemID:   req.PickingTaskItemID,
		OriginalProductID:   req.OriginalProductID,
		SubstituteProductID: req.SubstituteProductID,
		Quantity:            req.Quantity,
		Reason:              req.Reason,
		WarehouseID:         warehouseID,
		RequestedByID:       staffID,
		Status:              models.SubstitutionStatusPending,
	}

	if err := database.DB.Create(&sub).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create substitution request"})
		return
	}

	staffName, _ := c.Get("staff_name")
	services.LogWarehouseAction(warehouseID, staffID, fmt.Sprint(staffName), "create_substitution", "substitution_request",
		strconv.Itoa(int(sub.ID)), "", "status=pending")

	database.DB.Preload("Order").Preload("OriginalProduct").Preload("SubstituteProduct").First(&sub, sub.ID)
	c.JSON(http.StatusCreated, sub)
}

// GetSubstitutionRequests godoc
// GET /api/v1/warehouse/substitutions?status=&order_id=&page=&limit= (warehouse/store staff)
func GetSubstitutionRequests(c *gin.Context) {
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

	db := database.DB.Model(&models.SubstitutionRequest{}).Where("warehouse_id = ?", warehouseID)
	if status := c.Query("status"); status != "" {
		db = db.Where("status = ?", status)
	}
	if orderID := c.Query("order_id"); orderID != "" {
		db = db.Where("order_id = ?", orderID)
	}

	var total int64
	db.Count(&total)

	var subs []models.SubstitutionRequest
	offset := (page - 1) * limit
	if err := db.Preload("Order").Preload("OriginalProduct").Preload("SubstituteProduct").
		Order("created_at DESC").Offset(offset).Limit(limit).Find(&subs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch substitution requests"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"substitution_requests": subs,
		"page":                  page,
		"limit":                 limit,
		"total":                 total,
		"total_pages":           int((total + int64(limit) - 1) / int64(limit)),
	})
}

// GetSubstitutionRequest godoc
// GET /api/v1/warehouse/substitutions/:id (warehouse/store staff)
func GetSubstitutionRequest(c *gin.Context) {
	warehouseID := c.MustGet("warehouse_id").(uint)
	id := c.Param("id")

	var sub models.SubstitutionRequest
	if err := database.DB.Where("id = ? AND warehouse_id = ?", id, warehouseID).
		Preload("Order").Preload("OriginalProduct").Preload("SubstituteProduct").First(&sub).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Substitution request not found for your warehouse"})
		return
	}
	c.JSON(http.StatusOK, sub)
}

// ApproveSubstitutionRequest godoc
// PUT /api/v1/warehouse/substitutions/:id/approve (warehouse_manager only - enforce via route middleware)
func ApproveSubstitutionRequest(c *gin.Context) {
	decideSubstitution(c, models.SubstitutionStatusApproved, "approve_substitution")
}

// RejectSubstitutionRequest godoc
// PUT /api/v1/warehouse/substitutions/:id/reject (warehouse_manager only - enforce via route middleware)
func RejectSubstitutionRequest(c *gin.Context) {
	decideSubstitution(c, models.SubstitutionStatusRejected, "reject_substitution")
}

func decideSubstitution(c *gin.Context, newStatus string, auditAction string) {
	warehouseID := c.MustGet("warehouse_id").(uint)
	staffID := c.MustGet("staff_id").(uint)
	id := c.Param("id")

	var req models.DecideSubstitutionRequest
	// Note field is optional, so ignore a bind error on an empty body.
	_ = c.ShouldBindJSON(&req)

	var sub models.SubstitutionRequest
	if err := database.DB.Where("id = ? AND warehouse_id = ?", id, warehouseID).First(&sub).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Substitution request not found for your warehouse"})
		return
	}

	if sub.Status != models.SubstitutionStatusPending {
		c.JSON(http.StatusConflict, gin.H{"error": "Substitution request has already been decided"})
		return
	}

	previousStatus := sub.Status
	sub.Status = newStatus
	sub.DecisionNote = req.Note
	staffIDCopy := staffID
	sub.DecidedByID = &staffIDCopy
	now := time.Now()
	sub.DecidedAt = &now

	if err := database.DB.Save(&sub).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update substitution request"})
		return
	}

	staffName, _ := c.Get("staff_name")
	services.LogWarehouseAction(warehouseID, staffID, fmt.Sprint(staffName), auditAction, "substitution_request",
		strconv.Itoa(int(sub.ID)), "status="+previousStatus, "status="+sub.Status)

	database.DB.Preload("Order").Preload("OriginalProduct").Preload("SubstituteProduct").First(&sub, sub.ID)
	c.JSON(http.StatusOK, sub)
}
