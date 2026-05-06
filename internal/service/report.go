package service

import (
	"time"

	"github.com/TriStrac/Scarrow-Go-API/internal/repository"
)

type ReportService interface {
	GetSummary(userID, timeframe string) (map[string]interface{}, error)
	GetHubReport(hubID string, startDate, endDate *time.Time) (map[string]interface{}, error)
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

func (s *reportService) GetHubReport(hubID string, startDate, endDate *time.Time) (map[string]interface{}, error) {
	logs, err := s.deviceRepo.GetLogsByHubID(hubID, startDate, endDate, 0, 0)
	if err != nil {
		return nil, err
	}

	type pestGroup struct {
		PestType    string
		Count       int
		Logs        []map[string]interface{}
	}
	grouped := make(map[string]*pestGroup)

	for _, log := range logs {
		key := log.PestType
		if key == "" {
			key = "UNKNOWN"
		}
		if _, ok := grouped[key]; !ok {
			grouped[key] = &pestGroup{PestType: key, Logs: []map[string]interface{}{}}
		}
		grouped[key].Count++
		grouped[key].Logs = append(grouped[key].Logs, map[string]interface{}{
			"created_at":        log.CreatedAt,
			"duration_seconds":  log.DurationSeconds,
		})
	}

	total := 0
	detections := make([]map[string]interface{}, 0, len(grouped))
	for _, g := range grouped {
		total += g.Count
		detections = append(detections, map[string]interface{}{
			"pest_type": g.PestType,
			"count":     g.Count,
			"logs":      g.Logs,
		})
	}

	return map[string]interface{}{
		"total":      total,
		"detections": detections,
	}, nil
}
