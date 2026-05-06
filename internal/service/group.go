package service

import (
	"crypto/rand"
	"errors"
	"math/big"
	"time"

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
	DisbandGroup(id string) error
	AddMemberByUsername(groupID, username string) error
	RemoveMember(groupID, userID string) error
	GetGroupMembers(groupID string) ([]MemberResponse, error)
	GetMemberDevices(groupID, callerID, memberID string) ([]models.Device, error)
	GetMemberActivityLogs(groupID, callerID, memberID string, limit, offset int) ([]models.UserActivityLog, error)

	// Invitations
	CreateInvitation(groupID, creatorID string) (*models.GroupInvitation, error)
	JoinGroupByCode(code, userID string) error
	GetGroupDetails(groupID string, callerID string) (*GroupDetailResponse, error)
}

type MemberResponse struct {
	UserID      string `json:"user_id"`
	DisplayName string `json:"display_name"`
	Role        string `json:"role"`
}

type GroupDetailResponse struct {
	ID          string      `json:"id"`
	Name        string      `json:"name"`
	OwnerID     string      `json:"owner_id"`
	Role        string      `json:"role"`
	MemberCount int64       `json:"member_count"`
	Settings    interface{} `json:"settings"` // Placeholder for future settings
}

type groupService struct {
	groupRepo        repository.GroupRepository
	userRepo         repository.UserRepository
	deviceRepo       repository.DeviceRepository
	notification     NotificationService
	invitationRepo   repository.GroupInvitationRepository
	activityLogRepo  repository.ActivityLogRepository
}

// NewGroupService requires dependencies for cascading logic and notifications
func NewGroupService(
	groupRepo repository.GroupRepository,
	userRepo repository.UserRepository,
	deviceRepo repository.DeviceRepository,
	notification NotificationService,
	invitationRepo repository.GroupInvitationRepository,
	activityLogRepo repository.ActivityLogRepository,
) GroupService {
	return &groupService{
		groupRepo:        groupRepo,
		userRepo:         userRepo,
		deviceRepo:       deviceRepo,
		notification:     notification,
		invitationRepo:   invitationRepo,
		activityLogRepo:  activityLogRepo,
	}
}

