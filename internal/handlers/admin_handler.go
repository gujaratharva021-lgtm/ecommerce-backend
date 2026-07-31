package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/database"
	"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/models"
	"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/services"
	"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/utils"
	"gorm.io/gorm"
)

// ---------------------------------------------------------------------------
// Categories
// ---------------------------------------------------------------------------

// CreateCategory godoc
// POST /api/v1/admin/categories (admin only)
func CreateCategory(c *gin.Context) {
	var req models.CategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	category := models.Category{Name: req.Name, ImageURL: req.ImageURL}
	if err := database.DB.Create(&category).Error; err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Category already exists or could not be created"})
		return
	}

	c.JSON(http.StatusCreated, category)
}

// UpdateCategory godoc
// PUT /api/v1/admin/categories/:id (admin only)
func UpdateCategory(c *gin.Context) {
	id := c.Param("id")

	var category models.Category
	if err := database.DB.First(&category, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Category not found"})
		return
	}

	var req models.CategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	category.Name = req.Name
	category.ImageURL = req.ImageURL
	if err := database.DB.Save(&category).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update category"})
		return
	}

	c.JSON(http.StatusOK, category)
}

// DeleteCategory godoc
// DELETE /api/v1/admin/categories/:id (admin only)
// Blocked if any product still references this category, so products never
// end up pointing at a category that no longer exists.
func DeleteCategory(c *gin.Context) {
	id := c.Param("id")

	var category models.Category
	if err := database.DB.First(&category, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Category not found"})
		return
	}

	var productCount int64
	database.DB.Model(&models.Product{}).Where("category_id = ?", category.ID).Count(&productCount)
	if productCount > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot delete category with existing products. Reassign or delete those products first."})
		return
	}

	if err := database.DB.Delete(&category).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete category"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Category deleted"})
}

// ---------------------------------------------------------------------------
// Products
// ---------------------------------------------------------------------------

// CreateProduct godoc
// POST /api/v1/admin/products (admin only)
// Creates the product and its Inventory row (seeded with req.Stock) together.
func CreateProduct(c *gin.Context) {
	var req models.ProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var category models.Category
	if err := database.DB.First(&category, req.CategoryID).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Category not found"})
		return
	}

	product := models.Product{
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		ImageURL:    req.ImageURL,
		CategoryID:  req.CategoryID,
	}

	txErr := database.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&product).Error; err != nil {
			return err
		}
		inventory := models.Inventory{
			ProductID: product.ID,
			Stock:     req.Stock,
			InStock:   req.Stock > 0,
		}
		return tx.Create(&inventory).Error
	})

	if txErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create product"})
		return
	}

	database.DB.Preload("Category").Preload("Inventory").First(&product, product.ID)
	c.JSON(http.StatusCreated, product)
}

// UpdateProduct godoc
// PUT /api/v1/admin/products/:id (admin only)
// Updates product fields only Ã¢â‚¬â€ use PUT /admin/products/:id/inventory for stock.
func UpdateProduct(c *gin.Context) {
	id := c.Param("id")

	var product models.Product
	if err := database.DB.First(&product, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		return
	}

	var req models.ProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var category models.Category
	if err := database.DB.First(&category, req.CategoryID).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Category not found"})
		return
	}

	product.Name = req.Name
	product.Description = req.Description
	product.Price = req.Price
	product.ImageURL = req.ImageURL
	product.CategoryID = req.CategoryID

	if err := database.DB.Save(&product).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update product"})
		return
	}

	database.DB.Preload("Category").Preload("Inventory").First(&product, product.ID)
	c.JSON(http.StatusOK, product)
}

// DeleteProduct godoc
// DELETE /api/v1/admin/products/:id (admin only)
func DeleteProduct(c *gin.Context) {
	id := c.Param("id")

	var product models.Product
	if err := database.DB.First(&product, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		return
	}

	txErr := database.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("product_id = ?", product.ID).Delete(&models.Inventory{}).Error; err != nil {
			return err
		}
		return tx.Delete(&product).Error
	})

	if txErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete product"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Product deleted"})
}

