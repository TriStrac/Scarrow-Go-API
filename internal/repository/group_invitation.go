package repository

import (
	"errors"
	"time"

	"github.com/TriStrac/Scarrow-Go-API/internal/models"
	"gorm.io/gorm"
)

type GroupInvitationRepository interface {
	Create(invitation *models.GroupInvitation) error
	FindByCode(code string) (*models.GroupInvitation, error)
	DeleteByCode(code string) error
	DeleteExpired() error
	DeleteByGroupID(groupID string) error
}

type groupInvitationRepository struct {
	db *gorm.DB
}

func NewGroupInvitationRepository(db *gorm.DB) GroupInvitationRepository {
	return &groupInvitationRepository{db: db}
}

func (r *groupInvitationRepository) Create(invitation *models.GroupInvitation) error {
	return r.db.Create(invitation).Error
}

func (r *groupInvitationRepository) FindByCode(code string) (*models.GroupInvitation, error) {
	var invitation models.GroupInvitation
	err := r.db.Where("code = ? AND expires_at > ?", code, time.Now()).First(&invitation).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &invitation, nil
}

func (r *groupInvitationRepository) DeleteByCode(code string) error {
	return r.db.Where("code = ?", code).Delete(&models.GroupInvitation{}).Error
}

func (r *groupInvitationRepository) DeleteExpired() error {
	return r.db.Where("expires_at < ?", time.Now()).Delete(&models.GroupInvitation{}).Error
}

func (r *groupInvitationRepository) DeleteByGroupID(groupID string) error {
	return r.db.Where("group_id = ?", groupID).Delete(&models.GroupInvitation{}).Error
}
