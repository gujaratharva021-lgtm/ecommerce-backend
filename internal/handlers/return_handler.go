package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/database"
	"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/models"
	"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/utils"
	"gorm.io/gorm"
)

// ---------------------------------------------------------------------------
// Returns - Customer side
// ---------------------------------------------------------------------------

// RequestReturn godoc
// POST /api/v1/orders/:id/return (protected)
// Only delivered orders can be returned, and only once.
func RequestReturn(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)
	orderID := c.Param("id")

	var order models.Order
	if err := database.DB.First(&order, orderID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
		return
	}
	if order.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "You do not have access to this order"})
		return
	}
	if order.Status == models.OrderStatusReturned {
		c.JSON(http.StatusBadRequest, gin.H{"error": "This order has already been returned"})
		return
	}
	if order.Status == models.OrderStatusCancelled {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cancelled orders cannot be returned"})
		return
	}
	if order.Status != models.OrderStatusDelivered {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Only delivered orders can be returned"})
		return
	}

	var existing models.ReturnRequest
	if err := database.DB.Where("order_id = ? AND status IN ?", order.ID, []string{models.ReturnStatusPending, models.ReturnStatusApproved}).First(&existing).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "A return request already exists for this order"})
		return
	}

	var req models.ReturnRequestBody
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	returnReq := models.ReturnRequest{
		OrderID:      order.ID,
		UserID:       userID,
		Reason:       req.Reason,
		Status:       models.ReturnStatusPending,
		RefundAmount: order.TotalAmount,
	}

	if err := database.DB.Create(&returnReq).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create return request"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"return_request": returnReq})
}

// GetMyReturns godoc
// GET /api/v1/returns (protected)
func GetMyReturns(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)

	var returns []models.ReturnRequest
	if err := database.DB.Preload("Order").
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&returns).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load return requests"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"return_requests": returns})
}

// ---------------------------------------------------------------------------
// Returns - Admin side
// ---------------------------------------------------------------------------

// GetReturns godoc
// GET /api/v1/admin/returns (admin only) — optional ?status=
func GetReturns(c *gin.Context) {
	var returns []models.ReturnRequest
	query := database.DB.Preload("Order").Order("created_at DESC")

	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Find(&returns).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load return requests"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"return_requests": returns})
}

// ApproveReturn godoc
// PUT /api/v1/admin/returns/:id/approve (admin only)
// Restores stock for every item, refunds the order total to the customer's
// wallet, marks the order "returned", and closes out the return request.
func ApproveReturn(c *gin.Context) {
	adminID := c.MustGet("user_id").(uint)
	id := c.Param("id")

	var returnReq models.ReturnRequest
	if err := database.DB.First(&returnReq, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Return request not found"})
		return
	}
	if returnReq.Status != models.ReturnStatusPending {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Only pending return requests can be approved"})
		return
	}

	var order models.Order
	if err := database.DB.Preload("Items").First(&order, returnReq.OrderID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
		return
	}

	txErr := database.DB.Transaction(func(tx *gorm.DB) error {
		for _, item := range order.Items {
			var inventory models.Inventory
			if err := tx.Where("product_id = ?", item.ProductID).Order("id").First(&inventory).Error; err == nil {
				inventory.Stock += item.Quantity
				inventory.InStock = true
				if err := tx.Save(&inventory).Error; err != nil {
					return err
				}
			}
		}

		refID := order.ID
		if err := utils.CreditWallet(tx, returnReq.UserID, returnReq.RefundAmount, models.WalletReasonRefund, "return_request", &refID, "Refund for returned order #"+id); err != nil {
			return err
		}

		if err := tx.Model(&order).Update("status", models.OrderStatusReturned).Error; err != nil {
			return err
		}

		returnReq.Status = models.ReturnStatusApproved
		returnReq.ProcessedBy = &adminID
		return tx.Save(&returnReq).Error
	})

	if txErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to approve return: " + txErr.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"return_request": returnReq})
}

// RejectReturn godoc
// PUT /api/v1/admin/returns/:id/reject (admin only)
func RejectReturn(c *gin.Context) {
	adminID := c.MustGet("user_id").(uint)
	id := c.Param("id")

	var returnReq models.ReturnRequest
	if err := database.DB.First(&returnReq, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Return request not found"})
		return
	}
	if returnReq.Status != models.ReturnStatusPending {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Only pending return requests can be rejected"})
		return
	}

	returnReq.Status = models.ReturnStatusRejected
	returnReq.ProcessedBy = &adminID
	if err := database.DB.Save(&returnReq).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reject return request"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"return_request": returnReq})
}
