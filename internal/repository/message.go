package repository

import (
	"errors"

	"github.com/TriStrac/Scarrow-Go-API/internal/models"
	"gorm.io/gorm"
)

type MessageRepository interface {
	CreateThread(thread *models.MessageThread) error
	FindThreadByParticipants(userA, userB string) (*models.MessageThread, error)
	FindThreadsByUserID(userID string) ([]models.MessageThread, error)
	GetThreadWithMessages(threadID string, limit int) (*models.MessageThread, error)

	CreateMessage(message *models.Message) error
	MarkThreadAsRead(threadID, userID string) error
	UnreadCountByUser(userID string) (int64, error)
	GetRecentMessages(userID string, limit int) ([]models.Message, error)
}

type messageRepository struct {
	db *gorm.DB
}

func NewMessageRepository(db *gorm.DB) MessageRepository {
	return &messageRepository{db: db}
}

func (r *messageRepository) CreateThread(thread *models.MessageThread) error {
	return r.db.Create(thread).Error
}

func (r *messageRepository) FindThreadByParticipants(userA, userB string) (*models.MessageThread, error) {
	var thread models.MessageThread
	err := r.db.Where("(user_a_id = ? AND user_b_id = ?) OR (user_a_id = ? AND user_b_id = ?)",
		userA, userB, userB, userA).First(&thread).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &thread, nil
}

func (r *messageRepository) FindThreadsByUserID(userID string) ([]models.MessageThread, error) {
	var threads []models.MessageThread
	err := r.db.Preload("UserA").Preload("UserB").
		Where("user_a_id = ? OR user_b_id = ?", userID, userID).
		Order("updated_at desc").Find(&threads).Error
	return threads, err
}

func (r *messageRepository) GetThreadWithMessages(threadID string, limit int) (*models.MessageThread, error) {
	var thread models.MessageThread
	err := r.db.Preload("UserA").Preload("UserB").
		Preload("Messages", func(db *gorm.DB) *gorm.DB {
			return db.Order("created_at desc").Limit(limit)
		}).
		Where("id = ?", threadID).First(&thread).Error
	if err != nil {
		return nil, err
	}
	return &thread, nil
}

func (r *messageRepository) CreateMessage(message *models.Message) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(message).Error; err != nil {
			return err
		}
		// Update thread's updated_at
		return tx.Model(&models.MessageThread{}).Where("id = ?", message.ThreadID).Update("updated_at", gorm.Expr("NOW()")).Error
	})
}

func (r *messageRepository) MarkThreadAsRead(threadID, userID string) error {
	return r.db.Model(&models.Message{}).
		Where("thread_id = ? AND sender_id != ? AND is_read = ?", threadID, userID, false).
		Update("is_read", true).Error
}

func (r *messageRepository) UnreadCountByUser(userID string) (int64, error) {
	var count int64
	err := r.db.Model(&models.Message{}).
		Joins("JOIN message_threads ON messages.thread_id = message_threads.id").
		Where("(message_threads.user_a_id = ? OR message_threads.user_b_id = ?) AND messages.sender_id != ? AND messages.is_read = ?",
			userID, userID, userID, false).
		Count(&count).Error
	return count, err
}

func (r *messageRepository) GetRecentMessages(userID string, limit int) ([]models.Message, error) {
	var messages []models.Message
	err := r.db.Preload("Sender").
		Joins("JOIN message_threads ON messages.thread_id = message_threads.id").
		Where("(message_threads.user_a_id = ? OR message_threads.user_b_id = ?)", userID, userID).
		Order("messages.created_at desc").Limit(limit).Find(&messages).Error
	return messages, err
}
