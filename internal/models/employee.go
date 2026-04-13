package models

import "time"

type Employee struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Name      string    `json:"name" gorm:"size:120;not null"`
	Position  string    `json:"position" gorm:"size:100;not null"`
	Role      string    `json:"role" gorm:"size:50;null"` // admin, super_admin, manager, salesman, developer, staff, employee
	Salary    float64   `json:"salary" gorm:"not null"`
	CompanyID uint      `json:"company_id" gorm:"index;not null"`
	UserID    *uint     `json:"user_id" gorm:"index;null"` // Reference to user account if created with user
	CreatedAt time.Time `json:"created_at"`
}
