package service

import (
	"errors"

	"github.com/TriStrac/Scarrow-Go-API/internal/models"
	"github.com/TriStrac/Scarrow-Go-API/internal/repository"
	"github.com/google/uuid"
)

type CheckoutResponse struct {
	CheckoutURL string `json:"checkout_url"`
	ReferenceID string `json:"reference_id"`
}

type SubscriptionService interface {
	GetAvailablePlans() ([]models.SubscriptionPlan, error)
	GetMySubscription(userID string) (*models.UserSubscription, error)
	CreateCheckoutSession(userID string, planID string) (*CheckoutResponse, error)
	VerifyPayment(userID string, referenceID string) error
}

type subscriptionService struct {
	repo     repository.SubscriptionRepository
	userRepo repository.UserRepository
}

func NewSubscriptionService(repo repository.SubscriptionRepository, userRepo repository.UserRepository) SubscriptionService {
	// Let's seed default plans on initialization if they don't exist
	_ = repo.SeedDefaultPlans()
	
	return &subscriptionService{
		repo:     repo,
		userRepo: userRepo,
	}
}

func (s *subscriptionService) GetAvailablePlans() ([]models.SubscriptionPlan, error) {
	return s.repo.GetAvailablePlans()
}

func (s *subscriptionService) GetMySubscription(userID string) (*models.UserSubscription, error) {
	return s.repo.GetUserActiveSubscription(userID)
}

func (s *subscriptionService) CreateCheckoutSession(userID string, planID string) (*CheckoutResponse, error) {
	plan, err := s.repo.GetPlanByID(planID)
	if err != nil {
		return nil, err
	}
	if plan == nil {
		return nil, errors.New("plan not found")
	}

	// -------------------------------------------------------------------------
	// TODO: PAYMENT GATEWAY INTEGRATION POINT
	// -------------------------------------------------------------------------
	// 1. Call Stripe/PayMongo API to create a Checkout Session/Payment Intent
	// 2. Pass the 'plan.Price' and 'plan.Name'
	// 3. Receive the 'checkout_url' and the provider's 'reference_id'
	
	// Mock Integration values
	mockReferenceID := "mock_pay_" + uuid.New().String()
	mockCheckoutURL := "https://mock-payment-gateway.scarrow.com/checkout/" + mockReferenceID

	// Create a PENDING subscription record
	pendingSub := &models.UserSubscription{
		ID:          uuid.New().String(),
		UserID:      userID,
		PlanID:      planID,
		Status:      "PENDING",
		ReferenceID: mockReferenceID,
	}
	if err := s.repo.CreateOrUpdateUserSubscription(pendingSub); err != nil {
		return nil, err
	}

	return &CheckoutResponse{
		CheckoutURL: mockCheckoutURL,
		ReferenceID: mockReferenceID,
	}, nil
}

func (s *subscriptionService) VerifyPayment(userID string, referenceID string) error {
	// -------------------------------------------------------------------------
	// TODO: PAYMENT GATEWAY INTEGRATION POINT
	// -------------------------------------------------------------------------
	// 1. Call Stripe/PayMongo API to check the status of 'referenceID'
	// 2. If status is NOT 'paid' or 'succeeded', return error.
	
	// Assuming payment was successful, find the pending subscription
	// (In a real webhook, you wouldn't necessarily need userID, just the referenceID)
	// For this mock 'restore/verify' endpoint, we will just activate it.

	// For simplicity in this mock, we will fetch the plan duration and activate it.
	// We'd ideally fetch the exact pending 'UserSubscription' by ReferenceID,
	// but to avoid adding more repo methods for the mock, we'll just insert/update
	// a new ACTIVE one for now.
	
	// (Mocking successful verification)
	user, err := s.userRepo.FindByID(userID)
	if err != nil || user == nil {
		return errors.New("user not found")
	}

	// Update the user's string status
	user.SubscriptionStatus = "PREMIUM"
	_ = s.userRepo.UpdateUser(user)

	// In a real implementation, you would update the `UserSubscription` Status to "ACTIVE",
	// set `StartDate` to time.Now(), and `EndDate` to time.Now().AddDate(0, 0, plan.DurationDays)

	return nil
}
