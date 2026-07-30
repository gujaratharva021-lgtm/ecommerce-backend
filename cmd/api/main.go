package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/config"
	"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/database"
	"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/routes"
)

func main() {
	// 1. Load configuration from .env
	cfg := config.LoadConfig()

	// 2. Connect to PostgreSQL
	database.ConnectDatabase(cfg)

	// 3. Run auto-migration to create tables (dev/debug only;
	// In production (GIN_MODE=release), schema changes must go through
	// versioned migrations in migrations/ using the migrate CLI.
	if cfg.GinMode != gin.ReleaseMode {
		database.AutoMigrate()
	}

	// 4. Setup Gin router
	gin.SetMode(cfg.GinMode)
	router := gin.Default()

	// 5. Register all routes
	routes.SetupRoutes(router)

	// 6. Start server
	log.Printf("Server starting on port %s", cfg.Port)
	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

