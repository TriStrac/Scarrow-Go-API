package models

import (
	"time"

	"gorm.io/gorm"
)

// SubscriptionPlan represents a premium tier available for purchase.
// E.g., "Premium Farmer", "Enterprise Hub"
type SubscriptionPlan struct {
	ID           string         `gorm:"type:varchar(36);primaryKey" json:"id"`
	Name         string         `gorm:"type:varchar(100);not null;unique" json:"name"`
	Description  string         `gorm:"type:text" json:"description"`
	Price        float64        `json:"price"` // E.g., PHP 500.00
	Currency     string         `gorm:"type:varchar(10);default:'PHP'" json:"currency"`
	DurationDays int            `json:"duration_days"` // E.g., 30 for Monthly, 365 for Yearly
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

// UserSubscription tracks a user's active or past subscriptions.
type UserSubscription struct {
	ID          string    `gorm:"type:varchar(36);primaryKey" json:"id"`
	UserID      string    `gorm:"type:varchar(36);not null;index" json:"user_id"`
	PlanID      string    `gorm:"type:varchar(36);not null" json:"plan_id"`
	Status      string    `gorm:"type:varchar(50);default:'PENDING'" json:"status"` // PENDING, ACTIVE, EXPIRED, CANCELLED
	ReferenceID string    `gorm:"type:varchar(255)" json:"reference_id"`             // Stripe Checkout ID or PayMongo intent ID
	StartDate   time.Time `json:"start_date"`
	EndDate     time.Time `json:"end_date"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	Plan SubscriptionPlan `gorm:"foreignKey:PlanID;references:ID" json:"plan"`
}
