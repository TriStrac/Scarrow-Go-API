package models

import (
	"time"

	"gorm.io/gorm"
)

type GroupInvitation struct {
	Code      string         `gorm:"type:varchar(8);primaryKey" json:"code"`
	GroupID   string         `gorm:"type:varchar(36);not null;index" json:"group_id"`
	CreatedBy string         `gorm:"type:varchar(36);not null" json:"created_by"`
	ExpiresAt time.Time      `gorm:"not null" json:"expires_at"`
	CreatedAt time.Time      `json:"created_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	Group   *Group `gorm:"foreignKey:GroupID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"group"`
	Creator *User  `gorm:"foreignKey:CreatedBy;references:ID" json:"creator"`
}
