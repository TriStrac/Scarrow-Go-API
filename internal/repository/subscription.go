package repository

import (
	"errors"

	"github.com/TriStrac/Scarrow-Go-API/internal/models"
	"gorm.io/gorm"
)

type SubscriptionRepository interface {
	GetAvailablePlans() ([]models.SubscriptionPlan, error)
	GetPlanByID(planID string) (*models.SubscriptionPlan, error)
	GetUserActiveSubscription(userID string) (*models.UserSubscription, error)
	CreateOrUpdateUserSubscription(sub *models.UserSubscription) error
	
	// Seed function just for testing
	SeedDefaultPlans() error 
}

type subscriptionRepository struct {
	db *gorm.DB
}

func NewSubscriptionRepository(db *gorm.DB) SubscriptionRepository {
	return &subscriptionRepository{db: db}
}

func (r *subscriptionRepository) GetAvailablePlans() ([]models.SubscriptionPlan, error) {
	var plans []models.SubscriptionPlan
	err := r.db.Find(&plans).Error
	return plans, err
}

func (r *subscriptionRepository) GetPlanByID(planID string) (*models.SubscriptionPlan, error) {
	var plan models.SubscriptionPlan
	err := r.db.Where("id = ?", planID).First(&plan).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &plan, nil
}

func (r *subscriptionRepository) GetUserActiveSubscription(userID string) (*models.UserSubscription, error) {
	var sub models.UserSubscription
	// Fetch the most recent active subscription
	err := r.db.Preload("Plan").
		Where("user_id = ? AND status = ?", userID, "ACTIVE").
		Order("end_date desc").
		First(&sub).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // No active subscription
		}
		return nil, err
	}
	return &sub, nil
}

func (r *subscriptionRepository) CreateOrUpdateUserSubscription(sub *models.UserSubscription) error {
	return r.db.Save(sub).Error
}

func (r *subscriptionRepository) SeedDefaultPlans() error {
	// Simple seed for demonstration and immediate frontend integration testing
	var count int64
	r.db.Model(&models.SubscriptionPlan{}).Count(&count)
	if count == 0 {
		plans := []models.SubscriptionPlan{
			{ID: "plan_monthly", Name: "Premium Farmer (Monthly)", Description: "Full access to analytics and unlimited devices for 30 days.", Price: 499.00, DurationDays: 30},
			{ID: "plan_yearly", Name: "Premium Farmer (Yearly)", Description: "Full access to analytics and unlimited devices for 365 days. Save 20%!", Price: 4990.00, DurationDays: 365},
		}
		return r.db.Create(&plans).Error
	}
	return nil
}
