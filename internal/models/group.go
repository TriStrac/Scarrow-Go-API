package models

import (
	"time"

	"gorm.io/gorm"
)

type Group struct {
	ID        string         `gorm:"type:varchar(36);primaryKey" json:"id"`
	Name      string         `gorm:"type:varchar(100);unique;not null" json:"name"`
	OwnerID   string         `gorm:"type:varchar(36);not null;index" json:"owner_id"`
	IsDeleted bool           `gorm:"default:false" json:"is_deleted"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	Owner     *User          `gorm:"foreignKey:OwnerID" json:"owner"`
	Members   []User         `gorm:"many2many:group_members;" json:"members"`
}

type GroupMember struct {
	GroupID  string    `gorm:"type:varchar(36);primaryKey" json:"group_id"`
	UserID   string    `gorm:"type:varchar(36);primaryKey" json:"user_id"`
	JoinedAt time.Time `gorm:"autoCreateTime" json:"joined_at"`
}
