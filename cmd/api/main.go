package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
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
	// L-11: AddFunc returns (EntryID, error); both were previously discarded,
	// so a malformed cron spec would silently never run. Now we log the
	// error so a bad schedule is visible instead of a silent no-op.
	if _, err := c.AddFunc("0 11 * * *", func() {
		services.SendPushToAll("Quick Delivery", "\U0001F6D2 Everything you need, delivered to your doorstep in minutes.")
	}); err != nil {
		log.Printf("Failed to schedule 'Quick Delivery' push job: %v", err)
	}
	if _, err := c.AddFunc("0 15 * * *", func() {
		services.SendPushToAll("Fresh Deals", "\U0001F389 Fresh deals are waiting for you. Don't miss out!")
	}); err != nil {
		log.Printf("Failed to schedule 'Fresh Deals' push job: %v", err)
	}
	c.Start()
	defer c.Stop()

	// 6. Setup Gin router
	gin.SetMode(cfg.GinMode)
	router := gin.Default()

	// 7. Register all routes
	routes.SetupRoutes(router)

	// 8. Start server with graceful shutdown (L-11): previously router.Run
	// blocked with no signal handling, so every deploy/restart killed
	// in-flight requests outright. Now SIGINT/SIGTERM trigger an orderly
	// shutdown that lets active requests finish (bounded by a timeout).
	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: router,
	}

	go func() {
		log.Printf("Server starting on port %s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}
	log.Println("Server exited")
}
