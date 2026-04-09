package utils

import (
	"log"
)

// SmsService defines the interface for sending SMS messages
type SmsService interface {
	SendSMS(to string, message string) error
}

// MockSmsService is a mock implementation of SmsService for development
type MockSmsService struct{}

func (s *MockSmsService) SendSMS(to string, message string) error {
	log.Printf("[MOCK SMS] To: %s, Message: %s\n", to, message)
	return nil
}

func NewMockSmsService() SmsService {
	return &MockSmsService{}
}
