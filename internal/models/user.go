package models

import "time"

type User struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Name      string    `json:"name" gorm:"size:120;not null"`
	Email     string    `json:"email" gorm:"size:190;uniqueIndex;not null"`
	Password  string    `json:"-" gorm:"size:255;not null"`
	Role      string    `json:"role" gorm:"size:30;not null"`
	Status    string    `json:"status" gorm:"size:30;default:active"` // active, pending, rejected
	CompanyID uint      `json:"company_id" gorm:"index;not null"`
	CreatedBy uint      `json:"created_by" gorm:"index"` // Admin who created this user
	CreatedAt time.Time `json:"created_at"`
}
