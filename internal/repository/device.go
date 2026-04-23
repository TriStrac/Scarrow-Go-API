package repository

import (
	"errors"

	"github.com/TriStrac/Scarrow-Go-API/internal/models"
	"gorm.io/gorm"
)

type DeviceRepository interface {
	CreateDevice(device *models.Device) error
	FindAll() ([]models.Device, error)
	FindByID(id string) (*models.Device, error)
	UpdateDevice(device *models.Device) error
	SoftDelete(id string) error

	// Ownership
	GetDevicesByUserID(userID string) ([]models.Device, error)
	GetDevicesByUserIDs(userIDs []string) ([]models.Device, error)
	IsOwner(deviceID string, userID string) (bool, error)
	UnpairNodesByParent(parentID string) error

	// Logging
	CreateLog(log *models.DeviceLog) error
	GetLogsByDeviceID(deviceID string, limit int, offset int) ([]models.DeviceLog, error)
}

type deviceRepository struct {
	db *gorm.DB
}

func NewDeviceRepository(db *gorm.DB) DeviceRepository {
	return &deviceRepository{db: db}
}

func (r *deviceRepository) CreateDevice(device *models.Device) error {
	return r.db.Create(device).Error
}

func (r *deviceRepository) FindAll() ([]models.Device, error) {
	var devices []models.Device
	err := r.db.Find(&devices).Error
	return devices, err
}

func (r *deviceRepository) FindByID(id string) (*models.Device, error) {
	var device models.Device
	err := r.db.Where("device_id = ?", id).First(&device).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &device, nil
}

func (r *deviceRepository) UpdateDevice(device *models.Device) error {
	return r.db.Save(device).Error
}

func (r *deviceRepository) SoftDelete(id string) error {
	return r.db.Model(&models.Device{}).Where("device_id = ?", id).Update("is_deleted", true).Delete(&models.Device{ID: id}).Error
}

func (r *deviceRepository) GetDevicesByUserID(userID string) ([]models.Device, error) {
	var devices []models.Device
	err := r.db.Where("user_id = ?", userID).Find(&devices).Error
	return devices, err
}

func (r *deviceRepository) GetDevicesByUserIDs(userIDs []string) ([]models.Device, error) {
	var devices []models.Device
	if len(userIDs) == 0 {
		return devices, nil
	}
	err := r.db.Where("user_id IN ?", userIDs).Find(&devices).Error
	return devices, err
}

func (r *deviceRepository) IsOwner(deviceID string, userID string) (bool, error) {
	var count int64
	err := r.db.Model(&models.Device{}).Where("device_id = ? AND user_id = ?", deviceID, userID).Count(&count).Error
	return count > 0, err
}

func (r *deviceRepository) UnpairNodesByParent(parentID string) error {
	return r.db.Model(&models.Device{}).Where("parent_id = ?", parentID).Update("parent_id", nil).Error
}

func (r *deviceRepository) CreateLog(log *models.DeviceLog) error {
	return r.db.Create(log).Error
}

func (r *deviceRepository) GetLogsByDeviceID(deviceID string, limit int, offset int) ([]models.DeviceLog, error) {
	var logs []models.DeviceLog
	query := r.db.Where("device_id = ?", deviceID).Order("created_at desc")
	
	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	err := query.Find(&logs).Error
	return logs, err
}