func (s *groupService) CreateGroup(name, ownerID string) (*models.Group, error) {
	// 1. Check if user is already in a group (Company logic: strict 1-to-many)
	user, err := s.userRepo.FindByID(ownerID)
	if err != nil {
		return nil, err
	}
	if user.GroupID != nil && *user.GroupID != "" {
		return nil, errors.New("user already belongs to a group/company, cannot create another")
	}

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

	// Make the owner the head of this newly created group
	user.GroupID = &group.ID
	user.IsHead = true
	user.IsInGroup = true
	_ = s.userRepo.UpdateUser(user)

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

func (s *groupService) GetGroupDetails(groupID string, callerID string) (*GroupDetailResponse, error) {
	group, err := s.groupRepo.FindByID(groupID)
	if err != nil {
		return nil, err
	}
	if group == nil {
		return nil, errors.New("group not found")
	}

	members, err := s.groupRepo.FindMembersByGroupID(groupID)
	if err != nil {
		return nil, err
	}

	role := "GUEST" // Or NOT_MEMBER
	if group.OwnerID == callerID {
		role = "HEAD"
	} else {
		for _, member := range members {
			if member.ID == callerID {
				role = "MEMBER"
				break
			}
		}
	}

	return &GroupDetailResponse{
		ID:          group.ID,
		Name:        group.Name,
		OwnerID:     group.OwnerID,
		Role:        role,
		MemberCount: int64(len(members)),
		Settings:    map[string]interface{}{}, // empty for now
	}, nil
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

func (s *groupService) DisbandGroup(id string) error {
	group, err := s.groupRepo.FindByID(id)
	if err != nil {
		return err
	}
	if group == nil {
		return errors.New("group not found")
	}

	// 1. Get all members before clearing them to send notifications
	members, err := s.groupRepo.FindMembersByGroupID(id)
	if err == nil {
		for _, member := range members {
			_ = s.notification.CreateNotification(member.ID, "Group Disbanded", "The group/company '"+group.Name+"' has been disbanded.")
		}
	}

	// 2. Clear members from group
	_ = s.groupRepo.ClearGroupMembers(id)

	// 3. Invalidate/Delete all pending invitations for this group
	_ = s.invitationRepo.DeleteByGroupID(id)

	// 4. Soft delete the group
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

func (s *groupService) GetGroupMembers(groupID string) ([]MemberResponse, error) {
	// 1. We also need to know the owner to label them as HEAD.
	group, err := s.groupRepo.FindByID(groupID)
	if err != nil || group == nil {
		return nil, errors.New("group not found")
	}

	users, err := s.groupRepo.FindMembersByGroupID(groupID)
	if err != nil {
		return nil, err
	}

	responses := make([]MemberResponse, 0)
	for _, user := range users {
		displayName := user.Username
		if user.Profile != nil && user.Profile.FirstName != "" {
			displayName = user.Profile.FirstName + " " + user.Profile.LastName
		}

		role := "MEMBER"
		if user.ID == group.OwnerID {
			role = "HEAD"
		} else if user.IsHead {
			role = "HEAD" // In case you support multiple heads later
		}

		responses = append(responses, MemberResponse{
			UserID:      user.ID,
			DisplayName: displayName,
			Role:        role,
		})
	}

	return responses, nil
}

func (s *groupService) GetMemberDevices(groupID, callerID, memberID string) ([]models.Device, error) {
	group, err := s.groupRepo.FindByID(groupID)
	if err != nil || group == nil {
		return nil, errors.New("group not found")
	}
	if group.OwnerID != callerID {
		return nil, errors.New("forbidden: only group head can view member devices")
	}

	member, err := s.userRepo.FindByID(memberID)
	if err != nil {
		return nil, err
	}
	if member == nil {
		return nil, errors.New("member not found")
	}
	if member.GroupID == nil || *member.GroupID != groupID {
		return nil, errors.New("member does not belong to this group")
	}

	return s.deviceRepo.GetDevicesByUserID(memberID)
}

func (s *groupService) GetMemberActivityLogs(groupID, callerID, memberID string, limit, offset int) ([]models.UserActivityLog, error) {
	group, err := s.groupRepo.FindByID(groupID)
	if err != nil || group == nil {
		return nil, errors.New("group not found")
	}
	if group.OwnerID != callerID {
		return nil, errors.New("forbidden: only group head can view member activity logs")
	}

	member, err := s.userRepo.FindByID(memberID)
	if err != nil {
		return nil, err
	}
	if member == nil {
		return nil, errors.New("member not found")
	}
	if member.GroupID == nil || *member.GroupID != groupID {
		return nil, errors.New("member does not belong to this group")
	}

	return s.activityLogRepo.GetLogsByUserIDPaginated(memberID, limit, offset)
}

func (s *groupService) CreateInvitation(groupID, creatorID string) (*models.GroupInvitation, error) {
	code, err := s.generateRandomAlphanumeric(8)
	if err != nil {
		return nil, err
	}

	invitation := &models.GroupInvitation{
		Code:      code,
		GroupID:   groupID,
		CreatedBy: creatorID,
		ExpiresAt: time.Now().Add(48 * time.Hour),
	}

	err = s.invitationRepo.Create(invitation)
	if err != nil {
		return nil, err
	}

	return invitation, nil
}

func (s *groupService) JoinGroupByCode(code, userID string) error {
	invitation, err := s.invitationRepo.FindByCode(code)
	if err != nil {
		return err
	}
	if invitation == nil {
		return errors.New("invalid or expired invitation code")
	}

	// 1. Get Group to check ownership
	group, err := s.groupRepo.FindByID(invitation.GroupID)
	if err != nil || group == nil {
		return errors.New("group no longer exists")
	}

	if group.OwnerID == userID {
		return errors.New("you are already the owner of this group")
	}

	// 2. Check if user is already in a group
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return err
	}
	if user.GroupID != nil && *user.GroupID != "" {
		return errors.New("user already belongs to a group")
	}

	// 3. Join group
	return s.groupRepo.AddMember(invitation.GroupID, userID)
}

func (s *groupService) generateRandomAlphanumeric(length int) (string, error) {
	const charset = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789" // Removed confusing chars like 0, O, 1, I
	result := make([]byte, length)
	for i := range result {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", err
		}
		result[i] = charset[num.Int64()]
	}
	return string(result), nil
}
