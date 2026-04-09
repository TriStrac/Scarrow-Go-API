package repository

import (
	"time"

	"github.com/TriStrac/Scarrow-Go-API/internal/models"
	"gorm.io/gorm"
)

type OTPRepository interface {
	CreateOTP(otp *models.OTPCode) error
	GetLatestOTP(identifier string, purpose models.OTPPurpose) (*models.OTPCode, error)
	MarkAsUsed(id string) error
	CountRecentOTPs(identifier string, duration time.Duration) (int64, error)
}

type otpRepository struct {
	db *gorm.DB
}

func NewOTPRepository(db *gorm.DB) OTPRepository {
	return &otpRepository{db: db}
}

func (r *otpRepository) CreateOTP(otp *models.OTPCode) error {
	return r.db.Create(otp).Error
}

func (r *otpRepository) GetLatestOTP(identifier string, purpose models.OTPPurpose) (*models.OTPCode, error) {
	var otp models.OTPCode
	err := r.db.Where("identifier = ? AND purpose = ? AND is_used = ?", identifier, purpose, false).
		Order("created_at DESC").
		First(&otp).Error
	if err != nil {
		return nil, err
	}
	return &otp, nil
}

func (r *otpRepository) MarkAsUsed(id string) error {
	return r.db.Model(&models.OTPCode{}).Where("id = ?", id).Update("is_used", true).Error
}

func (r *otpRepository) CountRecentOTPs(identifier string, duration time.Duration) (int64, error) {
	var count int64
	since := time.Now().Add(-duration)
	err := r.db.Model(&models.OTPCode{}).
		Where("identifier = ? AND created_at > ?", identifier, since).
		Count(&count).Error
	return count, err
}
