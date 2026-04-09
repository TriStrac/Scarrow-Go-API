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
	AddOwner(deviceOwner *models.DeviceOwner) error
	RemoveOwner(deviceID, ownerID, ownerType string) error
	RemoveAllOwnersByOwner(ownerID, ownerType string) error
	GetOwnersByDeviceID(deviceID string) ([]models.DeviceOwner, error)
	GetDevicesByOwner(ownerID, ownerType string) ([]models.Device, error)
	GetDevicesByOwnerIDs(ownerIDs []string) ([]models.Device, error)
	IsOwner(deviceID string, ownerIDs []string) (bool, error)
	UnpairNodesByParent(parentID string) error

	// Logging
	CreateLog(log *models.DeviceLog) error
	GetLogsByDeviceID(deviceID string) ([]models.DeviceLog, error)
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
	err := r.db.Where("id = ?", id).First(&device).Error
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
	return r.db.Model(&models.Device{}).Where("id = ?", id).Update("is_deleted", true).Delete(&models.Device{ID: id}).Error
}

func (r *deviceRepository) AddOwner(deviceOwner *models.DeviceOwner) error {
	return r.db.Create(deviceOwner).Error
}

func (r *deviceRepository) RemoveOwner(deviceID, ownerID, ownerType string) error {
	return r.db.Where("device_id = ? AND owner_id = ? AND owner_type = ?", deviceID, ownerID, ownerType).Delete(&models.DeviceOwner{}).Error
}

func (r *deviceRepository) RemoveAllOwnersByOwner(ownerID, ownerType string) error {
	return r.db.Where("owner_id = ? AND owner_type = ?", ownerID, ownerType).Delete(&models.DeviceOwner{}).Error
}

func (r *deviceRepository) GetOwnersByDeviceID(deviceID string) ([]models.DeviceOwner, error) {
	var owners []models.DeviceOwner
	err := r.db.Where("device_id = ?", deviceID).Find(&owners).Error
	return owners, err
}

func (r *deviceRepository) GetDevicesByOwner(ownerID, ownerType string) ([]models.Device, error) {
	var devices []models.Device
	err := r.db.Table("devices").
		Joins("JOIN device_owners ON devices.id = device_owners.device_id").
		Where("device_owners.owner_id = ? AND device_owners.owner_type = ?", ownerID, ownerType).
		Find(&devices).Error
	return devices, err
}

func (r *deviceRepository) GetDevicesByOwnerIDs(ownerIDs []string) ([]models.Device, error) {
	var devices []models.Device
	if len(ownerIDs) == 0 {
		return devices, nil
	}
	err := r.db.Table("devices").
		Joins("JOIN device_owners ON devices.id = device_owners.device_id").
		Where("device_owners.owner_id IN ?", ownerIDs).
		Find(&devices).Error
	return devices, err
}

func (r *deviceRepository) IsOwner(deviceID string, ownerIDs []string) (bool, error) {
	if len(ownerIDs) == 0 {
		return false, nil
	}
	var count int64
	err := r.db.Model(&models.DeviceOwner{}).Where("device_id = ? AND owner_id IN ?", deviceID, ownerIDs).Count(&count).Error
	return count > 0, err
}

func (r *deviceRepository) UnpairNodesByParent(parentID string) error {
	return r.db.Model(&models.Device{}).Where("parent_id = ?", parentID).Update("parent_id", nil).Error
}

func (r *deviceRepository) CreateLog(log *models.DeviceLog) error {
	return r.db.Create(log).Error
}

func (r *deviceRepository) GetLogsByDeviceID(deviceID string) ([]models.DeviceLog, error) {
	var logs []models.DeviceLog
	err := r.db.Where("device_id = ?", deviceID).Order("created_at desc").Find(&logs).Error
	return logs, err
}
