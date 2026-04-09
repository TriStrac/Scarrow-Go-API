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
	AddMember(groupID, userID string) error
	RemoveMember(groupID, userID string) error
	ClearGroupMembers(groupID string) error
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

func (r *groupRepository) AddMember(groupID, userID string) error {
	// Directly update the user's group_id
	return r.db.Model(&models.User{}).Where("id = ?", userID).Updates(map[string]interface{}{
		"group_id":         groupID,
		"is_user_in_group": true,
	}).Error
}

func (r *groupRepository) RemoveMember(groupID, userID string) error {
	// Directly nullify the user's group_id
	return r.db.Model(&models.User{}).Where("id = ? AND group_id = ?", userID, groupID).Updates(map[string]interface{}{
		"group_id":         nil,
		"is_user_in_group": false,
	}).Error
}

func (r *groupRepository) ClearGroupMembers(groupID string) error {
	return r.db.Model(&models.User{}).Where("group_id = ?", groupID).Updates(map[string]interface{}{
		"group_id":         nil,
		"is_user_in_group": false,
	}).Error
}

func (r *groupRepository) FindMembersByGroupID(groupID string) ([]models.User, error) {
	var users []models.User
	err := r.db.Where("group_id = ?", groupID).Find(&users).Error
	return users, err
}

func (r *groupRepository) GroupNameExists(name string) (bool, error) {
	var count int64
	err := r.db.Model(&models.Group{}).Where("name = ?", name).Count(&count).Error
	return count > 0, err
}

func (r *groupRepository) IsMember(groupID, userID string) (bool, error) {
	var count int64
	err := r.db.Model(&models.User{}).Where("id = ? AND group_id = ?", userID, groupID).Count(&count).Error
	return count > 0, err
}
