package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/database"
	"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/models"
)

// getOrCreateCart returns the user's cart, creating one if it doesn't exist yet
// (signup auto-creates a cart, but this keeps the API resilient for older users/data).
func getOrCreateCart(userID uint) (*models.Cart, error) {
	var cart models.Cart
	err := database.DB.Where("user_id = ?", userID).First(&cart).Error
	if err == nil {
		return &cart, nil
	}

	cart = models.Cart{UserID: userID}
	if err := database.DB.Create(&cart).Error; err != nil {
		return nil, err
	}
	return &cart, nil
}

// buildCartResponse loads items (with product + inventory) and computes totals.
func buildCartResponse(cart *models.Cart) (models.CartResponse, error) {
	var items []models.CartItem
	if err := database.DB.
		Preload("Product").
		Preload("Product.Category").
		Preload("Product.Inventories").
		Where("cart_id = ?", cart.ID).
		Find(&items).Error; err != nil {
		return models.CartResponse{}, err
	}

	totalItems := 0
	totalAmount := 0.0
	for _, item := range items {
		totalItems += item.Quantity
		totalAmount += item.Product.Price * float64(item.Quantity)
	}

	return models.CartResponse{
		ID:          cart.ID,
		Items:       items,
		TotalItems:  totalItems,
		TotalAmount: totalAmount,
	}, nil
}

// GetCart godoc
// GET /api/v1/cart (protected)
func GetCart(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)

	cart, err := getOrCreateCart(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load cart"})
		return
	}

	resp, err := buildCartResponse(cart)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load cart items"})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// AddToCart godoc
// POST /api/v1/cart (protected)
// If the product is already in the cart, quantity is incremented instead of duplicated.
func AddToCart(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)

	var req models.AddToCartRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Ensure product exists
	var product models.Product
	if err := database.DB.First(&product, req.ProductID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		return
	}

	// Check combined stock across all warehouses.
	totalStock, err := database.GetTotalStock(req.ProductID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check stock"})
		return
	}
	if totalStock < req.Quantity {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Insufficient stock for this product"})
		return
	}

	cart, err := getOrCreateCart(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load cart"})
		return
	}

	var existingItem models.CartItem
	err = database.DB.Where("cart_id = ? AND product_id = ?", cart.ID, req.ProductID).First(&existingItem).Error
	if err == nil {
		// Item already in cart Ã¢â‚¬â€ increment quantity
		existingItem.Quantity += req.Quantity
		if err := database.DB.Save(&existingItem).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update cart item"})
			return
		}
	} else {
		newItem := models.CartItem{
			CartID:    cart.ID,
			ProductID: req.ProductID,
			Quantity:  req.Quantity,
		}
		if err := database.DB.Create(&newItem).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add item to cart"})
			return
		}
	}

	resp, err := buildCartResponse(cart)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load cart"})
		return
	}
	c.JSON(http.StatusCreated, resp)
}

// UpdateCartItem godoc
// PUT /api/v1/cart/:item_id (protected)
func UpdateCartItem(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)
	itemID := c.Param("item_id")

	var req models.UpdateCartItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var item models.CartItem
	if err := database.DB.First(&item, itemID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Cart item not found"})
		return
	}

	// Ownership check Ã¢â‚¬â€ item's cart must belong to the requesting user
	var cart models.Cart
	if err := database.DB.First(&cart, item.CartID).Error; err != nil || cart.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "You do not have access to this cart item"})
		return
	}

	// Check combined stock across all warehouses (same rule as AddToCart).
	totalStock, err := database.GetTotalStock(item.ProductID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check stock"})
		return
	}
	if totalStock < req.Quantity {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Insufficient stock for this product"})
		return
	}

	item.Quantity = req.Quantity
	if err := database.DB.Save(&item).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update cart item"})
		return
	}

	resp, err := buildCartResponse(&cart)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load cart"})
		return
	}
	c.JSON(http.StatusOK, resp)
}

// RemoveFromCart godoc
// DELETE /api/v1/cart/:item_id (protected)
func RemoveFromCart(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)
	itemID := c.Param("item_id")

	var item models.CartItem
	if err := database.DB.First(&item, itemID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Cart item not found"})
		return
	}

	var cart models.Cart
	if err := database.DB.First(&cart, item.CartID).Error; err != nil || cart.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "You do not have access to this cart item"})
		return
	}

	if err := database.DB.Delete(&item).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove cart item"})
		return
	}

	resp, err := buildCartResponse(&cart)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load cart"})
		return
	}
	c.JSON(http.StatusOK, resp)
}
