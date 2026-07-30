package main

import (
"log"
"time"

"github.com/gin-gonic/gin"
"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/config"
"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/database"
fb "github.com/gujaratharva021-lgtm/ecommerce-backend/internal/firebase"
"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/routes"
"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/services"
"github.com/robfig/cron/v3"
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

// 4. Initialize Firebase (push notifications)
fb.InitFirebase(cfg.FirebaseCredentialsPath)

// 5. Setup daily push notification schedule (IST)
loc, err := time.LoadLocation("Asia/Kolkata")
if err != nil {
log.Printf("Failed to load IST timezone, using server local time: %v", err)
loc = time.Local
}
c := cron.New(cron.WithLocation(loc))
c.AddFunc("0 11 * * *", func() {
services.SendPushToAll("Quick Delivery", "\U0001F6D2 Everything you need, delivered to your doorstep in minutes.")
})
c.AddFunc("0 15 * * *", func() {
services.SendPushToAll("Fresh Deals", "\U0001F389 Fresh deals are waiting for you. Don't miss out!")
})
c.Start()

// 6. Setup Gin router
gin.SetMode(cfg.GinMode)
router := gin.Default()

// 7. Register all routes
routes.SetupRoutes(router)

// 8. Start server
log.Printf("Server starting on port %s", cfg.Port)
if err := router.Run(":" + cfg.Port); err != nil {
log.Fatalf("Failed to start server: %v", err)
}
}
