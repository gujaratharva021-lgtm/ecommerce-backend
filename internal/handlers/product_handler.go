package handlers

import (
	"math"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/database"
	"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/models"
)

// GetProducts godoc
// GET /api/v1/products
// Supports: ?search=&category_id=&min_price=&max_price=&in_stock=&sort=&page=&limit=
func GetProducts(c *gin.Context) {
	var query models.ProductListQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if query.Page < 1 {
		query.Page = 1
	}
	if query.Limit < 1 || query.Limit > 100 {
		query.Limit = 20
	}

	db := database.DB.Model(&models.Product{}).Preload("Category").Preload("Inventory")

	// Search by name or description
	if strings.TrimSpace(query.Search) != "" {
		like := "%" + strings.TrimSpace(query.Search) + "%"
		db = db.Where("name ILIKE ? OR description ILIKE ?", like, like)
	}

	// Filter by category
	if query.CategoryID > 0 {
		db = db.Where("category_id = ?", query.CategoryID)
	}

	// Filter by price range
	if query.MinPrice > 0 {
		db = db.Where("price >= ?", query.MinPrice)
	}
	if query.MaxPrice > 0 {
		db = db.Where("price <= ?", query.MaxPrice)
	}

	// Filter by stock availability (joins inventory)
	if query.InStock != nil {
		db = db.Joins("JOIN inventories ON inventories.product_id = products.id").
			Where("inventories.in_stock = ?", *query.InStock)
	}

	// Count total before pagination
	var total int64
	if err := db.Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count products"})
		return
	}

	// Sorting
	switch query.Sort {
	case "price_asc":
		db = db.Order("price ASC")
	case "price_desc":
		db = db.Order("price DESC")
	case "name_asc":
		db = db.Order("name ASC")
	case "name_desc":
		db = db.Order("name DESC")
	case "newest":
		db = db.Order("created_at DESC")
	default:
		db = db.Order("created_at DESC")
	}

	// Pagination
	offset := (query.Page - 1) * query.Limit
	var products []models.Product
	if err := db.Offset(offset).Limit(query.Limit).Find(&products).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch products"})
		return
	}

	totalPages := int(math.Ceil(float64(total) / float64(query.Limit)))

	c.JSON(http.StatusOK, models.ProductListResponse{
		Products:   products,
		Page:       query.Page,
		Limit:      query.Limit,
		Total:      total,
		TotalPages: totalPages,
	})
}

// GetProductByID godoc
// GET /api/v1/products/:id
func GetProductByID(c *gin.Context) {
	id := c.Param("id")

	var product models.Product
	if err := database.DB.Preload("Category").Preload("Inventory").First(&product, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		return
	}

	c.JSON(http.StatusOK, product)
}

// GetCategories godoc
// GET /api/v1/categories
func GetCategories(c *gin.Context) {
	var categories []models.Category
	if err := database.DB.Order("name ASC").Find(&categories).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch categories"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"categories": categories})
}
