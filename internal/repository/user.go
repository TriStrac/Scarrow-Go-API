package repository

import (
	"errors"

	"github.com/TriStrac/Scarrow-Go-API/internal/models"
	"gorm.io/gorm"
)

type UserRepository interface {
	CreateUser(user *models.User) error
	FindByUsername(username string) (*models.User, error)
	FindByPhoneNumber(phoneNumber string) ([]models.User, error)
	FindByID(id string) (*models.User, error)
	FindWithGroupByID(id string) (*models.User, error)
	FindAll() ([]models.User, error)
	UpdateUser(user *models.User) error
	SoftDelete(id string) error
	HardDelete(id string) error
	UsernameExists(username string) (bool, error)

	// User Address
	CreateUserAddress(address *models.UserAddress) error
	UpdateUserAddress(address *models.UserAddress) error
}

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) CreateUser(user *models.User) error {
	return r.db.Create(user).Error
}

func (r *userRepository) FindByUsername(username string) (*models.User, error) {
	var user models.User
	err := r.db.Preload("Profile").Where("username = ?", username).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // Return nil, nil if not found to handle it in service
		}
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) FindByPhoneNumber(phoneNumber string) ([]models.User, error) {
	var users []models.User
	err := r.db.Joins("JOIN user_profiles ON user_profiles.user_id = users.user_id").
		Where("user_profiles.phone_number = ?", phoneNumber).
		Preload("Profile").
		Find(&users).Error
	return users, err
}

func (r *userRepository) FindByID(id string) (*models.User, error) {
	var user models.User
	// Preload nested Profile and Address data (Standard load for updates)
	err := r.db.Preload("Profile").Preload("Address").Where("user_id = ?", id).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) FindWithGroupByID(id string) (*models.User, error) {
	var user models.User
	// Preload nested Profile, Address, and Group data (Strictly for read-only session data)
	err := r.db.Preload("Profile").Preload("Address").Preload("Group").Where("user_id = ?", id).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) FindAll() ([]models.User, error) {
	var users []models.User
	err := r.db.Find(&users).Error
	return users, err
}

func (r *userRepository) UpdateUser(user *models.User) error {
	// Updates both the user and associated nested structs if they exist in the model
	return r.db.Session(&gorm.Session{FullSaveAssociations: true}).Save(user).Error
}

func (r *userRepository) SoftDelete(id string) error {
	return r.db.Model(&models.User{}).Where("user_id = ?", id).Update("is_deleted", true).Delete(&models.User{ID: id}).Error
}

func (r *userRepository) HardDelete(id string) error {
	// First un-scope delete directly related records to ensure clean wipe if DB constraints aren't doing it
	r.db.Unscoped().Where("user_id = ?", id).Delete(&models.UserProfile{})
	r.db.Unscoped().Where("user_id = ?", id).Delete(&models.UserAddress{})
	r.db.Unscoped().Where("user_id = ?", id).Delete(&models.PushToken{})
	r.db.Unscoped().Where("user_id = ?", id).Delete(&models.UserSubscription{})
	r.db.Unscoped().Where("user_id = ?", id).Delete(&models.Notification{})
	r.db.Unscoped().Where("user_id = ?", id).Delete(&models.UserActivityLog{})
	
	// Message threads (either user A or user B)
	var threads []models.MessageThread
	r.db.Unscoped().Where("user_a_id = ? OR user_b_id = ?", id, id).Find(&threads)
	for _, t := range threads {
		r.db.Unscoped().Where("thread_id = ?", t.ID).Delete(&models.Message{})
		r.db.Unscoped().Delete(&t)
	}

	// Delete standalone messages just in case
	r.db.Unscoped().Where("sender_id = ?", id).Delete(&models.Message{})

	// Finally hard delete the user itself
	return r.db.Unscoped().Where("user_id = ?", id).Delete(&models.User{}).Error
}

func (r *userRepository) UsernameExists(username string) (bool, error) {
	var count int64
	err := r.db.Model(&models.User{}).Where("username = ?", username).Count(&count).Error
	return count > 0, err
}

func (r *userRepository) SavePushToken(token *models.PushToken) error {
	// Use Save or Clause(clause.OnConflict) to handle updates if the token already exists
	var existing models.PushToken
	if err := r.db.Where("token = ?", token.Token).First(&existing).Error; err == nil {
		// Update the existing token's user ID if it was transferred
		existing.UserID = token.UserID
		existing.Platform = token.Platform
		return r.db.Save(&existing).Error
	}
	return r.db.Create(token).Error
}

func (r *userRepository) RemovePushToken(tokenID string) error {
	return r.db.Where("token_id = ?", tokenID).Delete(&models.PushToken{}).Error
}

func (r *userRepository) GetPushTokensByUser(userID string) ([]models.PushToken, error) {
	var tokens []models.PushToken
	err := r.db.Where("user_id = ?", userID).Find(&tokens).Error
	return tokens, err
}

func (r *userRepository) CreateUserAddress(address *models.UserAddress) error {
	return r.db.Create(address).Error
}

func (r *userRepository) UpdateUserAddress(address *models.UserAddress) error {
	return r.db.Save(address).Error
}
