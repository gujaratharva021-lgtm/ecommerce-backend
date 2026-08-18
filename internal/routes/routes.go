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
	api.GET("/debug-postgis", middleware.AuthMiddleware(), middleware.AdminOnly(), handlers.DebugCheckPostGIS)
	api.GET("/offers", handlers.GetActiveOffers)
	api.GET("/banners", handlers.GetActiveBanners)
	api.GET("/delivery-zones/check", handlers.CheckPincode)
	support := api.Group("/support")
	support.Use(middleware.AuthMiddleware())
	{
		support.POST("/tickets", handlers.CreateTicket)
		support.GET("/tickets", handlers.GetMyTickets)
		support.GET("/tickets/:id/messages", handlers.GetTicketMessages)
		support.POST("/tickets/:id/messages", handlers.ReplyToTicket)
	}

		// ---- Notification routes (protected) ----
		api.GET("/notifications", middleware.AuthMiddleware(), handlers.GetMyNotifications)

		// ---- Wallet routes (protected) ----
		api.GET("/wallet", middleware.AuthMiddleware(), handlers.GetWallet)
		api.GET("/returns", middleware.AuthMiddleware(), handlers.GetMyReturns)

		delivery := api.Group("/delivery")
		{
			delivery.POST("/send-otp", middleware.RateLimit(5, time.Minute), handlers.SendPartnerOTP)
			delivery.POST("/verify-otp", middleware.RateLimit(10, time.Minute), handlers.VerifyPartnerOTP)
			delivery.GET("/profile", middleware.AuthMiddleware(), middleware.DeliveryPartnerOnly(), handlers.GetDeliveryProfile)
			delivery.PUT("/profile", middleware.AuthMiddleware(), middleware.DeliveryPartnerOnly(), handlers.UpdateDeliveryProfile)
			delivery.GET("/availability", middleware.AuthMiddleware(), middleware.DeliveryPartnerOnly(), handlers.GetDeliveryAvailability)
			delivery.PUT("/availability", middleware.AuthMiddleware(), middleware.DeliveryPartnerOnly(), handlers.UpdateDeliveryAvailability)
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
			warehouseStockTransfers.Use(middleware.AuthMiddleware(), middleware.WarehouseStaffOnly(), middleware.InjectWarehouseScope())
			{
				warehouseStockTransfers.POST("", middleware.RequireWarehouseRole(middleware.RoleInventoryOps...), handlers.RequestStockTransfer)
				warehouseStockTransfers.GET("", middleware.RequireWarehouseRole(middleware.RoleAnyStaff...), handlers.GetMyStockTransfers)
				warehouseStockTransfers.PUT("/:id/receive", middleware.RequireWarehouseRole(middleware.RoleInventoryOps...), handlers.ReceiveStockTransfer)
				warehouseStockTransfers.PUT("/:id/approve", middleware.RequireWarehouseRole(middleware.RoleManagement...), handlers.ApproveStockTransferByWarehouseStaff)
				warehouseStockTransfers.PUT("/:id/reject", middleware.RequireWarehouseRole(middleware.RoleManagement...), handlers.RejectStockTransferByWarehouseStaff)
			}

			warehouseOrders := warehouse.Group("/orders")
			warehouseOrders.Use(middleware.AuthMiddleware(), middleware.WarehouseStaffOnly(), middleware.InjectWarehouseScope())
			{
				warehouseOrders.GET("", middleware.RequireWarehouseRole(middleware.RoleAnyStaff...), handlers.GetWarehouseOrders)
				warehouseOrders.PUT("/:id/accept", middleware.RequireWarehouseRole(middleware.RoleAnyStaff...), handlers.AcceptOrder)
				warehouseOrders.PUT("/:id/handover", middleware.RequireWarehouseRole(middleware.RoleManagement...), handlers.HandoverOrder)
				warehouseOrders.GET("/:id/handover", middleware.RequireWarehouseRole(middleware.RoleAnyStaff...), handlers.GetHandover)
warehouseOrders.GET("/:id/invoice", middleware.RequireWarehouseRole(middleware.RoleAnyStaff...), handlers.GetOrderInvoice)
warehouseOrders.GET("/:id/invoice/pdf", middleware.RequireWarehouseRole(middleware.RoleAnyStaff...), handlers.GetWarehouseOrderInvoicePDF)
			}

			warehousePicking := warehouse.Group("/picking")
			warehousePicking.Use(middleware.AuthMiddleware(), middleware.WarehouseStaffOnly(), middleware.InjectWarehouseScope(), middleware.RequireWarehouseRole(middleware.RolePickers...))
			{
				warehousePicking.GET("/:order_id", handlers.GetPickingTask)
				warehousePicking.PUT("/:order_id/start", handlers.StartPicking)
				warehousePicking.PUT("/:order_id/complete", handlers.CompletePicking)
				warehousePicking.PUT("/items/:item_id", handlers.MarkPickItem)
				warehousePicking.PUT("/items/:item_id/scan", handlers.ScanPickItem)
			}

			warehousePacking := warehouse.Group("/packing")
			warehousePacking.Use(middleware.AuthMiddleware(), middleware.WarehouseStaffOnly(), middleware.InjectWarehouseScope(), middleware.RequireWarehouseRole(middleware.RolePackers...))
			{
				warehousePacking.GET("/:order_id", handlers.GetPackingTask)
				warehousePacking.PUT("/:order_id/start", handlers.StartPacking)
				warehousePacking.PUT("/:order_id/complete", handlers.CompletePacking)
			}

			warehouseLocations := warehouse.Group("")
			warehouseLocations.Use(middleware.AuthMiddleware(), middleware.WarehouseStaffOnly(), middleware.InjectWarehouseScope())
			{
				warehouseLocations.GET("/zones", middleware.RequireWarehouseRole(middleware.RoleAnyStaff...), handlers.GetWarehouseZones)
				warehouseLocations.POST("/zones", middleware.RequireWarehouseRole(middleware.RoleInventoryOps...), handlers.CreateWarehouseZone)
				warehouseLocations.GET("/zones/:zone_id/racks", middleware.RequireWarehouseRole(middleware.RoleAnyStaff...), handlers.GetZoneRacks)
				warehouseLocations.POST("/zones/:zone_id/racks", middleware.RequireWarehouseRole(middleware.RoleInventoryOps...), handlers.CreateRack)
				warehouseLocations.GET("/racks/:rack_id/bins", middleware.RequireWarehouseRole(middleware.RoleAnyStaff...), handlers.GetRackBins)
				warehouseLocations.POST("/racks/:rack_id/bins", middleware.RequireWarehouseRole(middleware.RoleInventoryOps...), handlers.CreateBin)
				warehouseLocations.PUT("/inventory/:product_id/bin", middleware.RequireWarehouseRole(middleware.RoleInventoryOps...), handlers.AssignProductBin)
				warehouseLocations.GET("/inventory/:product_id", middleware.RequireWarehouseRole(middleware.RoleAnyStaff...), handlers.GetProductInventory)
				warehouseLocations.DELETE("/zones/:zone_id", middleware.RequireWarehouseRole(middleware.RoleManagement...), handlers.DeleteZone)
				warehouseLocations.DELETE("/racks/:rack_id", middleware.RequireWarehouseRole(middleware.RoleManagement...), handlers.DeleteRack)
				warehouseLocations.DELETE("/bins/:bin_id", middleware.RequireWarehouseRole(middleware.RoleManagement...), handlers.DeleteBin)
				warehouseLocations.POST("/inventory/:product_id/adjust", middleware.RequireWarehouseRole(middleware.RoleInventoryOps...), handlers.AdjustStock)
				warehouseLocations.GET("/stock-movements", middleware.RequireWarehouseRole(middleware.RoleAnyStaff...), handlers.GetStockMovements)
			}

			warehouseExceptions := warehouse.Group("/exceptions")
			warehouseExceptions.Use(middleware.AuthMiddleware(), middleware.WarehouseStaffOnly(), middleware.InjectWarehouseScope())
			{
				warehouseExceptions.GET("", middleware.RequireWarehouseRole(middleware.RoleAnyStaff...), handlers.GetWarehouseExceptions)
				warehouseExceptions.GET("/:id", middleware.RequireWarehouseRole(middleware.RoleAnyStaff...), handlers.GetWarehouseException)
				warehouseExceptions.PUT("/:id", middleware.RequireWarehouseRole(middleware.RoleManagement...), handlers.UpdateWarehouseException)
			}

			warehouse.GET("/audit-logs", middleware.AuthMiddleware(), middleware.WarehouseStaffOnly(), middleware.InjectWarehouseScope(), middleware.RequireWarehouseRole(middleware.RoleManagement...), handlers.GetWarehouseAuditLogs)
			warehouse.GET("/staff", middleware.AuthMiddleware(), middleware.WarehouseStaffOnly(), middleware.InjectWarehouseScope(), middleware.RequireWarehouseRole(middleware.RoleManagement...), handlers.GetWarehouseStaffOverview)
			warehouse.GET("/inventory", middleware.AuthMiddleware(), middleware.WarehouseStaffOnly(), middleware.InjectWarehouseScope(), middleware.RequireWarehouseRole(middleware.RoleAnyStaff...), handlers.GetWarehouseInventory)

			warehouseReceiving := warehouse.Group("/receiving")
			warehouseReceiving.Use(middleware.AuthMiddleware(), middleware.WarehouseStaffOnly(), middleware.InjectWarehouseScope(), middleware.RequireWarehouseRole(middleware.RoleInventoryOps...))
			{
				warehouseReceiving.POST("", handlers.CreateReceiving)
				warehouseReceiving.GET("", handlers.GetWarehouseReceivings)
				warehouseReceiving.GET("/:id", handlers.GetReceiving)
				warehouseReceiving.PUT("/:id/receive", handlers.MarkReceiving)
				warehouseReceiving.PUT("/:id/qc", handlers.QCReceiving)
				warehouseReceiving.PUT("/:id/putaway", handlers.PutAwayReceiving)
			}

			warehouseBatches := warehouse.Group("/batches")
			warehouseBatches.Use(middleware.AuthMiddleware(), middleware.WarehouseStaffOnly(), middleware.InjectWarehouseScope())
			{
				warehouseBatches.POST("", middleware.RequireWarehouseRole(middleware.RoleInventoryOps...), handlers.CreateBatch)
				warehouseBatches.GET("", middleware.RequireWarehouseRole(middleware.RoleAnyStaff...), handlers.GetWarehouseBatches)
				warehouseBatches.GET("/expiring", middleware.RequireWarehouseRole(middleware.RoleAnyStaff...), handlers.GetExpiringBatches)
				warehouseBatches.PUT("/:id/quantity", middleware.RequireWarehouseRole(middleware.RoleInventoryOps...), handlers.AdjustBatchQuantity)
				warehouseBatches.DELETE("/:id", middleware.RequireWarehouseRole(middleware.RoleManagement...), handlers.DeleteBatch)
			}

			warehouse.GET("/staff/performance", middleware.AuthMiddleware(), middleware.WarehouseStaffOnly(), middleware.InjectWarehouseScope(), middleware.RequireWarehouseRole(middleware.RoleManagement...), handlers.GetWarehouseStaffPerformance)
			warehouse.GET("/staff/performance/me", middleware.AuthMiddleware(), middleware.WarehouseStaffOnly(), middleware.InjectWarehouseScope(), middleware.RequireWarehouseRole(middleware.RoleAnyStaff...), handlers.GetMyPerformance)

			warehouse.GET("/dashboard", middleware.AuthMiddleware(), middleware.WarehouseStaffOnly(), middleware.InjectWarehouseScope(), middleware.RequireWarehouseRole(middleware.RoleManagement...), handlers.GetWarehouseDashboard)

			warehouseNotifications := warehouse.Group("/notifications")
			warehouseNotifications.Use(middleware.AuthMiddleware(), middleware.WarehouseStaffOnly(), middleware.InjectWarehouseScope())
			{
				warehouseNotifications.GET("", middleware.RequireWarehouseRole(middleware.RoleAnyStaff...), handlers.GetWarehouseNotifications)
				warehouseNotifications.PUT("/:id/read", middleware.RequireWarehouseRole(middleware.RoleAnyStaff...), handlers.MarkNotificationRead)
				warehouseNotifications.PUT("/read-all", middleware.RequireWarehouseRole(middleware.RoleAnyStaff...), handlers.MarkAllNotificationsRead)
			}
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
			orders.GET("/:id/invoice", handlers.GetMyOrderInvoice)
			orders.GET("/:id/invoice/pdf", handlers.GetMyOrderInvoicePDF)
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
				admin.GET("/audit-logs", handlers.GetAuditLogs)
				admin.POST("/notifications/broadcast", handlers.BroadcastNotification)
				adminPayments := admin.Group("/payments")
{
adminPayments.GET("", handlers.GetAdminPayments)
adminPayments.GET("/reconciliation", handlers.GetAdminPaymentReconciliation)
adminPayments.GET("/:order_id", handlers.GetAdminPaymentDetail)
adminPayments.PUT("/:order_id/status", handlers.UpdateAdminPaymentStatus)
}
adminOffers := admin.Group("/offers")
				{
					adminOffers.POST("", handlers.CreateOffer)
					adminOffers.GET("", handlers.GetOffers)
					adminOffers.PUT("/:id/status", handlers.UpdateOfferStatus)
					adminOffers.DELETE("/:id", handlers.DeleteOffer)
				}
				adminBanners := admin.Group("/banners")
				{
					adminBanners.POST("", handlers.CreateBanner)
					adminBanners.GET("", handlers.GetBanners)
					adminBanners.PUT("/:id", handlers.UpdateBanner)
					adminBanners.DELETE("/:id", handlers.DeleteBanner)
				}
				adminZones := admin.Group("/delivery-zones")
				{
					adminZones.POST("", handlers.CreateDeliveryZone)
					adminZones.GET("", handlers.GetDeliveryZones)
					adminZones.PUT("/:id", handlers.UpdateDeliveryZone)
					adminZones.DELETE("/:id", handlers.DeleteDeliveryZone)
				}
				adminSupport := admin.Group("/support")
				{
					adminSupport.GET("/tickets", handlers.GetAllTickets)
					adminSupport.GET("/tickets/:id/messages", handlers.GetTicketMessagesAdmin)
					adminSupport.POST("/tickets/:id/messages", handlers.AdminReplyToTicket)
					adminSupport.PUT("/tickets/:id/status", handlers.UpdateTicketStatus)
				}
				adminCustomers := admin.Group("/customers")
				{
					adminCustomers.GET("", handlers.GetCustomers)
					adminCustomers.GET("/:id", handlers.GetCustomerByID)
					adminCustomers.PUT("/:id/block", middleware.RequirePermission(middleware.PermBlockCustomer), handlers.BlockCustomer)
					adminCustomers.PUT("/:id/unblock", middleware.RequirePermission(middleware.PermBlockCustomer), handlers.UnblockCustomer)
				}

			adminCategories := admin.Group("/categories")
			{
				adminCategories.POST("", handlers.CreateCategory)
				adminCategories.PUT("/:id", handlers.UpdateCategory)
				adminCategories.DELETE("/:id", handlers.DeleteCategory)
			}

			adminProducts := admin.Group("/products")
			{
				adminProducts.POST("", handlers.CreateProduct)
				adminProducts.PUT("/:id", middleware.RequirePermission(middleware.PermEditPrice), handlers.UpdateProduct)
				adminProducts.DELETE("/:id", handlers.DeleteProduct)
				adminProducts.PUT("/:id/inventory", handlers.UpdateInventory)
				adminProducts.POST("/:id/barcode", handlers.GenerateProductBarcode)
			}

				admin.GET("/inventory", handlers.GetInventoryOverview)
				adminSettings := admin.Group("/settings")
				{
					adminSettings.GET("", handlers.GetSettings)
					adminSettings.PUT("", middleware.RequirePermission(middleware.PermManageSettings), handlers.UpdateSettings)
				}

				adminStaff := admin.Group("/staff")
				{
					adminStaff.GET("", handlers.GetAdminStaff)
					adminStaff.PUT("/:id/role", middleware.RequirePermission(middleware.PermManageStaff), handlers.UpdateStaffRole)
				}


			adminOrders := admin.Group("/orders")
			{
				adminOrders.GET("", handlers.GetAllOrders) // ?status=&page=&limit=
				adminOrders.PUT("/:id/status", handlers.UpdateOrderStatus)
			}

			adminInvoices := admin.Group("/invoices")
			{
				adminInvoices.GET("", handlers.SearchInvoices)
				adminInvoices.GET("/:id", handlers.GetAdminInvoiceByID)
				adminInvoices.GET("/:id/pdf", handlers.GetAdminInvoicePDF)
			}

			adminReturns := admin.Group("/returns")
			{
				adminReturns.GET("", handlers.GetReturns)
				adminReturns.PUT("/:id/approve", middleware.RequirePermission(middleware.PermApproveRefund), handlers.ApproveReturn)
				adminReturns.PUT("/:id/reject", handlers.RejectReturn)
			}

			adminCoupons := admin.Group("/coupons")
			{
				adminCoupons.POST("", middleware.RequirePermission(middleware.PermDeleteCoupon), handlers.CreateCoupon)
				adminCoupons.GET("", handlers.GetCoupons)
				adminCoupons.PUT("/:id/status", middleware.RequirePermission(middleware.PermDeleteCoupon), handlers.UpdateCouponStatus)
				adminCoupons.DELETE("/:id", middleware.RequirePermission(middleware.PermDeleteCoupon), handlers.DeleteCoupon)
			}
			adminAnalytics := admin.Group("/analytics")
			{
				adminAnalytics.GET("/summary", handlers.GetAnalyticsSummary)
				adminAnalytics.GET("/products", handlers.GetProductPerformance)
				adminAnalytics.GET("/dashboard", handlers.GetDashboardOverview)
			}
            adminReports := admin.Group("/reports")
            {
                adminReports.GET("/daily-sales", handlers.GetDailySalesReport)
                adminReports.GET("/daily-sales/export", handlers.ExportDailySalesReport)
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
					adminWarehouses.PUT("/:id/service-area", handlers.SetWarehouseServiceArea)

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


