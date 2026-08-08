package database

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/config"
	"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// ConnectDatabase opens a connection to PostgreSQL and stores it in DB.
func ConnectDatabase(cfg *config.Config) {
	// Only log every SQL query in dev mode. In production this slows down
	// every request and floods the logs.
	logLevel := logger.Info
	if cfg.GinMode == gin.ReleaseMode {
		logLevel = logger.Warn
	}

	db, err := gorm.Open(postgres.Open(cfg.DSN()), &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
	})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("Failed to get underlying sql.DB: %v", err)
	}
	// Connection pool tuning - prevents stale/hanging connections on hosted
	// Postgres (common cause of slow requests after idle periods on Render).
	sqlDB.SetMaxOpenConns(10)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)
	sqlDB.SetConnMaxIdleTime(2 * time.Minute)

	DB = db
	log.Println("Database connected successfully")
}

// AutoMigrate creates/updates all tables based on the models.
// This covers: Users, OTPs, Categories, Products, Inventory, Cart,
// CartItems, Addresses, Orders, OrderItems, Payments.
func AutoMigrate() {
	err := DB.AutoMigrate(
		&models.User{},
		&models.OTP{},
		&models.Category{},
		&models.Product{},
		&models.Inventory{},
		&models.Cart{},
		&models.CartItem{},
		&models.Address{},
		&models.Order{},
		&models.OrderItem{},
		&models.Payment{},
		&models.Coupon{},
		&models.OrderCoupon{},
		&models.Notification{},
		&models.DeviceToken{},
		&models.Wishlist{},
		&models.Review{},
		&models.DeliveryPartner{},
		&models.AuditLog{},
		&models.Warehouse{},
		&models.WarehouseStaff{},
		&models.StockTransfer{},
		&models.Wallet{},
		&models.WalletTransaction{},
		&models.ReturnRequest{},
		&models.ReturnRequestItem{},
	)
	if err != nil {
		log.Fatalf("Failed to auto-migrate database: %v", err)
	}
	log.Println("Database migration completed")
}
