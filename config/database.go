package config

import (
	"log"
	"time"

	"cms/internal/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func NewDatabase(databaseURL string, isProduction bool) *gorm.DB {
	logLevel := logger.Info
	if isProduction {
		logLevel = logger.Warn
	}

	// ✅ Debug log
	log.Println("🚀 Connecting to database...")

	// ✅ SSL FIX (Railway required)
	db, err := gorm.Open(postgres.Open(databaseURL+"?sslmode=require"), &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
	})
	if err != nil {
		log.Fatalf("❌ failed to connect to database: %v", err)
	}

	// ✅ Get SQL DB
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("❌ failed to access sql DB: %v", err)
	}

	// ✅ Connection pool config
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(30 * time.Minute)

	log.Println("✅ Database connected")

	// ⚠️ TEMP: Comment this if crash happens
	log.Println("⚠️ Dropping ApprovalRequest table (if exists)")
	err = db.Migrator().DropTable(&models.ApprovalRequest{})
	if err != nil {
		log.Println("⚠️ DropTable warning:", err)
	}

	// ✅ Run migrations
	log.Println("🚀 Running AutoMigrate...")

	err = db.AutoMigrate(
		&models.Company{},
		&models.User{},
		&models.Employee{},
		&models.Task{},
		&models.Attendance{},
		&models.Subscription{},
		&models.ApprovalRequest{},
		&models.Expense{},
		&models.Sale{},
	)
	if err != nil {
		log.Fatalf("❌ failed to migrate database: %v", err)
	}

	log.Println("✅ Migrations completed")

	// ✅ Safe updates (ignore error, no crash)
	log.Println("🚀 Updating employee roles...")

	db.Model(&models.Employee{}).
		Where("role IS NULL AND position = ?", "Salesman").
		Update("role", "salesman")

	db.Model(&models.Employee{}).
		Where("role IS NULL AND position = ?", "Developer").
		Update("role", "developer")

	db.Model(&models.Employee{}).
		Where("role IS NULL AND position = ?", "Staff").
		Update("role", "staff")

	db.Model(&models.Employee{}).
		Where("role IS NULL AND position = ?", "Manager").
		Update("role", "manager")

	db.Model(&models.Employee{}).
		Where("role IS NULL").
		Update("role", "employee")

	log.Println("✅ Database setup complete")

	return db
}
