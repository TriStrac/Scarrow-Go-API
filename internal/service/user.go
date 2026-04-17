package service

import (
	"errors"

	"github.com/TriStrac/Scarrow-Go-API/internal/models"
	"github.com/TriStrac/Scarrow-Go-API/internal/repository"
	"github.com/TriStrac/Scarrow-Go-API/pkg/utils"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type UserFullResponse struct {
	User           *models.User     `json:"user"`
	Devices        []models.Device  `json:"devices"`
	RecentMessages []models.Message `json:"recent_messages"`
	UnreadCount    int64            `json:"unread_messages_count"`
}

type MeResponse struct {
	UserID             string  `json:"user_id"`
	Username           string  `json:"username"`
	Role               string  `json:"role"`
	GroupID            *string `json:"group_id"`
	GroupName          *string `json:"group_name"`
	SubscriptionStatus string  `json:"subscription_status"`
	ProfileComplete    bool    `json:"profile_complete"`
}

type UserService interface {
	Register(user *models.User) (*models.User, error)
	VerifyUser(identifier string) error
	ValidateCredentials(username, password string) (*models.User, error)
	Login(userId string) (string, error)
	GetAllUsers() ([]models.User, error)
	GetUserByID(id string) (*models.User, error)
	GetUserFullProfile(id string) (*UserFullResponse, error)
	GetMeSession(id string) (*MeResponse, error)
	UpdateUser(id string, user *models.User) error
	ChangePassword(id, newPassword string) error
	SoftDelete(id string) error
	UsernameExists(username string) (bool, error)
	FindByUsername(username string) (*models.User, error)
	FindByPhoneNumber(phoneNumber string) ([]models.User, error)
	SavePushToken(userID, token, platform string) error
	RemovePushToken(tokenID string) error
}

type userService struct {
	repo        repository.UserRepository
	deviceRepo  repository.DeviceRepository
	messageRepo repository.MessageRepository
}

func NewUserService(repo repository.UserRepository, deviceRepo repository.DeviceRepository, messageRepo repository.MessageRepository) UserService {
	return &userService{repo: repo, deviceRepo: deviceRepo, messageRepo: messageRepo}
}

func (s *userService) Register(user *models.User) (*models.User, error) {
	// Check if username exists
	exists, err := s.repo.UsernameExists(user.Username)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("username already exists")
	}

	// Check if phone number is already registered
	if user.Profile != nil && user.Profile.PhoneNumber != "" {
		existingUsers, err := s.repo.FindByPhoneNumber(user.Profile.PhoneNumber)
		if err != nil {
			return nil, err
		}
		if len(existingUsers) > 0 {
			return nil, errors.New("phone number is already registered")
		}
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	user.Password = string(hashedPassword)

	// Generate UUIDs
	user.ID = uuid.New().String()

	// Initial profile if provided
	if user.Profile != nil {
		user.Profile.ID = uuid.New().String()
		user.Profile.UserID = user.ID
	}

	err = s.repo.CreateUser(user)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (s *userService) VerifyUser(identifier string) error {
	user, err := s.repo.FindByUsername(identifier)
	if err != nil {
		return err
	}
	if user == nil {
		return errors.New("user not found")
	}

	user.IsVerified = true
	return s.repo.UpdateUser(user)
}

func (s *userService) ValidateCredentials(username, password string) (*models.User, error) {
	user, err := s.repo.FindByUsername(username)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("invalid username or password")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return nil, errors.New("invalid username or password")
	}

	if !user.IsVerified {
		return user, errors.New("user is not verified")
	}

	return user, nil
}

func (s *userService) Login(userId string) (string, error) {
	// Generate JWT
	token, err := utils.GenerateToken(userId)
	if err != nil {
		return "", err
	}

	return token, nil
}

func (s *userService) GetAllUsers() ([]models.User, error) {
	return s.repo.FindAll()
}

func (s *userService) GetUserByID(id string) (*models.User, error) {
	user, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("user not found")
	}
	return user, nil
}

func (s *userService) GetUserFullProfile(id string) (*UserFullResponse, error) {
	user, err := s.repo.FindByID(id)
	if err != nil || user == nil {
		return nil, errors.New("user not found")
	}

	// Fetch Devices
	ownerIDs := []string{user.ID}
	if user.GroupID != nil {
		ownerIDs = append(ownerIDs, *user.GroupID)
	}
	devices, _ := s.deviceRepo.GetDevicesByOwnerIDs(ownerIDs)

	// Fetch Recent Messages
	recentMessages, _ := s.messageRepo.GetRecentMessages(user.ID, 5)

	// Fetch Unread Count
	unreadCount, _ := s.messageRepo.UnreadCountByUser(user.ID)

	return &UserFullResponse{
		User:           user,
		Devices:        devices,
		RecentMessages: recentMessages,
		UnreadCount:    unreadCount,
	}, nil
}

