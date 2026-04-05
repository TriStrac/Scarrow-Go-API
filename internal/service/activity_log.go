package service

import (
	"github.com/TriStrac/Scarrow-Go-API/internal/models"
	"github.com/TriStrac/Scarrow-Go-API/internal/repository"
)

type ActivityLogService interface {
	LogActivity(userID string, action string, module string) error
	GetLogsByUserID(userID string) ([]models.UserActivityLog, error)
	GetAllLogs() ([]models.UserActivityLog, error)
}

type activityLogService struct {
	repo repository.ActivityLogRepository
}

func NewActivityLogService(repo repository.ActivityLogRepository) ActivityLogService {
	return &activityLogService{repo: repo}
}

func (s *activityLogService) LogActivity(userID string, action string, module string) error {
	log := &models.UserActivityLog{
		UserID: userID,
		Action: action,
		Module: module,
	}
	return s.repo.CreateUserActivityLog(log)
}

func (s *activityLogService) GetLogsByUserID(userID string) ([]models.UserActivityLog, error) {
	return s.repo.GetLogsByUserID(userID)
}

func (s *activityLogService) GetAllLogs() ([]models.UserActivityLog, error) {
	return s.repo.GetAllLogs()
}
