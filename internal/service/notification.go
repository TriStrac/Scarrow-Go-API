package service

import (
	"github.com/TriStrac/Scarrow-Go-API/internal/models"
	"github.com/TriStrac/Scarrow-Go-API/internal/repository"
	"github.com/google/uuid"
)

type NotificationService interface {
	CreateNotification(userID, title, message string) error
	GetNotificationsByUserID(userID string) ([]models.Notification, error)
	MarkAsRead(id string) error
	MarkAllAsRead(userID string) error
}

type notificationService struct {
	repo repository.NotificationRepository
}

func NewNotificationService(repo repository.NotificationRepository) NotificationService {
	return &notificationService{repo: repo}
}

func (s *notificationService) CreateNotification(userID, title, message string) error {
	notification := &models.Notification{
		ID:      uuid.New().String(),
		UserID:  userID,
		Title:   title,
		Message: message,
	}
	return s.repo.Create(notification)
}

func (s *notificationService) GetNotificationsByUserID(userID string) ([]models.Notification, error) {
	return s.repo.FindByUserID(userID)
}

func (s *notificationService) MarkAsRead(id string) error {
	return s.repo.MarkAsRead(id)
}

func (s *notificationService) MarkAllAsRead(userID string) error {
	return s.repo.MarkAllAsRead(userID)
}
