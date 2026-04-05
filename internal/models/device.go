package models

import (
	"time"

	"gorm.io/gorm"
)

type Device struct {
	ID        string         `gorm:"type:varchar(36);primaryKey" json:"id"`
	Name      string         `gorm:"type:varchar(100);not null" json:"name"`
	Status    string         `gorm:"type:varchar(50);default:'OFFLINE'" json:"status"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	IsDeleted bool           `gorm:"default:false" json:"is_deleted"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

type DeviceOwner struct {
	DeviceID  string `gorm:"type:varchar(36);primaryKey" json:"device_id"`
	OwnerID   string `gorm:"type:varchar(36);primaryKey" json:"owner_id"`
	OwnerType string `gorm:"type:varchar(20);primaryKey" json:"owner_type"` // 'USER' or 'GROUP'
}

type DeviceLog struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	DeviceID  string    `gorm:"type:varchar(36);not null;index" json:"device_id"`
	LogType   string    `gorm:"type:varchar(50);not null" json:"log_type"`
	Payload   string    `gorm:"type:text" json:"payload"`
	CreatedAt time.Time `json:"created_at"`
}
