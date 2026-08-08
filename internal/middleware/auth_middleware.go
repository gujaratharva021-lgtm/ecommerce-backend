package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/utils"
)

// AuthMiddleware validates the "Authorization: Bearer <token>" header
// and injects user_id, email, role into the Gin context for handlers to use.
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header missing"})
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header format"})
			c.Abort()
			return
		}

		claims, err := utils.ValidateJWT(parts[1])
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("phone", claims.Phone)
		c.Set("role", claims.Role)
		c.Next()
	}
}

// AdminOnly restricts access to users with role "admin". Must run after AuthMiddleware.
func AdminOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists || role != "admin" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Admin access required"})
			c.Abort()
			return
		}
		c.Next()
	}
}

// DeliveryPartnerOnly restricts access to users with role "delivery_partner". Must run after AuthMiddleware.
func DeliveryPartnerOnly() gin.HandlerFunc {
return func(c *gin.Context) {
role, exists := c.Get("role")
if !exists || role != "delivery_partner" {
c.JSON(http.StatusForbidden, gin.H{"error": "Delivery partner access required"})
c.Abort()
return
}
c.Next()
}
}

// WarehouseStaffOnly restricts access to users with role "warehouse_staff". Must run after AuthMiddleware.
func WarehouseStaffOnly() gin.HandlerFunc {
return func(c *gin.Context) {
role, exists := c.Get("role")
if !exists || role != "warehouse_staff" {
c.JSON(http.StatusForbidden, gin.H{"error": "Warehouse staff access required"})
c.Abort()
return
}
c.Next()
}
}

// CustomerOnly ensures the authenticated principal is a customer, not a
// delivery partner or warehouse staff member re-using their own numeric ID
// against customer-scoped routes (wallet, cart, orders, etc.). AuthMiddleware
// only proves "is logged in", not "is a customer" -- this closes that gap.
func CustomerOnly() gin.HandlerFunc {
return func(c *gin.Context) {
role, exists := c.Get("role")
if !exists || role != "customer" {
c.JSON(http.StatusForbidden, gin.H{"error": "Customer access required"})
c.Abort()
return
}
c.Next()
}
}
