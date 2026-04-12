package service

import (
	"github.com/TriStrac/Scarrow-Go-API/internal/repository"
)

type ReportService interface {
	GetSummary(userID, timeframe string) (map[string]interface{}, error)
}

type reportService struct {
	deviceRepo repository.DeviceRepository
	userRepo   repository.UserRepository
}

func NewReportService(deviceRepo repository.DeviceRepository, userRepo repository.UserRepository) ReportService {
	return &reportService{
		deviceRepo: deviceRepo,
		userRepo:   userRepo,
	}
}

func (s *reportService) GetSummary(userID, timeframe string) (map[string]interface{}, error) {
	// 1. Resolve ownership context (User + Group)
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, err
	}

	ownerIDs := []string{userID}
	if user != nil && user.GroupID != nil {
		ownerIDs = append(ownerIDs, *user.GroupID)
	}

	// 2. Fetch accessible devices
	devices, err := s.deviceRepo.GetDevicesByOwnerIDs(ownerIDs)
	if err != nil {
		return nil, err
	}

	// 3. (Mock Logic) Aggregate Logs across accessible devices
	// In a real scenario, you'd execute a complex GROUP BY query in the repository.
	// For now, we return a structured skeleton that matches what a frontend chart would expect.

	totalDevices := len(devices)
	totalAlerts := 0

	// Mock aggregation
	pestData := map[string]int{
		"LOCUST": 14,
		"RATS":   5,
		"BIRDS":  22,
	}
	
	for _, count := range pestData {
		totalAlerts += count
	}

	summary := map[string]interface{}{
		"timeframe": timeframe,
		"overview": map[string]interface{}{
			"total_devices": totalDevices,
			"total_alerts":  totalAlerts,
		},
		"pest_distribution": pestData,
		"daily_trends": []map[string]interface{}{
			{"date": "2026-04-06", "count": 2},
			{"date": "2026-04-07", "count": 5},
			{"date": "2026-04-08", "count": 1},
			// ...
		},
	}

	return summary, nil
}
