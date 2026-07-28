package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// NOTE: Placeholder stubs for Day 1. Full implementation happens on Day 2
// as per the task plan: "Create cart module".

// AddToCart godoc
// POST /api/v1/cart (protected)
func AddToCart(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Add to cart endpoint - to be implemented on Day 2"})
}

// GetCart godoc
// GET /api/v1/cart (protected)
func GetCart(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Get cart endpoint - to be implemented on Day 2"})
}
