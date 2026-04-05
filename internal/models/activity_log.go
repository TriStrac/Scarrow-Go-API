package models

import (
	"time"
)

type UserActivityLog struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    string    `gorm:"type:varchar(36);not null;index" json:"user_id"`
	Action    string    `gorm:"type:varchar(255);not null" json:"action"`
	Module    string    `gorm:"type:varchar(100)" json:"module"`
	CreatedAt time.Time `json:"created_at"`
}
