package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// NOTE: These are placeholder stubs for Day 1.
// Full implementation (DB queries, filters, search) happens on Day 2
// as per the task plan: "Implement Product APIs".

// GetProducts godoc
// GET /api/v1/products
func GetProducts(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Get products endpoint - to be implemented on Day 2"})
}

// GetProductByID godoc
// GET /api/v1/products/:id
func GetProductByID(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Get product details endpoint - to be implemented on Day 2"})
}

// GetCategories godoc
// GET /api/v1/categories
func GetCategories(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Get categories endpoint - to be implemented on Day 2"})
}
