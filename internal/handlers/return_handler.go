package handlers

import (
"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/database"
	"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/models"
"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/services"
	"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/utils"
	"gorm.io/gorm"
)

// ---------------------------------------------------------------------------
// Returns - Customer side
// ---------------------------------------------------------------------------

// RequestReturn godoc
// POST /api/v1/orders/:id/return (protected)
// Body: { reason, items: [{order_item_id, quantity}] }
// Only delivered orders can be returned. Each order item can be returned at
// most once in total across all pending/approved return requests (partial
// quantities allowed). Delivery charge is refunded only if this request,
// combined with any other pending/approved requests on the order, covers
// every unit of every item on the order.
func RequestReturn(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)
	orderID := c.Param("id")

	var order models.Order
	if err := database.DB.Preload("Items").First(&order, orderID).Error; err != nil {
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

	var req models.ReturnRequestBody
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Index order items by ID for quick lookup / validation.
	orderItemsByID := make(map[uint]models.OrderItem, len(order.Items))
	for _, it := range order.Items {
		orderItemsByID[it.ID] = it
	}

	// Sum of quantities already tied up in pending/approved return requests,
	// per order item, so we don't allow returning more than was bought.
	type qtyRow struct {
		OrderItemID uint
		Total       int
	}
	var existingRows []qtyRow
	if err := database.DB.Table("return_request_items").
		Select("order_item_id, SUM(quantity) as total").
		Joins("JOIN return_requests ON return_requests.id = return_request_items.return_request_id").
		Where("return_requests.order_id = ? AND return_requests.status IN ?", order.ID, []string{models.ReturnStatusPending, models.ReturnStatusApproved}).
		Group("order_item_id").
		Scan(&existingRows).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to validate return quantities"})
		return
	}
	alreadyReturned := make(map[uint]int, len(existingRows))
	for _, row := range existingRows {
		alreadyReturned[row.OrderItemID] = row.Total
	}

	if len(req.Items) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "At least one item is required"})
		return
	}

	var returnItems []models.ReturnRequestItem
	var itemsRefund float64
	seenInThisRequest := make(map[uint]bool)

	for _, reqItem := range req.Items {
		if seenInThisRequest[reqItem.OrderItemID] {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Duplicate order_item_id in request"})
			return
		}
		seenInThisRequest[reqItem.OrderItemID] = true

		orderItem, ok := orderItemsByID[reqItem.OrderItemID]
		if !ok {
			c.JSON(http.StatusBadRequest, gin.H{"error": "order_item_id does not belong to this order"})
			return
		}

		remaining := orderItem.Quantity - alreadyReturned[orderItem.ID]
		if reqItem.Quantity > remaining {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot return more units than remain for one or more items"})
			return
		}

		lineRefund := orderItem.Price * float64(reqItem.Quantity)
		itemsRefund += lineRefund

		returnItems = append(returnItems, models.ReturnRequestItem{
			OrderItemID:  orderItem.ID,
			Quantity:     reqItem.Quantity,
			RefundAmount: lineRefund,
		})
	}

	// Does this request, combined with existing pending/approved requests,
	// account for every unit of every item on the order? If so, refund the
	// delivery charge too.
	coversWholeOrder := true
	for _, it := range order.Items {
		covered := alreadyReturned[it.ID]
		for _, ri := range returnItems {
			if ri.OrderItemID == it.ID {
				covered += ri.Quantity
			}
		}
		if covered < it.Quantity {
			coversWholeOrder = false
			break
		}
	}

	refundAmount := itemsRefund
	if coversWholeOrder {
		refundAmount += order.DeliveryCharge
	}

	returnReq := models.ReturnRequest{
		OrderID:      order.ID,
		UserID:       userID,
		Reason:       req.Reason,
		Status:       models.ReturnStatusPending,
		RefundAmount: refundAmount,
		Items:        returnItems,
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
	if err := database.DB.Preload("Order").Preload("Items.OrderItem.Product").
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
// GET /api/v1/admin/returns (admin only) ÃƒÆ’Ã†â€™Ãƒâ€ Ã¢â‚¬â„¢ÃƒÆ’Ã¢â‚¬Â ÃƒÂ¢Ã¢â€šÂ¬Ã¢â€žÂ¢ÃƒÆ’Ã†â€™ÃƒÂ¢Ã¢â€šÂ¬Ã…Â¡ÃƒÆ’Ã¢â‚¬Å¡Ãƒâ€šÃ‚Â¢ÃƒÆ’Ã†â€™Ãƒâ€ Ã¢â‚¬â„¢ÃƒÆ’Ã¢â‚¬Å¡Ãƒâ€šÃ‚Â¢ÃƒÆ’Ã†â€™Ãƒâ€šÃ‚Â¢ÃƒÆ’Ã‚Â¢ÃƒÂ¢Ã¢â€šÂ¬Ã…Â¡Ãƒâ€šÃ‚Â¬ÃƒÆ’Ã¢â‚¬Â¦Ãƒâ€šÃ‚Â¡ÃƒÆ’Ã†â€™ÃƒÂ¢Ã¢â€šÂ¬Ã…Â¡ÃƒÆ’Ã¢â‚¬Å¡Ãƒâ€šÃ‚Â¬ÃƒÆ’Ã†â€™Ãƒâ€ Ã¢â‚¬â„¢ÃƒÆ’Ã¢â‚¬Å¡Ãƒâ€šÃ‚Â¢ÃƒÆ’Ã†â€™Ãƒâ€šÃ‚Â¢ÃƒÆ’Ã‚Â¢ÃƒÂ¢Ã¢â‚¬Å¡Ã‚Â¬Ãƒâ€¦Ã‚Â¡ÃƒÆ’Ã¢â‚¬Å¡Ãƒâ€šÃ‚Â¬ÃƒÆ’Ã†â€™ÃƒÂ¢Ã¢â€šÂ¬Ã…Â¡ÃƒÆ’Ã¢â‚¬Å¡Ãƒâ€šÃ‚Â optional ?status=
func GetReturns(c *gin.Context) {
	var returns []models.ReturnRequest
	query := database.DB.Preload("Order").Preload("Items.OrderItem.Product").Order("created_at DESC")

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
// Restores stock for each returned line item, refunds the pre-computed
// amount to the customer's wallet, and ÃƒÆ’Ã†â€™Ãƒâ€ Ã¢â‚¬â„¢ÃƒÆ’Ã¢â‚¬Â ÃƒÂ¢Ã¢â€šÂ¬Ã¢â€žÂ¢ÃƒÆ’Ã†â€™ÃƒÂ¢Ã¢â€šÂ¬Ã…Â¡ÃƒÆ’Ã¢â‚¬Å¡Ãƒâ€šÃ‚Â¢ÃƒÆ’Ã†â€™Ãƒâ€ Ã¢â‚¬â„¢ÃƒÆ’Ã¢â‚¬Å¡Ãƒâ€šÃ‚Â¢ÃƒÆ’Ã†â€™Ãƒâ€šÃ‚Â¢ÃƒÆ’Ã‚Â¢ÃƒÂ¢Ã¢â€šÂ¬Ã…Â¡Ãƒâ€šÃ‚Â¬ÃƒÆ’Ã¢â‚¬Â¦Ãƒâ€šÃ‚Â¡ÃƒÆ’Ã†â€™ÃƒÂ¢Ã¢â€šÂ¬Ã…Â¡ÃƒÆ’Ã¢â‚¬Å¡Ãƒâ€šÃ‚Â¬ÃƒÆ’Ã†â€™Ãƒâ€ Ã¢â‚¬â„¢ÃƒÆ’Ã¢â‚¬Å¡Ãƒâ€šÃ‚Â¢ÃƒÆ’Ã†â€™Ãƒâ€šÃ‚Â¢ÃƒÆ’Ã‚Â¢ÃƒÂ¢Ã¢â‚¬Å¡Ã‚Â¬Ãƒâ€¦Ã‚Â¡ÃƒÆ’Ã¢â‚¬Å¡Ãƒâ€šÃ‚Â¬ÃƒÆ’Ã†â€™ÃƒÂ¢Ã¢â€šÂ¬Ã…Â¡ÃƒÆ’Ã¢â‚¬Å¡Ãƒâ€šÃ‚Â only if every item on the order
// has now been fully returned across all approved requests ÃƒÆ’Ã†â€™Ãƒâ€ Ã¢â‚¬â„¢ÃƒÆ’Ã¢â‚¬Â ÃƒÂ¢Ã¢â€šÂ¬Ã¢â€žÂ¢ÃƒÆ’Ã†â€™ÃƒÂ¢Ã¢â€šÂ¬Ã…Â¡ÃƒÆ’Ã¢â‚¬Å¡Ãƒâ€šÃ‚Â¢ÃƒÆ’Ã†â€™Ãƒâ€ Ã¢â‚¬â„¢ÃƒÆ’Ã¢â‚¬Å¡Ãƒâ€šÃ‚Â¢ÃƒÆ’Ã†â€™Ãƒâ€šÃ‚Â¢ÃƒÆ’Ã‚Â¢ÃƒÂ¢Ã¢â€šÂ¬Ã…Â¡Ãƒâ€šÃ‚Â¬ÃƒÆ’Ã¢â‚¬Â¦Ãƒâ€šÃ‚Â¡ÃƒÆ’Ã†â€™ÃƒÂ¢Ã¢â€šÂ¬Ã…Â¡ÃƒÆ’Ã¢â‚¬Å¡Ãƒâ€šÃ‚Â¬ÃƒÆ’Ã†â€™Ãƒâ€ Ã¢â‚¬â„¢ÃƒÆ’Ã¢â‚¬Å¡Ãƒâ€šÃ‚Â¢ÃƒÆ’Ã†â€™Ãƒâ€šÃ‚Â¢ÃƒÆ’Ã‚Â¢ÃƒÂ¢Ã¢â‚¬Å¡Ã‚Â¬Ãƒâ€¦Ã‚Â¡ÃƒÆ’Ã¢â‚¬Å¡Ãƒâ€šÃ‚Â¬ÃƒÆ’Ã†â€™ÃƒÂ¢Ã¢â€šÂ¬Ã…Â¡ÃƒÆ’Ã¢â‚¬Å¡Ãƒâ€šÃ‚Â marks the
// order "returned".
func ApproveReturn(c *gin.Context) {
	adminID := c.MustGet("user_id").(uint)
	id := c.Param("id")

	var returnReq models.ReturnRequest
	if err := database.DB.Preload("Items").First(&returnReq, id).Error; err != nil {
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

	orderItemsByID := make(map[uint]models.OrderItem, len(order.Items))
	for _, it := range order.Items {
		orderItemsByID[it.ID] = it
	}

var creditNote *models.CreditNote
	txErr := database.DB.Transaction(func(tx *gorm.DB) error {
		// Restore stock for each returned line item.
		for _, ri := range returnReq.Items {
			orderItem, ok := orderItemsByID[ri.OrderItemID]
			if !ok {
				continue
			}
			var inventory models.Inventory
			q := tx.Where("product_id = ?", orderItem.ProductID)
			if order.WarehouseID != nil {
				q = q.Where("warehouse_id = ?", *order.WarehouseID)
			}
			if err := q.Order("id").First(&inventory).Error; err == nil {
				inventory.Stock += ri.Quantity
				inventory.InStock = true
				if err := tx.Save(&inventory).Error; err != nil {
					return err
				}
			}
		}

		refID := order.ID
		if err := utils.CreditWallet(tx, returnReq.UserID, returnReq.RefundAmount, models.WalletReasonRefund, "return_request", &refID, "Refund for return request #"+id); err != nil {
			return err
		}

// Auto-issue a Credit Note against the order's invoice for this return
// (SRS 12.17). Runs in the same transaction as the stock restore and
// wallet credit above, so a credit note is never created for a return
// that doesn't actually complete.
var cnErr error
creditNote, cnErr = services.GenerateCreditNoteForReturn(tx, returnReq)
if cnErr != nil {
return cnErr
}

		// Check whether every item on the order is now fully covered by
		// approved return requests (including this one, which we're about
		// to mark approved).
		type qtyRow struct {
			OrderItemID uint
			Total       int
		}
		var approvedRows []qtyRow
		if err := tx.Table("return_request_items").
			Select("order_item_id, SUM(quantity) as total").
			Joins("JOIN return_requests ON return_requests.id = return_request_items.return_request_id").
			Where("return_requests.order_id = ? AND (return_requests.status = ? OR return_requests.id = ?)", order.ID, models.ReturnStatusApproved, returnReq.ID).
			Group("order_item_id").
			Scan(&approvedRows).Error; err != nil {
			return err
		}
		approvedQty := make(map[uint]int, len(approvedRows))
		for _, row := range approvedRows {
			approvedQty[row.OrderItemID] = row.Total
		}

		fullyReturned := true
		for _, it := range order.Items {
			if approvedQty[it.ID] < it.Quantity {
				fullyReturned = false
				break
			}
		}

		if fullyReturned {
			if err := tx.Model(&order).Update("status", models.OrderStatusReturned).Error; err != nil {
				return err
			}
		}

		returnReq.Status = models.ReturnStatusApproved
		returnReq.ProcessedBy = &adminID
		return tx.Save(&returnReq).Error
	})

	if txErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to approve return: " + txErr.Error()})
		return
	}

if creditNote != nil {
if err := services.PostCreditNoteLedgerEntry(creditNote.ID); err != nil {
log.Printf("failed to post credit note ledger entry for credit note %d: %v", creditNote.ID, err)
}
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
