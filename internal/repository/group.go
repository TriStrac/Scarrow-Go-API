package repository

import (
	"errors"

	"github.com/TriStrac/Scarrow-Go-API/internal/models"
	"gorm.io/gorm"
)

type GroupRepository interface {
	CreateGroup(group *models.Group) error
	FindAll() ([]models.Group, error)
	FindByOwnerID(ownerID string) ([]models.Group, error)
	FindByID(id string) (*models.Group, error)
	UpdateGroup(group *models.Group) error
	SoftDelete(id string) error
	AddMember(groupMember *models.GroupMember) error
	RemoveMember(groupID, userID string) error
	FindMembersByGroupID(groupID string) ([]models.User, error)
	GroupNameExists(name string) (bool, error)
	IsMember(groupID, userID string) (bool, error)
}

type groupRepository struct {
	db *gorm.DB
}

func NewGroupRepository(db *gorm.DB) GroupRepository {
	return &groupRepository{db: db}
}

func (r *groupRepository) CreateGroup(group *models.Group) error {
	return r.db.Create(group).Error
}

func (r *groupRepository) FindAll() ([]models.Group, error) {
	var groups []models.Group
	err := r.db.Preload("Owner").Find(&groups).Error
	return groups, err
}

func (r *groupRepository) FindByOwnerID(ownerID string) ([]models.Group, error) {
	var groups []models.Group
	err := r.db.Preload("Owner").Where("owner_id = ?", ownerID).Find(&groups).Error
	return groups, err
}

func (r *groupRepository) FindByID(id string) (*models.Group, error) {
	var group models.Group
	err := r.db.Preload("Owner").Where("id = ?", id).First(&group).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &group, nil
}

func (r *groupRepository) UpdateGroup(group *models.Group) error {
	return r.db.Save(group).Error
}

func (r *groupRepository) SoftDelete(id string) error {
	return r.db.Model(&models.Group{}).Where("id = ?", id).Update("is_deleted", true).Delete(&models.Group{ID: id}).Error
}

func (r *groupRepository) AddMember(groupMember *models.GroupMember) error {
	return r.db.Create(groupMember).Error
}

func (r *groupRepository) RemoveMember(groupID, userID string) error {
	// Gorm Many-to-Many deletion or simply delete from join table
	return r.db.Where("group_id = ? AND user_id = ?", groupID, userID).Delete(&models.GroupMember{}).Error
}

func (r *groupRepository) FindMembersByGroupID(groupID string) ([]models.User, error) {
	var group models.Group
	err := r.db.Preload("Members").Where("id = ?", groupID).First(&group).Error
	if err != nil {
		return nil, err
	}
	return group.Members, nil
}

func (r *groupRepository) GroupNameExists(name string) (bool, error) {
	var count int64
	err := r.db.Model(&models.Group{}).Where("name = ?", name).Count(&count).Error
	return count > 0, err
}

func (r *groupRepository) IsMember(groupID, userID string) (bool, error) {
	var count int64
	err := r.db.Model(&models.GroupMember{}).Where("group_id = ? AND user_id = ?", groupID, userID).Count(&count).Error
	return count > 0, err
}
