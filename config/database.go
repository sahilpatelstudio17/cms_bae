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

	db, err := gorm.Open(postgres.Open(databaseURL), &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
	})
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("failed to access sql DB: %v", err)
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(30 * time.Minute)

	// Drop table to ensure clean schema without FK constraints
	// This is needed because previous versions had FK constraints that conflict with company signups
	db.Migrator().DropTable(&models.ApprovalRequest{})

	if err := db.AutoMigrate(
		&models.Company{},
		&models.User{},
		&models.Employee{},
		&models.Task{},
		&models.Attendance{},
		&models.Subscription{},
		&models.ApprovalRequest{},
		&models.Expense{},
		&models.Sale{},
	); err != nil {
		log.Fatalf("failed to migrate database: %v", err)
	}

	// Update existing employees with NULL roles based on their position
	// This ensures all employees have a role for filtering
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

	return db
}
