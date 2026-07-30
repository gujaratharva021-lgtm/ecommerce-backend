package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/database"
	"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/models"
	"gorm.io/gorm"
)

// freeDeliveryThreshold and flatDeliveryCharge implement a simple, common
// Indian-ecommerce delivery pricing rule: free delivery above the threshold,
// otherwise a flat charge. Swap for a real shipping/rate-card service later.
const (
	freeDeliveryThreshold = 500.0
	flatDeliveryCharge    = 50.0
)

// Checkout godoc
// POST /api/v1/orders/checkout (protected)
// Converts the user's current cart into an order: validates stock for every
// item, snapshots prices, deducts inventory, creates the order + order
// items, and empties the cart — all inside a single DB transaction so a
// failure partway through never leaves stock or the cart in a bad state.
func Checkout(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)

	var req models.CheckoutRequest
	// Body is optional — an empty/missing body just means "use default address".
	_ = c.ShouldBindJSON(&req)

	// Resolve the delivery address: explicit address_id, else the user's default.
	var address models.Address
	if req.AddressID != 0 {
		if err := database.DB.First(&address, req.AddressID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Address not found"})
			return
		}
		if address.UserID != userID {
			c.JSON(http.StatusForbidden, gin.H{"error": "You do not have access to this address"})
			return
		}
	} else {
		if err := database.DB.Where("user_id = ? AND is_default = ?", userID, true).First(&address).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "No address_id given and no default address saved. Add a delivery address first."})
			return
		}
	}

	cart, err := getOrCreateCart(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load cart"})
		return
	}

	var cartItems []models.CartItem
	if err := database.DB.Preload("Product").Where("cart_id = ?", cart.ID).Find(&cartItems).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load cart items"})
		return
	}
	if len(cartItems) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Your cart is empty"})
		return
	}

	var order models.Order

	txErr := database.DB.Transaction(func(tx *gorm.DB) error {
		itemsAmount := 0.0
		orderItems := make([]models.OrderItem, 0, len(cartItems))

		for _, ci := range cartItems {
			var inventory models.Inventory
			if err := tx.Where("product_id = ?", ci.ProductID).First(&inventory).Error; err == nil {
				if !inventory.InStock || inventory.Stock < ci.Quantity {
					return errors.New("insufficient stock for " + ci.Product.Name)
				}
				inventory.Stock -= ci.Quantity
				if inventory.Stock <= 0 {
					inventory.InStock = false
				}
				if err := tx.Save(&inventory).Error; err != nil {
					return err
				}
			}

			itemsAmount += ci.Product.Price * float64(ci.Quantity)
			orderItems = append(orderItems, models.OrderItem{
				ProductID: ci.ProductID,
				Quantity:  ci.Quantity,
				Price:     ci.Product.Price,
			})
		}

		deliveryCharge := flatDeliveryCharge
		if itemsAmount >= freeDeliveryThreshold {
			deliveryCharge = 0
		}

		order = models.Order{
			UserID:         userID,
			AddressID:      address.ID,
			ItemsAmount:    itemsAmount,
			DeliveryCharge: deliveryCharge,
			TotalAmount:    itemsAmount + deliveryCharge,
			Status:         models.OrderStatusPending,
			Items:          orderItems,
		}
		if err := tx.Create(&order).Error; err != nil {
			return err
		}

		if err := tx.Where("cart_id = ?", cart.ID).Delete(&models.CartItem{}).Error; err != nil {
			return err
		}

		return nil
	})

	if txErr != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": txErr.Error()})
		return
	}

	database.DB.Preload("Items.Product").Preload("Address").First(&order, order.ID)
	c.JSON(http.StatusCreated, order)
}

// GetOrders godoc
// GET /api/v1/orders (protected) — ?page=&limit=
func GetOrders(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	var total int64
	database.DB.Model(&models.Order{}).Where("user_id = ?", userID).Count(&total)

	var orders []models.Order
	if err := database.DB.
		Preload("Items.Product").
		Preload("Address").
		Where("user_id = ?", userID).
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

// GetOrderByID godoc
// GET /api/v1/orders/:id (protected)
func GetOrderByID(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)
	orderID := c.Param("id")

	var order models.Order
	if err := database.DB.Preload("Items.Product").Preload("Address").First(&order, orderID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
		return
	}
	if order.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "You do not have access to this order"})
		return
	}

	c.JSON(http.StatusOK, order)
}

// CancelOrder godoc
// PUT /api/v1/orders/:id/cancel (protected)
// Only pending or confirmed orders can be cancelled; stock is restored.
func CancelOrder(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)
	orderID := c.Param("id")

	var order models.Order
	if err := database.DB.Preload("Items.Product").Preload("Address").First(&order, orderID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
		return
	}
	if order.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "You do not have access to this order"})
		return
	}
	if order.Status != models.OrderStatusPending && order.Status != models.OrderStatusConfirmed {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Only pending or confirmed orders can be cancelled"})
		return
	}

	txErr := database.DB.Transaction(func(tx *gorm.DB) error {
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
		return tx.Model(&order).Update("status", models.OrderStatusCancelled).Error
	})

	if txErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to cancel order"})
		return
	}

	order.Status = models.OrderStatusCancelled
	c.JSON(http.StatusOK, order)
}
