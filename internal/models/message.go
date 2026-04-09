package models

import (
	"time"

	"gorm.io/gorm"
)

type MessageThread struct {
	ID        string         `gorm:"type:varchar(36);primaryKey" json:"id"`
	UserA_ID  string         `gorm:"type:varchar(36);not null;index" json:"user_a_id"`
	UserB_ID  string         `gorm:"type:varchar(36);not null;index" json:"user_b_id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	UserA    *User     `gorm:"foreignKey:UserA_ID;references:ID" json:"user_a"`
	UserB    *User     `gorm:"foreignKey:UserB_ID;references:ID" json:"user_b"`
	Messages []Message `gorm:"foreignKey:ThreadID;references:ID" json:"messages,omitempty"`
}

type Message struct {
	ID        string         `gorm:"type:varchar(36);primaryKey" json:"id"`
	ThreadID  string         `gorm:"type:varchar(36);not null;index" json:"thread_id"`
	SenderID  string         `gorm:"type:varchar(36);not null" json:"sender_id"`
	Content   string         `gorm:"type:text;not null" json:"content"`
	IsRead    bool           `gorm:"default:false" json:"is_read"`
	CreatedAt time.Time      `json:"created_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	Thread *MessageThread `gorm:"foreignKey:ThreadID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"-"`
	Sender *User          `gorm:"foreignKey:SenderID;references:ID" json:"sender"`
}
