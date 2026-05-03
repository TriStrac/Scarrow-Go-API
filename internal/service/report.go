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
	devices, err := s.deviceRepo.GetDevicesByUserID(userID)
	if err != nil {
		return nil, err
	}

	pestData, err := s.deviceRepo.GetPestDistributionByUserID(userID, timeframe)
	if err != nil {
		return nil, err
	}

	totalDevices := len(devices)
	totalAlerts := 0
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
		},
	}

	return summary, nil
}
