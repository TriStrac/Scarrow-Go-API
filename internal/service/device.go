package service

import (
	"errors"

	"github.com/TriStrac/Scarrow-Go-API/internal/models"
	"github.com/TriStrac/Scarrow-Go-API/internal/repository"
	"github.com/TriStrac/Scarrow-Go-API/pkg/utils"
	"github.com/google/uuid"
)

type DeviceService interface {
	CreateDevice(name string, userID string, deviceType models.DeviceType, parentID *string) (*models.Device, error)
	RegisterHub(name string, userID string, lat, lng *float64) (*models.Device, error)
	RegisterNode(name string, userID string, hubID string, nodeType string) (*models.Device, error)
	GetAllDevices() ([]models.Device, error)
	GetDeviceByID(id string) (*models.Device, error)
	UpdateDevice(id string, name string, status string) error
	SoftDelete(id string) error

	// Ownership
	GetMyDevices(userID string) ([]models.Device, error)
	IsOwner(deviceID, userID string) (bool, error)

	// Logging
	CreateLog(deviceID, logType, payload, pestType string, freq float64, duration int) error
	GetLogsByDeviceID(deviceID string, limit int, offset int) ([]models.DeviceLog, error)
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

func (s *deviceService) CreateDevice(name string, userID string, deviceType models.DeviceType, parentID *string) (*models.Device, error) {
	device := &models.Device{
		ID:       uuid.New().String(),
		Name:     name,
		UserID:   userID,
		Type:     deviceType,
		ParentID: parentID,
		Status:   "OFFLINE",
	}

	err := s.repo.CreateDevice(device)
	if err != nil {
		return nil, err
	}

	return device, nil
}

func (s *deviceService) RegisterHub(name string, userID string, lat, lng *float64) (*models.Device, error) {
	device := &models.Device{
		ID:       utils.GenerateDeviceID("HUB"),
		Name:     name,
		UserID:   userID,
		Type:     models.DeviceTypeCentral,
		Status:   "active",
		Secret:   utils.GenerateSecret(32),
		Lat:      lat,
		Lng:      lng,
	}

	err := s.repo.CreateDevice(device)
	if err != nil {
		return nil, err
	}

	return device, nil
}

func (s *deviceService) RegisterNode(name string, userID string, hubID string, nodeType string) (*models.Device, error) {
	// Verify hub exists and caller is owner
	hub, err := s.GetDeviceByID(hubID)
	if err != nil {
		return nil, err
	}
	isOwner, err := s.IsOwner(hub.ID, userID)
	if err != nil {
		return nil, err
	}
	if !isOwner {
		return nil, errors.New("unauthorized: not an owner of the specified hub")
	}

	device := &models.Device{
		ID:       utils.GenerateDeviceID("NODE"),
		Name:     name,
		UserID:   userID,
		Type:     models.DeviceTypeNode,
		Status:   "active",
		Secret:   utils.GenerateSecret(32),
		ParentID: &hub.ID,
		NodeType: nodeType,
	}

	err = s.repo.CreateDevice(device)
	if err != nil {
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
	// Unpair nodes if this is a central device
	_ = s.repo.UnpairNodesByParent(id)
	return s.repo.SoftDelete(id)
}

func (s *deviceService) GetMyDevices(userID string) ([]models.Device, error) {
	return s.repo.GetDevicesByUserID(userID)
}

func (s *deviceService) IsOwner(deviceID, userID string) (bool, error) {
	return s.repo.IsOwner(deviceID, userID)
}

func (s *deviceService) CreateLog(deviceID, logType, payload, pestType string, freq float64, duration int) error {
	log := &models.DeviceLog{
		ID:              uuid.New().String(),
		DeviceID:        deviceID,
		LogType:         logType,
		Payload:         payload,
		PestType:        pestType,
		FrequencyHz:     freq,
		DurationSeconds: duration,
	}
	return s.repo.CreateLog(log)
}

func (s *deviceService) GetLogsByDeviceID(deviceID string, limit int, offset int) ([]models.DeviceLog, error) {
	return s.repo.GetLogsByDeviceID(deviceID, limit, offset)
}
