package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/handlers"
	"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/middleware"
)

// SetupRoutes defines every API endpoint for the application.
func SetupRoutes(router *gin.Engine) {
	router.GET("/health", handlers.HealthCheck)

	api := router.Group("/api/v1")
	{
		// ---- Auth routes (public) ----
		auth := api.Group("/auth")
		{
			auth.POST("/signup", handlers.Signup)
			auth.POST("/login", handlers.Login)
			auth.GET("/me", middleware.AuthMiddleware(), handlers.Me)
		}

		// ---- Product routes (public, structure ready for Day 2) ----
		products := api.Group("/products")
		{
			products.GET("", handlers.GetProducts)
			products.GET("/:id", handlers.GetProductByID)
		}

		// ---- Category routes (public, structure ready for Day 2) ----
		categories := api.Group("/categories")
		{
			categories.GET("", handlers.GetCategories)
		}

		// ---- Cart routes (protected, structure ready for Day 2) ----
		cart := api.Group("/cart")
		cart.Use(middleware.AuthMiddleware())
		{
			cart.GET("", handlers.GetCart)
			cart.POST("", handlers.AddToCart)
		}
	}
}
