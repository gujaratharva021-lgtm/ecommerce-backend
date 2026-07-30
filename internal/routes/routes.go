package routes

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/handlers"
	"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/middleware"
)

// SetupRoutes defines every API endpoint for the application.
func SetupRoutes(router *gin.Engine) {
	router.Use(middleware.CORS())

	router.GET("/health", handlers.HealthCheck)

	// Serve uploaded images (e.g. /uploads/169999.jpg)
	router.Static("/uploads", "./uploads")

	api := router.Group("/api/v1")
	{
		// ---- Auth routes (public) ----
		auth := api.Group("/auth")
		{
			// Rate-limited â€” OTP endpoints are otherwise open to spam and brute force.
			auth.POST("/send-otp", middleware.RateLimit(5, time.Minute), handlers.SendOTP)
			auth.POST("/verify-otp", middleware.RateLimit(10, time.Minute), handlers.VerifyOTP)
			auth.GET("/me", middleware.AuthMiddleware(), handlers.Me)
			auth.PUT("/me", middleware.AuthMiddleware(), handlers.UpdateProfile)
		}

		// ---- Product routes (public) ----
		products := api.Group("/products")
		{
			products.GET("", handlers.GetProducts) // ?search=&category_id=&min_price=&max_price=&in_stock=&sort=&page=&limit=
			products.GET("/:id", handlers.GetProductByID)
		}

		// ---- Category routes (public) ----
		categories := api.Group("/categories")
		{
			categories.GET("", handlers.GetCategories)
		}

		api.POST("/device-token", handlers.RegisterDeviceToken)

		// ---- Cart routes (protected) ----
		cart := api.Group("/cart")
		cart.Use(middleware.AuthMiddleware())
		{
			cart.GET("", handlers.GetCart)
			cart.POST("", handlers.AddToCart)
			cart.PUT("/:item_id", handlers.UpdateCartItem)
			cart.DELETE("/:item_id", handlers.RemoveFromCart)
		}

		// ---- Address routes (protected) ----
		addresses := api.Group("/addresses")
		addresses.Use(middleware.AuthMiddleware())
		{
			addresses.GET("", handlers.ListAddresses)
			addresses.POST("", handlers.CreateAddress)
			addresses.PUT("/:id", handlers.UpdateAddress)
			addresses.DELETE("/:id", handlers.DeleteAddress)
			addresses.PUT("/:id/default", handlers.SetDefaultAddress)
		}

		// ---- Order routes (protected) ----
		orders := api.Group("/orders")
		orders.Use(middleware.AuthMiddleware())
		{
			orders.POST("/checkout", handlers.Checkout) // body: { address_id?, payment_method?: "cod"|"online" }
			orders.GET("", handlers.GetOrders)
			orders.GET("/:id", handlers.GetOrderByID)
			orders.PUT("/:id/cancel", handlers.CancelOrder)
			orders.POST("/:id/payment", handlers.CreatePaymentOrder)   // creates Razorpay order (payment_method: online only)
			orders.POST("/:id/payment/verify", handlers.VerifyPayment) // verifies signature, marks order paid
		}

		// ---- Coupon routes (protected) ----
		coupons := api.Group("/coupons")
		coupons.Use(middleware.AuthMiddleware())
		{
			coupons.POST("/validate", handlers.ValidateCouponHandler)
		}

		// ---- Upload routes (protected, structure ready for product/category images) ----
		upload := api.Group("/upload")
		upload.Use(middleware.AuthMiddleware())
		{
			upload.POST("", handlers.UploadImage)
		}

		// ---- Admin routes (protected, admin role only) ----
		admin := api.Group("/admin")
		admin.Use(middleware.AuthMiddleware(), middleware.AdminOnly())
		{
			adminCategories := admin.Group("/categories")
			{
				adminCategories.POST("", handlers.CreateCategory)
				adminCategories.PUT("/:id", handlers.UpdateCategory)
				adminCategories.DELETE("/:id", handlers.DeleteCategory)
			}

			adminProducts := admin.Group("/products")
			{
				adminProducts.POST("", handlers.CreateProduct)
				adminProducts.PUT("/:id", handlers.UpdateProduct)
				adminProducts.DELETE("/:id", handlers.DeleteProduct)
				adminProducts.PUT("/:id/inventory", handlers.UpdateInventory)
			}

			adminOrders := admin.Group("/orders")
			{
				adminOrders.GET("", handlers.GetAllOrders) // ?status=&page=&limit=
				adminOrders.PUT("/:id/status", handlers.UpdateOrderStatus)
			}

			adminCoupons := admin.Group("/coupons")
			{
				adminCoupons.POST("", handlers.CreateCoupon)
				adminCoupons.GET("", handlers.GetCoupons)
				adminCoupons.PUT("/:id/status", handlers.UpdateCouponStatus)
			}
		}
	}
}

