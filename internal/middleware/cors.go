package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// CORS allows cross-origin requests so browser-based clients (a web admin
// panel, Flutter web build, etc.) can call this API directly. Native mobile
// apps aren't affected by CORS and don't need this, but it's required the
// moment anything runs in a browser.
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Authorization, Content-Type")

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}
