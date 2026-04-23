package config

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/TriStrac/Scarrow-Go-API/internal/models"
	"github.com/joho/godotenv"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// InitDB initializes the database connection using environment variables
func InitDB() {
	// Load .env file
	_ = godotenv.Load() // Ignore error if .env is missing in Docker environment

	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbName := os.Getenv("DB_NAME")

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		dbUser, dbPassword, dbHost, dbPort, dbName)

	config := &gorm.Config{
		Logger:                                   logger.Default.LogMode(logger.Info),
		DisableForeignKeyConstraintWhenMigrating: true,
	}

	db, err := gorm.Open(mysql.Open(dsn), config)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Configure connection pooling
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("Failed to get database instance: %v", err)
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	log.Println("Database connection established successfully")
	DB = db

	// Auto Migration
	err = db.AutoMigrate(
		&models.Group{}, // Must be created before User because User references Group
		&models.User{},
		&models.UserProfile{},
		&models.UserAddress{},
		&models.PushToken{},
		&models.SubscriptionPlan{},
		&models.UserSubscription{},
		&models.Device{},
		&models.DeviceLog{},
		&models.UserActivityLog{},
		&models.OTPCode{},
		&models.Notification{},
		&models.GroupInvitation{},
		&models.MessageThread{},
		&models.Message{},
	)
	if err != nil {
		log.Fatalf("Failed to perform auto migration: %v", err)
	}
	log.Println("Database auto-migration completed")
}
