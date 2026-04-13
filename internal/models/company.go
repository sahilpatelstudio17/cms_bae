package models

import "time"

type Company struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Name      string    `json:"name" gorm:"size:120;not null"`
	Email     string    `json:"email" gorm:"size:190;uniqueIndex;not null"`
	CreatedAt time.Time `json:"created_at"`
}
