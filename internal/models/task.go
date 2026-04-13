package models

import "time"

type Task struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	Title       string    `json:"title" gorm:"size:150;not null"`
	Description string    `json:"description" gorm:"type:text"`
	Status      string    `json:"status" gorm:"size:30;not null;default:pending"`
	AssignedTo  uint      `json:"assigned_to" gorm:"index;not null"`
	CompanyID   uint      `json:"company_id" gorm:"index;not null"`
	CreatedAt   time.Time `json:"created_at"`
}
