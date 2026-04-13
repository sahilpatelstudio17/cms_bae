package models

import "time"

type Attendance struct {
	ID         uint       `json:"id" gorm:"primaryKey"`
	EmployeeID uint       `json:"employee_id" gorm:"index;not null"`
	Date       time.Time  `json:"date" gorm:"type:date;not null"`
	InTime     *time.Time `json:"in_time"`   // Check-in time with timestamp
	OutTime    *time.Time `json:"out_time"`  // Check-out time with timestamp
	CompanyID  uint       `json:"company_id" gorm:"index;not null"`
	CreatedAt  time.Time  `json:"created_at"`
	
	// Relations
	Employee *Employee `json:"employee,omitempty" gorm:"foreignKey:EmployeeID;references:ID"`
}
