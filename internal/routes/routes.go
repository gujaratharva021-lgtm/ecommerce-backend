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
	router.HEAD("/health", handlers.HealthCheck)

	// Serve uploaded images (e.g. /uploads/169999.jpg)
	router.Static("/uploads", "./uploads")

	api := router.Group("/api/v1")
	{
		// ---- Auth routes (public) ----
		auth := api.Group("/auth")
		{
                        // Rate-limited - OTP endpoints are otherwise open to spam and brute force.
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

			products.GET("/:id/reviews", handlers.GetProductReviews)
			products.POST("/:id/reviews", middleware.AuthMiddleware(), handlers.UpsertReview)
			products.DELETE("/:id/reviews", middleware.AuthMiddleware(), handlers.DeleteReview)
		}

		// ---- Category routes (public) ----
		categories := api.Group("/categories")
		{
			categories.GET("", handlers.GetCategories)
		}

		api.POST("/device-token", handlers.RegisterDeviceToken)

		// ---- Serviceability routes (public) ----
		api.GET("/serviceability", handlers.CheckServiceability)
			api.GET("/run-cart-reservation-migration", handlers.RunCartReservationMigration)

		// ---- Notification routes (protected) ----
		api.GET("/notifications", middleware.AuthMiddleware(), handlers.GetMyNotifications)

		// ---- Wallet routes (protected) ----
		api.GET("/wallet", middleware.AuthMiddleware(), handlers.GetWallet)
		api.GET("/returns", middleware.AuthMiddleware(), handlers.GetMyReturns)

		delivery := api.Group("/delivery")
		{
			delivery.POST("/send-otp", middleware.RateLimit(5, time.Minute), handlers.SendPartnerOTP)
			delivery.POST("/verify-otp", middleware.RateLimit(10, time.Minute), handlers.VerifyPartnerOTP)
			delivery.PUT("/location", middleware.AuthMiddleware(), middleware.DeliveryPartnerOnly(), handlers.UpdateLocation)
			delivery.GET("/orders", middleware.AuthMiddleware(), middleware.DeliveryPartnerOnly(), handlers.GetMyDeliveries)
			delivery.PUT("/orders/:id/status", middleware.AuthMiddleware(), middleware.DeliveryPartnerOnly(), handlers.UpdateDeliveryOrderStatus)
			delivery.PUT("/orders/:id/deliver", middleware.AuthMiddleware(), middleware.DeliveryPartnerOnly(), handlers.ConfirmDelivery)
			delivery.GET("/earnings", middleware.AuthMiddleware(), middleware.DeliveryPartnerOnly(), handlers.GetMyEarnings)
		}

		warehouse := api.Group("/warehouse")
		{
			warehouse.POST("/send-otp", middleware.RateLimit(5, time.Minute), handlers.SendWarehouseStaffOTP)
			warehouse.POST("/verify-otp", middleware.RateLimit(10, time.Minute), handlers.VerifyWarehouseStaffOTP)

			warehouseStockTransfers := warehouse.Group("/stock-transfers")
			warehouseStockTransfers.Use(middleware.AuthMiddleware(), middleware.WarehouseStaffOnly())
			{
				warehouseStockTransfers.POST("", handlers.RequestStockTransfer)
				warehouseStockTransfers.GET("", handlers.GetMyStockTransfers)
				warehouseStockTransfers.PUT("/:id/receive", handlers.ReceiveStockTransfer)
				warehouseStockTransfers.PUT("/:id/approve", handlers.ApproveStockTransferByWarehouseStaff)
				warehouseStockTransfers.PUT("/:id/reject", handlers.RejectStockTransferByWarehouseStaff)
			}
		}

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

		// ---- Wishlist routes (protected) ----
		wishlist := api.Group("/wishlist")
		wishlist.Use(middleware.AuthMiddleware())
		{
			wishlist.GET("", handlers.GetWishlist)
			wishlist.POST("", handlers.AddToWishlist)
			wishlist.DELETE("/:product_id", handlers.RemoveFromWishlist)
		}

		// ---- Order routes (protected) ----
		orders := api.Group("/orders")
		orders.Use(middleware.AuthMiddleware())
		{
			orders.POST("/checkout", handlers.Checkout) // body: { address_id?, payment_method?: "cod"|"online" }
			orders.GET("", handlers.GetOrders)
			orders.GET("/:id", handlers.GetOrderByID)
			orders.GET("/:id/tracking", handlers.GetOrderTracking)
			orders.PUT("/:id/cancel", handlers.CancelOrder)
			orders.POST("/:id/return", handlers.RequestReturn)
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

			adminReturns := admin.Group("/returns")
			{
				adminReturns.GET("", handlers.GetReturns)
				adminReturns.PUT("/:id/approve", handlers.ApproveReturn)
				adminReturns.PUT("/:id/reject", handlers.RejectReturn)
			}

			adminCoupons := admin.Group("/coupons")
			{
				adminCoupons.POST("", handlers.CreateCoupon)
				adminCoupons.GET("", handlers.GetCoupons)
				adminCoupons.PUT("/:id/status", handlers.UpdateCouponStatus)
			}
			adminAnalytics := admin.Group("/analytics")
			{
				adminAnalytics.GET("/summary", handlers.GetAnalyticsSummary)
				adminAnalytics.GET("/products", handlers.GetProductPerformance)
			}

			adminDeliveryPartners := admin.Group("/delivery-partners")
			{
				adminDeliveryPartners.POST("", handlers.CreateDeliveryPartner)
				adminDeliveryPartners.GET("", handlers.GetDeliveryPartners)
				adminDeliveryPartners.PUT("/:id", handlers.UpdateDeliveryPartner)
				adminDeliveryPartners.DELETE("/:id", handlers.DeleteDeliveryPartner)
			}

			admin.PUT("/orders/:id/assign-delivery", handlers.AssignDeliveryPartner)

			adminWarehouses := admin.Group("/warehouses")
			{
				adminWarehouses.POST("", handlers.CreateWarehouse)
				adminWarehouses.GET("", handlers.GetWarehouses)
				adminWarehouses.GET("/:id", handlers.GetWarehouse)
				adminWarehouses.PUT("/:id", handlers.UpdateWarehouse)
				adminWarehouses.DELETE("/:id", handlers.DeleteWarehouse)

				adminWarehouseStaff := admin.Group("/warehouse-staff")
				{
					adminWarehouseStaff.POST("", handlers.CreateWarehouseStaff)
					adminWarehouseStaff.GET("", handlers.GetWarehouseStaff)
					adminWarehouseStaff.PUT("/:id", handlers.UpdateWarehouseStaff)
					adminWarehouseStaff.DELETE("/:id", handlers.DeleteWarehouseStaff)
				}

				adminStockTransfers := admin.Group("/stock-transfers")
				{
					adminStockTransfers.GET("", handlers.GetStockTransfers)
					adminStockTransfers.PUT("/:id/approve", handlers.ApproveStockTransfer)
					adminStockTransfers.PUT("/:id/reject", handlers.RejectStockTransfer)
				}
                                adminStockTransfers.PUT("/:id/cancel", handlers.CancelStockTransfer)

				admin.POST("/wallet/credit/:user_id", handlers.AdminCreditWallet)
			}
		}
	}
}