// UpdateInventory godoc
// PUT /api/v1/admin/products/:id/inventory (admin only)
// Sets stock to an absolute value and recomputes in_stock from it.
func UpdateInventory(c *gin.Context) {
	id := c.Param("id")

	var product models.Product
	if err := database.DB.First(&product, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		return
	}

	var req models.InventoryUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var inventory models.Inventory
	if err := database.DB.Where("product_id = ?", product.ID).First(&inventory).Error; err != nil {
		// Product predates the inventory row (shouldn't normally happen since
		// CreateProduct always creates one) Ã¢â‚¬â€ create it now instead of failing.
		inventory = models.Inventory{ProductID: product.ID}
	}

	inventory.Stock = req.Stock
	inventory.InStock = req.Stock > 0

	if err := database.DB.Save(&inventory).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update inventory"})
		return
	}

	c.JSON(http.StatusOK, inventory)
}

// ---------------------------------------------------------------------------
// Orders
// ---------------------------------------------------------------------------

// validOrderTransitions defines which status changes an admin may make.
// pending and confirmed orders can also be cancelled (stock is restored).
var validOrderTransitions = map[string]map[string]bool{
	models.OrderStatusPending:   {models.OrderStatusConfirmed: true, models.OrderStatusCancelled: true},
	models.OrderStatusConfirmed: {models.OrderStatusShipped: true, models.OrderStatusCancelled: true},
	models.OrderStatusShipped:   {models.OrderStatusDelivered: true},
}

// GetAllOrders godoc
// GET /api/v1/admin/orders (admin only) Ã¢â‚¬â€ ?status=&page=&limit=
func GetAllOrders(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	db := database.DB.Model(&models.Order{})
	if status := c.Query("status"); status != "" {
		db = db.Where("status = ?", status)
	}

	var total int64
	db.Count(&total)

	var orders []models.Order
	if err := db.
		Preload("Items.Product").
		Preload("Address").
		Order("created_at DESC").
		Offset((page - 1) * limit).
		Limit(limit).
		Find(&orders).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load orders"})
		return
	}

	totalPages := int((total + int64(limit) - 1) / int64(limit))
	c.JSON(http.StatusOK, models.OrderListResponse{
		Orders:     orders,
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: totalPages,
	})
}

// UpdateOrderStatus godoc
// PUT /api/v1/admin/orders/:id/status (admin only)
// Enforces the pending -> confirmed -> shipped -> delivered flow (plus
// cancellation from pending/confirmed) and restores stock on cancellation,
// same as the customer-facing CancelOrder.
func UpdateOrderStatus(c *gin.Context) {
	orderID := c.Param("id")

	var order models.Order
	if err := database.DB.Preload("Items.Product").Preload("Address").First(&order, orderID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
		return
	}

	var req models.OrderStatusUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	allowed := validOrderTransitions[order.Status]
	if !allowed[req.Status] {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Cannot change status from '" + order.Status + "' to '" + req.Status + "'",
		})
		return
	}

	txErr := database.DB.Transaction(func(tx *gorm.DB) error {
		if req.Status == models.OrderStatusCancelled {
			for _, item := range order.Items {
				var inventory models.Inventory
				if err := tx.Where("product_id = ?", item.ProductID).First(&inventory).Error; err == nil {
					inventory.Stock += item.Quantity
					inventory.InStock = true
					if err := tx.Save(&inventory).Error; err != nil {
						return err
					}
				}
			}
		}
		return tx.Model(&order).Update("status", req.Status).Error
	})

	if txErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update order status"})
		return
	}

	order.Status = req.Status

	message := "Your order #" + orderID + " status is now: " + req.Status
	utils.SendNotification(order.Address.Phone, message, "order_status_"+req.Status, &order.ID)
	services.SendPushToUser(order.UserID, "Order Update", message)

	c.JSON(http.StatusOK, order)
}