func (s *userService) GetMeSession(id string) (*MeResponse, error) {
	user, err := s.repo.FindWithGroupByID(id)
	if err != nil || user == nil {
		return nil, errors.New("user not found")
	}

	role := "SOLO"
	if user.GroupID != nil {
		if user.IsHead {
			role = "HEAD"
		} else {
			role = "MEMBER"
		}
	}

	var groupName *string
	if user.Group != nil {
		groupName = &user.Group.Name
	}

	profileComplete := false
	if user.Profile != nil && user.Profile.FirstName != "" && user.Profile.LastName != "" {
		profileComplete = true
	}

	return &MeResponse{
		UserID:             user.ID,
		Username:           user.Username,
		Role:               role,
		GroupID:            user.GroupID,
		GroupName:          groupName,
		SubscriptionStatus: user.SubscriptionStatus,
		ProfileComplete:    profileComplete,
	}, nil
}

func (s *userService) UpdateUser(id string, inputData *models.User) error {
	user, err := s.repo.FindByID(id)
	if err != nil {
		return err
	}
	if user == nil {
		return errors.New("user not found")
	}

	// Update basic fields
	if inputData.Username != "" && inputData.Username != user.Username {
		exists, err := s.repo.UsernameExists(inputData.Username)
		if err != nil {
			return err
		}
		if exists {
			return errors.New("username already exists")
		}
		user.Username = inputData.Username
	}

	// Update Profile (if provided)
	if inputData.Profile != nil {
		if user.Profile == nil {
			user.Profile = &models.UserProfile{ID: uuid.New().String(), UserID: user.ID}
		}
		if inputData.Profile.FirstName != "" {
			user.Profile.FirstName = inputData.Profile.FirstName
		}
		if inputData.Profile.MiddleName != "" {
			user.Profile.MiddleName = inputData.Profile.MiddleName
		}
		if inputData.Profile.LastName != "" {
			user.Profile.LastName = inputData.Profile.LastName
		}
		if inputData.Profile.PhoneNumber != "" {
			user.Profile.PhoneNumber = inputData.Profile.PhoneNumber
		}
		if !inputData.Profile.BirthDate.IsZero() {
			user.Profile.BirthDate = inputData.Profile.BirthDate
		}
	}

	// Update Address (if provided)
	if inputData.Address != nil {
		if user.Address == nil {
			user.Address = &models.UserAddress{ID: uuid.New().String(), UserID: user.ID}
		}
		if inputData.Address.StreetName != "" {
			user.Address.StreetName = inputData.Address.StreetName
		}
		if inputData.Address.Baranggay != "" {
			user.Address.Baranggay = inputData.Address.Baranggay
		}
		if inputData.Address.Town != "" {
			user.Address.Town = inputData.Address.Town
		}
		if inputData.Address.Province != "" {
			user.Address.Province = inputData.Address.Province
		}
		if inputData.Address.ZipCode != "" {
			user.Address.ZipCode = inputData.Address.ZipCode
		}
	}

	return s.repo.UpdateUser(user)
}

func (s *userService) ChangePassword(id, newPassword string) error {
	user, err := s.repo.FindByID(id)
	if err != nil {
		return err
	}
	if user == nil {
		return errors.New("user not found")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.Password = string(hashedPassword)

	return s.repo.UpdateUser(user)
}

func (s *userService) SoftDelete(id string) error {
	// Unpair all devices owned by this user
	_ = s.deviceRepo.RemoveAllOwnersByOwner(id, "USER")
	return s.repo.SoftDelete(id)
}

func (s *userService) UsernameExists(username string) (bool, error) {
	return s.repo.UsernameExists(username)
}

func (s *userService) FindByUsername(username string) (*models.User, error) {
	return s.repo.FindByUsername(username)
}

func (s *userService) FindByPhoneNumber(phoneNumber string) ([]models.User, error) {
	return s.repo.FindByPhoneNumber(phoneNumber)
}

func (s *userService) SavePushToken(userID, token, platform string) error {
	pushToken := &models.PushToken{
		ID:       uuid.New().String(),
		UserID:   userID,
		Token:    token,
		Platform: platform,
	}
	return s.repo.SavePushToken(pushToken)
}

func (s *userService) RemovePushToken(tokenID string) error {
	return s.repo.RemovePushToken(tokenID)
}
