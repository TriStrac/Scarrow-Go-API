package service

import (
	"errors"

	"github.com/TriStrac/Scarrow-Go-API/internal/models"
	"github.com/TriStrac/Scarrow-Go-API/internal/repository"
	"github.com/TriStrac/Scarrow-Go-API/pkg/utils"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type UserService interface {
	Register(user *models.User) (*models.User, error)
	Login(username, password string) (string, error)
	GetAllUsers() ([]models.User, error)
	GetUserByID(id string) (*models.User, error)
	UpdateUser(id string, user *models.User) error
	ChangePassword(id, newPassword string) error
	SoftDelete(id string) error
	UsernameExists(username string) (bool, error)
}

type userService struct {
	repo repository.UserRepository
}

func NewUserService(repo repository.UserRepository) UserService {
	return &userService{repo: repo}
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

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	user.Password = string(hashedPassword)

	// Generate UUIDs
	user.ID = uuid.New().String()
	user.Profile.ID = uuid.New().String()
	user.Profile.UserID = user.ID
	user.Address.ID = uuid.New().String()
	user.Address.UserID = user.ID

	err = s.repo.CreateUser(user)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (s *userService) Login(username, password string) (string, error) {
	user, err := s.repo.FindByUsername(username)
	if err != nil {
		return "", err
	}
	if user == nil {
		return "", errors.New("invalid username or password")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return "", errors.New("invalid username or password")
	}

	// Generate JWT
	token, err := utils.GenerateToken(user.ID)
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
	return s.repo.SoftDelete(id)
}

func (s *userService) UsernameExists(username string) (bool, error) {
	return s.repo.UsernameExists(username)
}
