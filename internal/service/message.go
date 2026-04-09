package service

import (
	"errors"

	"github.com/TriStrac/Scarrow-Go-API/internal/models"
	"github.com/TriStrac/Scarrow-Go-API/internal/repository"
	"github.com/google/uuid"
)

type MessageService interface {
	SendMessage(senderID, receiverID, content string) (*models.Message, error)
	GetThreads(userID string) ([]models.MessageThread, error)
	GetThreadMessages(threadID, userID string) (*models.MessageThread, error)
	GetUnreadCount(userID string) (int64, error)
	GetRecentMessages(userID string) ([]models.Message, error)
}

type messageService struct {
	repo     repository.MessageRepository
	userRepo repository.UserRepository
}

func NewMessageService(repo repository.MessageRepository, userRepo repository.UserRepository) MessageService {
	return &messageService{repo: repo, userRepo: userRepo}
}

func (s *messageService) SendMessage(senderID, receiverID, content string) (*models.Message, error) {
	if senderID == receiverID {
		return nil, errors.New("cannot send message to yourself")
	}

	// 1. Find or create thread
	thread, err := s.repo.FindThreadByParticipants(senderID, receiverID)
	if err != nil {
		return nil, err
	}

	if thread == nil {
		thread = &models.MessageThread{
			ID:       uuid.New().String(),
			UserA_ID: senderID,
			UserB_ID: receiverID,
		}
		if err := s.repo.CreateThread(thread); err != nil {
			return nil, err
		}
	}

	// 2. Create message
	message := &models.Message{
		ID:       uuid.New().String(),
		ThreadID: thread.ID,
		SenderID: senderID,
		Content:  content,
	}

	if err := s.repo.CreateMessage(message); err != nil {
		return nil, err
	}

	return message, nil
}

func (s *messageService) GetThreads(userID string) ([]models.MessageThread, error) {
	return s.repo.FindThreadsByUserID(userID)
}

func (s *messageService) GetThreadMessages(threadID, userID string) (*models.MessageThread, error) {
	// Mark messages as read first
	_ = s.repo.MarkThreadAsRead(threadID, userID)
	return s.repo.GetThreadWithMessages(threadID, 50) // Limit last 50
}

func (s *messageService) GetUnreadCount(userID string) (int64, error) {
	return s.repo.UnreadCountByUser(userID)
}

func (s *messageService) GetRecentMessages(userID string) ([]models.Message, error) {
	return s.repo.GetRecentMessages(userID, 5) // Last 5 for pre-caching
}
