package handlers

import (
"net/http"

"github.com/gin-gonic/gin"
"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/database"
)

func DebugCartReservations(c *gin.Context) {
var results []map[string]interface{}
err := database.DB.Table("cart_reservations").
Where("product_id = ?", 393).
Find(&results).Error
if err != nil {
c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
return
}
c.JSON(http.StatusOK, gin.H{"reservations": results})
}
