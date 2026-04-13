package migrations

import (
	"gorm.io/gorm"
)

type Employee struct {
	Role string `gorm:"size:50;null"`
}

func AddRoleToEmployees(db *gorm.DB) error {
	// Add role column if it doesn't exist
	if err := db.Migrator().AddColumn(&Employee{}, "role"); err != nil {
		return err
	}

	// Update existing employees with NULL roles based on their position
	// Map positions to roles
	positionToRoleMap := map[string]string{
		"Salesman":    "salesman",
		"Developer":   "developer",
		"Staff":       "staff",
		"Manager":     "manager",
		"Employee":    "employee",
		"Admin":       "admin",
		"Super Admin": "super_admin",
	}

	for position, role := range positionToRoleMap {
		if err := db.Model(&Employee{}).
			Where("role IS NULL AND position = ?", position).
			Update("role", role).Error; err != nil {
			return err
		}
	}

	// For any remaining NULL roles, set to 'employee'
	if err := db.Model(&Employee{}).
		Where("role IS NULL").
		Update("role", "employee").Error; err != nil {
		return err
	}

	return nil
}
