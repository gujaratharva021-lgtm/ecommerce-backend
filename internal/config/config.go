package config

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port              string
	GinMode           string
	DBHost            string
	DBPort            string
	DBUser            string
	DBPassword        string
	DBName            string
	DBSSLMode         string
	JWTSecret         string
	JWTExpiryHours    string
	RazorpayKeyID     string
	RazorpayKeySecret string
}

var AppConfig *Config

// LoadConfig reads the .env file (if present) and populates AppConfig.
// It does not fail if .env is missing (useful in production where
// real environment variables are injected directly).
func LoadConfig() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, relying on system environment variables")
	}

	cfg := &Config{
		Port:              getEnv("PORT", "8080"),
		GinMode:           getEnv("GIN_MODE", "debug"),
		DBHost:            getEnv("DB_HOST", "localhost"),
		DBPort:            getEnv("DB_PORT", "5432"),
		DBUser:            getEnv("DB_USER", "postgres"),
		DBPassword:        getEnv("DB_PASSWORD", "postgres"),
		DBName:            getEnv("DB_NAME", "ecommerce_db"),
		DBSSLMode:         getEnv("DB_SSLMODE", "disable"),
		JWTSecret:         getEnv("JWT_SECRET", "default_secret_change_me"),
		JWTExpiryHours:    getEnv("JWT_EXPIRY_HOURS", "72"),
		RazorpayKeyID:     getEnv("RAZORPAY_KEY_ID", ""),
		RazorpayKeySecret: getEnv("RAZORPAY_KEY_SECRET", ""),
	}

	if cfg.RazorpayKeyID == "" || cfg.RazorpayKeySecret == "" {
		log.Println("RAZORPAY_KEY_ID / RAZORPAY_KEY_SECRET not set — online payment endpoints will return an error until configured. COD checkout is unaffected.")
	}

	// Refuse to start in production with the placeholder JWT secret — running
	// with it silently would mean anyone could forge valid tokens.
	if cfg.GinMode == "release" && cfg.JWTSecret == "default_secret_change_me" {
		log.Fatal("JWT_SECRET must be set to a strong random value before running with GIN_MODE=release")
	}

	AppConfig = cfg
	return cfg
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

// DSN builds the PostgreSQL connection string for GORM.
func (c *Config) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.DBHost, c.DBPort, c.DBUser, c.DBPassword, c.DBName, c.DBSSLMode,
	)
}
