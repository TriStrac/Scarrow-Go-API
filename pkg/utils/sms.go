package utils

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// SmsService defines the interface for sending SMS messages
type SmsService interface {
	SendSMS(to string, message string) error
}

// RealSmsService implements SmsService using Semaphore API
type RealSmsService struct {
	apiKey string
	apiUrl string
}

func NewRealSmsService(apiKey string) SmsService {
	return &RealSmsService{
		apiKey: apiKey,
		apiUrl: "https://api.semaphore.co/api/v4/messages",
	}
}

func (s *RealSmsService) formatPhoneNumber(to string) string {
	to = strings.TrimSpace(to)
	// Semaphore prefers 09XXXXXXXXX or +639XXXXXXXXX
	return to
}

func (s *RealSmsService) SendSMS(to string, message string) error {
	formattedTo := s.formatPhoneNumber(to)

	err := s.sendViaSemaphore(formattedTo, message)
	if err != nil {
		log.Printf("[SMS FAILED] To: %s, Error: %v\n", formattedTo, err)
		return err
	}

	log.Printf("[SMS SUCCESS] To: %s\n", formattedTo)
	return nil
}

func (s *RealSmsService) sendViaSemaphore(to string, message string) error {
	data := url.Values{}
	data.Set("apikey", s.apiKey)
	data.Set("number", to)
	data.Set("message", message)

	req, err := http.NewRequest("POST", s.apiUrl, strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status code: %d, response: %s", resp.StatusCode, string(bodyBytes))
	}

	return nil
}

/*
// MockSmsService is a mock implementation of SmsService for development
type MockSmsService struct{}

func (s *MockSmsService) SendSMS(to string, message string) error {
	log.Printf("[MOCK SMS] To: %s, Message: %s\n", to, message)
	return nil
}

func NewMockSmsService() SmsService {
	return &MockSmsService{}
}
*/
