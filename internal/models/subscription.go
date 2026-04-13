package models

import "time"

type Subscription struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	CompanyID uint      `json:"company_id" gorm:"uniqueIndex;not null"`
	Plan      string    `json:"plan" gorm:"size:50;not null"`
	Status    string    `json:"status" gorm:"size:30;not null"`
	CreatedAt time.Time `json:"created_at"`
}
