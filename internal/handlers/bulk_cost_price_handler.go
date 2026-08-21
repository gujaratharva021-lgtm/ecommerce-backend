package handlers

import (
"net/http"

"github.com/gin-gonic/gin"
"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/database"
)

type bulkCostPriceItem struct {
ID        uint    `json:"id" binding:"required"`
CostPrice float64 `json:"cost_price" binding:"gte=0"`
}

type bulkCostPriceReq struct {
Items []bulkCostPriceItem `json:"items" binding:"required,dive"`
}

func BulkUpdateCostPrice(c *gin.Context) {
var req bulkCostPriceReq
if err := c.ShouldBindJSON(&req); err != nil {
c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
return
}

updated := 0
failed := []map[string]interface{}{}

for _, item := range req.Items {
res := database.DB.Table("products").Where("id = ?", item.ID).Update("cost_price", item.CostPrice)
if res.Error != nil {
failed = append(failed, map[string]interface{}{"id": item.ID, "error": res.Error.Error()})
continue
}
if res.RowsAffected == 0 {
failed = append(failed, map[string]interface{}{"id": item.ID, "error": "product not found"})
continue
}
updated++
}

c.JSON(http.StatusOK, gin.H{
"total":   len(req.Items),
"updated": updated,
"failed":  failed,
})
}
