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
		&models.CartReservation{},
		&models.Setting{},
		&models.Offer{},
		&models.Banner{},
		&models.DeliveryZone{},
		&models.SupportTicket{},
		&models.SupportMessage{},
		&models.PickingTask{},
		&models.PickingTaskItem{},
		&models.PackingTask{},
		&models.OrderHandover{},
		&models.WarehouseZone{},
		&models.WarehouseRack{},
		&models.WarehouseBin{},
		&models.StockMovement{},
		&models.WarehouseException{},
		&models.SubstitutionRequest{},
		&models.WarehouseAuditLog{},
		&models.Receiving{},
		&models.Batch{},
		&models.Invoice{},
		&models.InvoiceItem{},
                &models.Expense{},
                &models.Payroll{},
		&models.WarehouseNotification{},
	)
	if err != nil {
		log.Fatalf("Failed to auto-migrate database: %v", err)
	}
	if err := DB.Exec(`ALTER TABLE warehouses ADD COLUMN IF NOT EXISTS service_area geometry(Polygon, 4326)`).Error; err != nil {
		log.Fatalf("Failed to add service_area column: %v", err)
	}
	if err := DB.Exec(`ALTER TABLE products ADD COLUMN IF NOT EXISTS gst_percent DOUBLE PRECISION NOT NULL DEFAULT 0`).Error; err != nil {
		log.Fatalf("Failed to add gst_percent column: %v", err)
	}

	seedDefaultSettings()

	log.Println("Database migration completed")
}


// EnsureProductionSchemaPatches applies small, idempotent schema patches
// that are needed even in production (GIN_MODE=release), where the full
// AutoMigrate() is skipped in favor of versioned migrations. This lets a
// column land immediately on deploy without requiring someone to run the
// migrate CLI by hand first - safe because ADD COLUMN IF NOT EXISTS is a
// no-op once the versioned migration has actually been applied.
func EnsureProductionSchemaPatches() {
if err := DB.Exec(`ALTER TABLE products ADD COLUMN IF NOT EXISTS gst_percent DOUBLE PRECISION NOT NULL DEFAULT 0`).Error; err != nil {
log.Fatalf("Failed to add gst_percent column: %v", err)
}
if err := DB.Exec(`ALTER TABLE orders ADD COLUMN IF NOT EXISTS platform_fee DOUBLE PRECISION NOT NULL DEFAULT 0`).Error; err != nil {
log.Fatalf("Failed to add platform_fee column to orders: %v", err)
}
if err := DB.Exec(`ALTER TABLE invoices ADD COLUMN IF NOT EXISTS platform_fee DOUBLE PRECISION NOT NULL DEFAULT 0`).Error; err != nil {
log.Fatalf("Failed to add platform_fee column to invoices: %v", err)
}
if err := DB.Exec(`ALTER TABLE delivery_partners ADD COLUMN IF NOT EXISTS is_online BOOLEAN NOT NULL DEFAULT false`).Error; err != nil {
log.Fatalf("Failed to add is_online column to delivery_partners: %v", err)
}
if err := DB.Exec(`CREATE TABLE IF NOT EXISTS expenses (
id BIGSERIAL PRIMARY KEY,
amount DOUBLE PRECISION NOT NULL,
category VARCHAR(30) NOT NULL,
expense_date TIMESTAMPTZ NOT NULL,
warehouse_id BIGINT NULL REFERENCES warehouses(id),
note TEXT,
receipt_url TEXT,
added_by_id BIGINT NOT NULL REFERENCES users(id),
created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
)`).Error; err != nil {
log.Fatalf("Failed to create expenses table: %v", err)
}
if err := DB.Exec(`CREATE INDEX IF NOT EXISTS idx_expenses_category ON expenses(category)`).Error; err != nil {
log.Fatalf("Failed to create expenses category index: %v", err)
}
if err := DB.Exec(`CREATE INDEX IF NOT EXISTS idx_expenses_expense_date ON expenses(expense_date)`).Error; err != nil {
log.Fatalf("Failed to create expenses date index: %v", err)
}
if err := DB.Exec(`CREATE INDEX IF NOT EXISTS idx_expenses_warehouse_id ON expenses(warehouse_id)`).Error; err != nil {
log.Fatalf("Failed to create expenses warehouse index: %v", err)
}
if err := DB.Exec(`CREATE TABLE IF NOT EXISTS payrolls (
id BIGSERIAL PRIMARY KEY,
staff_id BIGINT NOT NULL REFERENCES warehouse_staffs(id),
amount DOUBLE PRECISION NOT NULL,
month INTEGER NOT NULL,
year INTEGER NOT NULL,
status VARCHAR(20) NOT NULL DEFAULT 'pending',
payment_method VARCHAR(20),
note TEXT,
paid_by_id BIGINT NULL REFERENCES users(id),
paid_at TIMESTAMPTZ NULL,
created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
)`).Error; err != nil {
log.Fatalf("Failed to create payrolls table: %v", err)
}
if err := DB.Exec(`CREATE INDEX IF NOT EXISTS idx_payrolls_staff_id ON payrolls(staff_id)`).Error; err != nil {
log.Fatalf("Failed to create payrolls staff index: %v", err)
}
if err := DB.Exec(`CREATE INDEX IF NOT EXISTS idx_payrolls_status ON payrolls(status)`).Error; err != nil {
log.Fatalf("Failed to create payrolls status index: %v", err)
}
if err := DB.Exec(`CREATE INDEX IF NOT EXISTS idx_payrolls_month_year ON payrolls(month, year)`).Error; err != nil {
log.Fatalf("Failed to create payrolls month_year index: %v", err)
}
if err := DB.Exec(`ALTER TABLE products ADD COLUMN IF NOT EXISTS cost_price DOUBLE PRECISION NOT NULL DEFAULT 0`).Error; err != nil {
log.Fatalf("Failed to add cost_price column to products: %v", err)
}
seedDefaultSettings()
}
