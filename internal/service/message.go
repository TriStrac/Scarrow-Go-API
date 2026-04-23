package service

import (
	"errors"

	"github.com/TriStrac/Scarrow-Go-API/internal/models"
	"github.com/TriStrac/Scarrow-Go-API/internal/repository"
	"github.com/google/uuid"
)

type MessageService interface {
	SendMessage(senderID, receiverID, content string) (*models.Message, error)
	GetThreads(userID string, limit int, offset int) ([]models.MessageThread, error)
	GetThreadMessages(threadID, userID string, limit int, offset int) (*models.MessageThread, error)
	GetUnreadCount(userID string) (int64, error)
	GetRecentMessages(userID string) ([]models.Message, error)
}

type messageService struct {
	repo      repository.MessageRepository
	userRepo  repository.UserRepository
	groupRepo repository.GroupRepository
}

func NewMessageService(repo repository.MessageRepository, userRepo repository.UserRepository, groupRepo repository.GroupRepository) MessageService {
	return &messageService{repo: repo, userRepo: userRepo, groupRepo: groupRepo}
}

func (s *messageService) SendMessage(senderID, receiverID, content string) (*models.Message, error) {
	if senderID == receiverID {
		return nil, errors.New("cannot send message to yourself")
	}

	// Validate sender and enforce organization restriction
	sender, err := s.userRepo.FindByID(senderID)
	if err != nil {
		return nil, err
	}
	if sender == nil {
		return nil, errors.New("sender not found")
	}
	if sender.GroupID == nil || *sender.GroupID == "" {
		return nil, errors.New("messaging is only available for organization members")
	}

	// Resolve the receiver (Handle frontend bug where display name is passed instead of ID)
	members, err := s.groupRepo.FindMembersByGroupID(*sender.GroupID)
	if err != nil {
		return nil, errors.New("failed to retrieve organization members")
	}

	var actualReceiverID string
	for _, member := range members {
		if member.ID == receiverID {
			actualReceiverID = member.ID
			break
		}
		if member.Username == receiverID {
			actualReceiverID = member.ID
			break
		}
		
		displayName := member.Username
		if member.Profile != nil && member.Profile.FirstName != "" {
			displayName = member.Profile.FirstName + " " + member.Profile.LastName
		}
		if displayName == receiverID {
			actualReceiverID = member.ID
			break
		}
	}

	if actualReceiverID == "" {
		return nil, errors.New("receiver not found in your organization")
	}

	if senderID == actualReceiverID {
		return nil, errors.New("cannot send message to yourself")
	}

	// 1. Find or create thread
	thread, err := s.repo.FindThreadByParticipants(senderID, actualReceiverID)
	if err != nil {
		return nil, err
	}

	if thread == nil {
		thread = &models.MessageThread{
			ID:       uuid.New().String(),
			UserA_ID: senderID,
			UserB_ID: actualReceiverID,
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

func (s *messageService) GetThreads(userID string, limit int, offset int) ([]models.MessageThread, error) {
	return s.repo.FindThreadsByUserID(userID, limit, offset)
}

func (s *messageService) GetThreadMessages(threadID, userID string, limit int, offset int) (*models.MessageThread, error) {
	// Mark messages as read first
	_ = s.repo.MarkThreadAsRead(threadID, userID)
	return s.repo.GetThreadWithMessages(threadID, limit, offset)
}

func (s *messageService) GetUnreadCount(userID string) (int64, error) {
	return s.repo.UnreadCountByUser(userID)
}

func (s *messageService) GetRecentMessages(userID string) ([]models.Message, error) {
	return s.repo.GetRecentMessages(userID, 5) // Last 5 for pre-caching
}
