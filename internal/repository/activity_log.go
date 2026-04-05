package repository

import (
	"github.com/TriStrac/Scarrow-Go-API/internal/models"
	"gorm.io/gorm"
)

type ActivityLogRepository interface {
	CreateUserActivityLog(log *models.UserActivityLog) error
	GetLogsByUserID(userID string) ([]models.UserActivityLog, error)
	GetAllLogs() ([]models.UserActivityLog, error)
}

type activityLogRepository struct {
	db *gorm.DB
}

func NewActivityLogRepository(db *gorm.DB) ActivityLogRepository {
	return &activityLogRepository{db: db}
}

func (r *activityLogRepository) CreateUserActivityLog(log *models.UserActivityLog) error {
	return r.db.Create(log).Error
}

func (r *activityLogRepository) GetLogsByUserID(userID string) ([]models.UserActivityLog, error) {
	var logs []models.UserActivityLog
	err := r.db.Where("user_id = ?", userID).Order("created_at desc").Find(&logs).Error
	return logs, err
}

func (r *activityLogRepository) GetAllLogs() ([]models.UserActivityLog, error) {
	var logs []models.UserActivityLog
	err := r.db.Order("created_at desc").Find(&logs).Error
	return logs, err
}
