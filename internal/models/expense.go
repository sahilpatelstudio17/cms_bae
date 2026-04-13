package models

import "time"

type Expense struct {
	ID          uint       `json:"id" gorm:"primaryKey"`
	EmployeeID  uint       `json:"employee_id" gorm:"index;not null"`
	CompanyID   uint       `json:"company_id" gorm:"index;not null"`
	Amount      float64    `json:"amount" gorm:"not null"`
	Category    string     `json:"category" gorm:"size:50;not null"`
	Description string     `json:"description" gorm:"type:text"`
	Status      string     `json:"status" gorm:"size:30;not null;default:pending"`
	ApprovedBy  *uint      `json:"approved_by"`
	ApprovedAt  *time.Time `json:"approved_at"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`

	// Relations
	Employee *Employee `json:"employee,omitempty" gorm:"foreignKey:EmployeeID;references:ID"`
	Company  *Company  `json:"company,omitempty" gorm:"foreignKey:CompanyID;references:ID"`
}
