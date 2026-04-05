package service

import (
	"errors"

	"github.com/TriStrac/Scarrow-Go-API/internal/models"
	"github.com/TriStrac/Scarrow-Go-API/internal/repository"
	"github.com/google/uuid"
)

type DeviceService interface {
	CreateDevice(name string, userID string, ownerType string) (*models.Device, error)
	GetAllDevices() ([]models.Device, error)
	GetDeviceByID(id string) (*models.Device, error)
	UpdateDevice(id string, name string, status string) error
	SoftDelete(id string) error

	// Ownership
	AddOwner(deviceID, ownerID, ownerType string) error
	RemoveOwner(deviceID, ownerID, ownerType string) error
	GetOwnersByDeviceID(deviceID string) ([]models.DeviceOwner, error)
	GetMyDevices(userID string) ([]models.Device, error)
	IsOwner(deviceID, userID string) (bool, error)

	// Logging
	CreateLog(deviceID, logType, payload string) error
	GetLogsByDeviceID(deviceID string) ([]models.DeviceLog, error)
}

type deviceService struct {
	repo     repository.DeviceRepository
	userRepo repository.UserRepository
}

func NewDeviceService(repo repository.DeviceRepository, userRepo repository.UserRepository) DeviceService {
	return &deviceService{
		repo:     repo,
		userRepo: userRepo,
	}
}

func (s *deviceService) CreateDevice(name string, userID string, ownerType string) (*models.Device, error) {
	// Determine the actual owner ID based on ownerType
	actualOwnerID := userID

	if ownerType == "GROUP" {
		user, err := s.userRepo.FindByID(userID)
		if err != nil {
			return nil, err
		}
		if user == nil {
			return nil, errors.New("user not found")
		}
		if user.GroupID == nil {
			return nil, errors.New("user does not belong to a group")
		}
		actualOwnerID = *user.GroupID
	}

	device := &models.Device{
		ID:     uuid.New().String(),
		Name:   name,
		Status: "OFFLINE",
	}

	err := s.repo.CreateDevice(device)
	if err != nil {
		return nil, err
	}

	// Add initial owner
	owner := &models.DeviceOwner{
		DeviceID:  device.ID,
		OwnerID:   actualOwnerID,
		OwnerType: ownerType,
	}

	err = s.repo.AddOwner(owner)
	if err != nil {
		// Rollback device creation if owner add fails
		return nil, err
	}

	return device, nil
}

func (s *deviceService) GetAllDevices() ([]models.Device, error) {
	return s.repo.FindAll()
}

func (s *deviceService) GetDeviceByID(id string) (*models.Device, error) {
	device, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if device == nil {
		return nil, errors.New("device not found")
	}
	return device, nil
}

func (s *deviceService) UpdateDevice(id string, name string, status string) error {
	device, err := s.repo.FindByID(id)
	if err != nil {
		return err
	}
	if device == nil {
		return errors.New("device not found")
	}

	if name != "" {
		device.Name = name
	}
	if status != "" {
		device.Status = status
	}

	return s.repo.UpdateDevice(device)
}

func (s *deviceService) SoftDelete(id string) error {
	return s.repo.SoftDelete(id)
}

func (s *deviceService) AddOwner(deviceID, ownerID, ownerType string) error {
	owner := &models.DeviceOwner{
		DeviceID:  deviceID,
		OwnerID:   ownerID,
		OwnerType: ownerType,
	}
	return s.repo.AddOwner(owner)
}

func (s *deviceService) RemoveOwner(deviceID, ownerID, ownerType string) error {
	return s.repo.RemoveOwner(deviceID, ownerID, ownerType)
}

func (s *deviceService) GetOwnersByDeviceID(deviceID string) ([]models.DeviceOwner, error) {
	return s.repo.GetOwnersByDeviceID(deviceID)
}

func (s *deviceService) GetMyDevices(userID string) ([]models.Device, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, err
	}

	ownerIDs := []string{userID}
	if user != nil && user.GroupID != nil {
		ownerIDs = append(ownerIDs, *user.GroupID)
	}

	return s.repo.GetDevicesByOwnerIDs(ownerIDs)
}

func (s *deviceService) IsOwner(deviceID, userID string) (bool, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return false, err
	}

	ownerIDs := []string{userID}
	if user != nil && user.GroupID != nil {
		ownerIDs = append(ownerIDs, *user.GroupID)
	}

	return s.repo.IsOwner(deviceID, ownerIDs)
}

func (s *deviceService) CreateLog(deviceID, logType, payload string) error {
	log := &models.DeviceLog{
		DeviceID: deviceID,
		LogType:  logType,
		Payload:  payload,
	}
	return s.repo.CreateLog(log)
}

func (s *deviceService) GetLogsByDeviceID(deviceID string) ([]models.DeviceLog, error) {
	return s.repo.GetLogsByDeviceID(deviceID)
}
