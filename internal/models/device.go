package models

import (
	"time"

	"gorm.io/gorm"
)

type DeviceType string

const (
	DeviceTypeCentral DeviceType = "CENTRAL"
	DeviceTypeNode    DeviceType = "NODE"
)

type Device struct {
	ID        string         `gorm:"column:device_id;type:varchar(36);primaryKey" json:"id"`
	Name      string         `gorm:"type:varchar(100);not null" json:"name"`
	UserID    string         `gorm:"type:varchar(36);not null;index" json:"owner_id"`
	Type      DeviceType     `gorm:"type:varchar(20);not null;default:'CENTRAL'" json:"device_type"`
	ParentID  *string        `gorm:"type:varchar(36);index" json:"parent_id"`
	Status    string         `gorm:"type:varchar(50);default:'OFFLINE'" json:"status"`
	Secret    string         `gorm:"type:varchar(64)" json:"secret,omitempty"`
	Lat       *float64       `gorm:"type:decimal(10,8)" json:"lat,omitempty"`
	Lng       *float64       `gorm:"type:decimal(11,8)" json:"lng,omitempty"`
	NodeType  string         `gorm:"type:varchar(50)" json:"node_type,omitempty"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	IsDeleted bool           `gorm:"default:false" json:"is_deleted"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	User   *User    `gorm:"foreignKey:UserID;references:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"-"`
	Parent *Device  `gorm:"foreignKey:ParentID;references:DeviceID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;" json:"-"`
	Nodes  []Device `gorm:"foreignKey:ParentID;references:DeviceID" json:"nodes,omitempty"`
}

type DeviceLog struct {
	ID              string    `gorm:"column:log_id;type:varchar(36);primaryKey" json:"id"`
	DeviceID        string    `gorm:"type:varchar(36);not null;index" json:"device_id"`
	LogType         string    `gorm:"type:varchar(50);not null" json:"log_type"`
	PestType        string    `gorm:"type:varchar(50)" json:"pest_type"`
	FrequencyHz     float64   `json:"frequency_hz"`
	DurationSeconds int       `json:"duration_seconds"`
	Payload         string    `gorm:"type:text" json:"payload"`
	CreatedAt       time.Time `json:"created_at"`
}
