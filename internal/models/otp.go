package models

import (
	"time"
)

type OTPPurpose string

const (
	OTPPurposeRegistration   OTPPurpose = "REGISTRATION"
	OTPPurposeLogin          OTPPurpose = "LOGIN"
	OTPPurposeForgotPassword OTPPurpose = "FORGOT_PASSWORD"
)

type OTPCode struct {
	ID         string     `gorm:"type:varchar(36);primaryKey" json:"id"`
	Identifier string     `gorm:"type:varchar(100);index;not null" json:"identifier"` // Username or Phone Number
	Code       string     `gorm:"type:varchar(6);not null" json:"code"`
	Purpose    OTPPurpose `gorm:"type:varchar(20);not null" json:"purpose"`
	ExpiresAt  time.Time  `gorm:"not null" json:"expires_at"`
	IsUsed     bool       `gorm:"default:false" json:"is_used"`
	CreatedAt  time.Time  `json:"created_at"`
}
