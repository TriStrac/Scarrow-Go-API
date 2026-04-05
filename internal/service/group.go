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

	// 3. Check if user is already a member
	isMember, err := s.groupRepo.IsMember(groupID, user.ID)
	if err != nil {
		return err
	}
	if isMember {
		return errors.New("user is already a member of this group")
	}

	// 4. Add Member
	groupMember := &models.GroupMember{
		GroupID: groupID,
		UserID:  user.ID,
	}

	err = s.groupRepo.AddMember(groupMember)
	if err != nil {
		return err
	}

	// 5. Update user to mark them as in a group (based on the legacy nodejs logic implied)
	user.IsInGroup = true
	return s.userRepo.UpdateUser(user)
}

func (s *groupService) RemoveMember(groupID, userID string) error {
	// 1. Remove Member
	err := s.groupRepo.RemoveMember(groupID, userID)
	if err != nil {
		return err
	}

	// 2. Check if the user is in ANY other groups
	user, err := s.userRepo.FindByID(userID)
	if err == nil && user != nil {
		count, countErr := s.groupRepo.CountGroupsByUserID(userID)
		if countErr == nil {
			if count == 0 {
				user.IsInGroup = false
			} else {
				user.IsInGroup = true
			}
			s.userRepo.UpdateUser(user)
		}
	}

	return nil
}
func (s *groupService) GetGroupMembers(groupID string) ([]models.User, error) {
	return s.groupRepo.FindMembersByGroupID(groupID)
}
