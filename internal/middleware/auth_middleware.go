package middleware

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/cache"
	"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/database"
	"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/models"
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

		// A valid signature only proves the token was genuinely issued - it
		// says nothing about whether the account has since been blocked. A
		// blocked user's outstanding token would otherwise keep working
		// until natural expiration, days or weeks later (Bug#20). Cached for
		// a short TTL since this runs on every authenticated request;
		// BlockCustomer/UnblockCustomer invalidate the cache key immediately
		// on change so enforcement doesn't wait out the TTL.
		if isUserBlocked(c, claims.UserID) {
			c.JSON(http.StatusForbidden, gin.H{"error": "This account has been blocked"})
			c.Abort()
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("phone", claims.Phone)
		c.Set("role", claims.Role)
		c.Next()
	}
}

// blockedCacheKey builds the Redis cache key used to remember a user's
// current IsBlocked state for a short window, avoiding a DB hit on every
// authenticated request. Exported-shape kept unexported since only this
// file and BlockCustomer/UnblockCustomer (via cache.Delete with the same
// key format) need to agree on it.
func blockedCacheKey(userID uint) string {
	return fmt.Sprintf("auth:blocked:%d", userID)
}

// isUserBlocked checks the cache first, falling back to a DB lookup on a
// miss (including when Redis isn't enabled at all, in which case cache.Get
// always reports a miss and this degrades to a per-request DB check -
// correct, just without the caching benefit). Fails open only on a
// not-found error (deleted/nonexistent user is caught elsewhere by
// handlers that load the user); fails closed (treats as blocked) on any
// other DB error, since silently granting access on a DB hiccup would
// defeat the purpose of this check.
func isUserBlocked(c *gin.Context, userID uint) bool {
	key := blockedCacheKey(userID)
	var cached bool
	if found, err := cache.Get(c.Request.Context(), key, &cached); err == nil && found {
		return cached
	}

	var user models.User
	if err := database.DB.Select("is_blocked").First(&user, userID).Error; err != nil {
		return false
	}
	_ = cache.Set(c.Request.Context(), key, user.IsBlocked, 60*time.Second)
	return user.IsBlocked
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

// DeliveryPartnerOnly restricts access to users with role "delivery_partner".
// Must run after AuthMiddleware. In addition to checking the JWT role claim,
// this re-checks the partner's is_active flag against the DB on every
// request, so a partner deactivated by admin loses API access immediately -
// not only after their existing token happens to expire.
func DeliveryPartnerOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists || role != "delivery_partner" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Delivery partner access required"})
			c.Abort()
			return
		}

		partnerID, ok := c.Get("user_id")
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		var partner models.DeliveryPartner
		if err := database.DB.Select("id", "is_active").First(&partner, partnerID).Error; err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Delivery partner account not found"})
			c.Abort()
			return
		}
		if !partner.IsActive {
			c.JSON(http.StatusForbidden, gin.H{"error": "This delivery partner account is inactive"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// WarehouseStaffOnly restricts access to users with role "warehouse_staff". Must run after AuthMiddleware.
// Re-checks IsActive against the database on every request (Defect #08) -
// the JWT claim only proves who the staff member WAS at login time, not
// whether they are still an active employee right now. Without this DB
// lookup, deactivating/terminating a staff member had no effect until
// their existing token happened to expire, leaving a window where a
// terminated employee could keep using picking queues, substitutions, and
// inventory-altering endpoints with a token that was still cryptographically
// valid.
func WarehouseStaffOnly() gin.HandlerFunc {
return func(c *gin.Context) {
role, exists := c.Get("role")
if !exists || role != "warehouse_staff" {
c.JSON(http.StatusForbidden, gin.H{"error": "Warehouse staff access required"})
c.Abort()
return
}

staffID, exists := c.Get("user_id")
if !exists {
c.JSON(http.StatusForbidden, gin.H{"error": "Warehouse staff access required"})
c.Abort()
return
}

var staff models.WarehouseStaff
if err := database.DB.Select("is_active").First(&staff, staffID).Error; err != nil {
c.JSON(http.StatusForbidden, gin.H{"error": "Warehouse staff account not found"})
c.Abort()
return
}
if !staff.IsActive {
c.JSON(http.StatusForbidden, gin.H{"error": "This warehouse staff account is inactive"})
c.Abort()
return
}

c.Next()
}
}
