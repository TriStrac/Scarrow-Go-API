package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID                 string         `gorm:"type:varchar(36);primaryKey" json:"id"`
	Username           string         `gorm:"type:varchar(100);unique;index;not null" json:"username"`
	Password           string         `gorm:"type:varchar(255);not null" json:"-"`
	GroupID            *string        `gorm:"type:varchar(36);index" json:"group_id"`
	Group              *Group         `gorm:"foreignKey:GroupID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;" json:"-"`
	IsInGroup          bool           `gorm:"default:false" json:"is_user_in_group"`
	IsHead             bool           `gorm:"default:false" json:"is_user_head"`
	IsVerified         bool           `gorm:"default:false" json:"is_verified"`
	SubscriptionStatus string         `gorm:"type:varchar(50);default:'FREE'" json:"subscription_status"`
	IsDeleted          bool           `gorm:"default:false" json:"is_deleted"`
	CreatedAt          time.Time      `json:"created_at"`
	UpdatedAt          time.Time      `json:"updated_at"`
	DeletedAt          gorm.DeletedAt `gorm:"index" json:"-"`

	Profile *UserProfile `gorm:"foreignKey:UserID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"profile"`
	Address *UserAddress `gorm:"foreignKey:UserID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"address"`
}

type UserProfile struct {
	ID          string    `gorm:"type:varchar(36);primaryKey" json:"id"`
	UserID      string    `gorm:"type:varchar(36);not null;index" json:"user_id"`
	FirstName   string    `gorm:"type:varchar(100);not null" json:"first_name"`
	MiddleName  string    `gorm:"type:varchar(100)" json:"middle_name"`
	LastName    string    `gorm:"type:varchar(100);not null" json:"last_name"`
	BirthDate   time.Time `gorm:"type:date" json:"birth_date"`
	PhoneNumber string    `gorm:"type:varchar(20)" json:"phone_number"`
}

type UserAddress struct {
	ID         string `gorm:"type:varchar(36);primaryKey" json:"id"`
	UserID     string `gorm:"type:varchar(36);not null;index" json:"user_id"`
	StreetName string `gorm:"type:varchar(255)" json:"street_name"`
	Baranggay  string `gorm:"type:varchar(100)" json:"baranggay"`
	Town       string `gorm:"type:varchar(100)" json:"town"`
	Province   string `gorm:"type:varchar(100)" json:"province"`
	ZipCode    string `gorm:"type:varchar(10)" json:"zip_code"`
}
