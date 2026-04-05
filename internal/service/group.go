package service

import (
	"errors"

	"github.com/TriStrac/Scarrow-Go-API/internal/models"
	"github.com/TriStrac/Scarrow-Go-API/internal/repository"
	"github.com/google/uuid"
)

type GroupService interface {
	CreateGroup(name, ownerID string) (*models.Group, error)
	GetAllGroups() ([]models.Group, error)
	GetGroupsByOwner(ownerID string) ([]models.Group, error)
	GetGroupByID(id string) (*models.Group, error)
	UpdateGroup(id string, name string) error
	SoftDeleteGroup(id string) error
	AddMemberByUsername(groupID, username string) error
	RemoveMember(groupID, userID string) error
	GetGroupMembers(groupID string) ([]models.User, error)
}

type groupService struct {
	groupRepo repository.GroupRepository
	userRepo  repository.UserRepository
}

// NewGroupService requires both the group repo and the user repo
// because we need to look up users by username when adding members.
func NewGroupService(groupRepo repository.GroupRepository, userRepo repository.UserRepository) GroupService {
	return &groupService{
		groupRepo: groupRepo,
		userRepo:  userRepo,
	}
}

func (s *groupService) CreateGroup(name, ownerID string) (*models.Group, error) {
	exists, err := s.groupRepo.GroupNameExists(name)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("group name already exists")
	}

	group := &models.Group{
		ID:      uuid.New().String(),
		Name:    name,
		OwnerID: ownerID,
	}

	err = s.groupRepo.CreateGroup(group)
	if err != nil {
		return nil, err
	}
	return group, nil
}

func (s *groupService) GetAllGroups() ([]models.Group, error) {
	return s.groupRepo.FindAll()
}

func (s *groupService) GetGroupsByOwner(ownerID string) ([]models.Group, error) {
	return s.groupRepo.FindByOwnerID(ownerID)
}

func (s *groupService) GetGroupByID(id string) (*models.Group, error) {
	group, err := s.groupRepo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if group == nil {
		return nil, errors.New("group not found")
	}
	return group, nil
}

func (s *groupService) UpdateGroup(id string, name string) error {
	group, err := s.groupRepo.FindByID(id)
	if err != nil {
		return err
	}
	if group == nil {
		return errors.New("group not found")
	}

	if name != "" && name != group.Name {
		exists, err := s.groupRepo.GroupNameExists(name)
		if err != nil {
			return err
		}
		if exists {
			return errors.New("group name already exists")
		}
		group.Name = name
	}

	return s.groupRepo.UpdateGroup(group)
}

func (s *groupService) SoftDeleteGroup(id string) error {
	return s.groupRepo.SoftDelete(id)
}

func (s *groupService) AddMemberByUsername(groupID, username string) error {
	// 1. Validate Group Exists
	group, err := s.groupRepo.FindByID(groupID)
	if err != nil {
		return err
	}
	if group == nil {
		return errors.New("group not found")
	}

	// 2. Fetch User by Username
	user, err := s.userRepo.FindByUsername(username)
	if err != nil {
		return err
	}
	if user == nil {
		return errors.New("user not found")
	}

	// 3. Check if user is already in a group (Company logic: strict 1-to-many)
	if user.GroupID != nil && *user.GroupID != "" {
		return errors.New("user already belongs to a group/company")
	}

	// 4. Update the User's GroupID directly
	return s.groupRepo.AddMember(groupID, user.ID)
}

func (s *groupService) RemoveMember(groupID, userID string) error {
	// Remove Member (nullify GroupID)
	return s.groupRepo.RemoveMember(groupID, userID)
}

func (s *groupService) GetGroupMembers(groupID string) ([]models.User, error) {
	return s.groupRepo.FindMembersByGroupID(groupID)
}
