package service

import (
	"crypto/rand"
	"errors"
	"fmt"
	"log"
	"math/big"
	"time"

	"github.com/TriStrac/Scarrow-Go-API/internal/models"
	"github.com/TriStrac/Scarrow-Go-API/internal/repository"
	"github.com/TriStrac/Scarrow-Go-API/pkg/utils"
	"github.com/google/uuid"
)

type OTPService interface {
	GenerateAndSendOTP(identifier string, destination string, purpose models.OTPPurpose, payload string) (string, error)
	VerifyOTP(identifier string, code string, purpose models.OTPPurpose) (*models.OTPCode, error)
	GetLatestOTP(identifier string, purpose models.OTPPurpose) (*models.OTPCode, error)
}

type otpService struct {
	repo       repository.OTPRepository
	smsService utils.SmsService
}

func NewOTPService(repo repository.OTPRepository, smsService utils.SmsService) OTPService {
	return &otpService{
		repo:       repo,
		smsService: smsService,
	}
}

func (s *otpService) GetLatestOTP(identifier string, purpose models.OTPPurpose) (*models.OTPCode, error) {
	return s.repo.GetLatestOTP(identifier, purpose)
}

func (s *otpService) GenerateAndSendOTP(identifier string, destination string, purpose models.OTPPurpose, payload string) (string, error) {
	// 1. Check for existing unused and unexpired OTP
	existingOtp, err := s.repo.GetLatestOTP(identifier, purpose)
	if err == nil && existingOtp != nil {
		// Check if it's still valid (e.g., has at least 1 minute left)
		if existingOtp.ExpiresAt.After(time.Now().Add(1 * time.Minute)) {
			// Reuse the existing code but update the payload and destination if changed
			updated := false
			if existingOtp.Payload != payload {
				existingOtp.Payload = payload
				updated = true
			}
			if existingOtp.Destination != destination {
				existingOtp.Destination = destination
				updated = true
			}

			if updated {
				if err := s.repo.UpdateOTP(existingOtp); err != nil {
					return "", err
				}
			}
			
			// Send SMS again
			message := fmt.Sprintf("Your Scarrow verification code is: %s. Valid for 5 minutes.", existingOtp.Code)
			log.Printf("[OTP REUSED] Identifier: %s, Destination: %s, Purpose: %s, Code: %s\n", identifier, destination, purpose, existingOtp.Code)
			_ = s.smsService.SendSMS(destination, message)
			
			return existingOtp.Code, nil
		}
	}

	// 2. Rate Limiting: Max 3 OTPs per 10 minutes
	count, err := s.repo.CountRecentOTPs(identifier, 10*time.Minute)
	if err != nil {
		return "", err
	}
	if count >= 3 {
		return "", errors.New("too many OTP requests. please try again in 10 minutes")
	}

	// 3. Generate 6-digit code
	code, err := s.generateRandomCode(6)
	if err != nil {
		return "", err
	}

	// 4. Create OTP record
	otp := &models.OTPCode{
		ID:          uuid.New().String(),
		Identifier:  identifier,
		Destination: destination,
		Code:        code,
		Purpose:     purpose,
		Payload:     payload,
		ExpiresAt:   time.Now().Add(5 * time.Minute), // 5 minutes expiration
	}

	if err := s.repo.CreateOTP(otp); err != nil {
		return "", err
	}

	// 5. Send SMS
	message := fmt.Sprintf("Your Scarrow verification code is: %s. Valid for 5 minutes.", code)
	log.Printf("[OTP GENERATED] Identifier: %s, Destination: %s, Purpose: %s, Code: %s\n", identifier, destination, purpose, code)
	_ = s.smsService.SendSMS(destination, message)

	return code, nil
}

func (s *otpService) VerifyOTP(identifier string, code string, purpose models.OTPPurpose) (*models.OTPCode, error) {
	otp, err := s.repo.GetLatestOTP(identifier, purpose)
	if err != nil {
		return nil, errors.New("invalid or expired OTP")
	}

	if otp.ExpiresAt.Before(time.Now()) {
		return nil, errors.New("OTP has expired")
	}

	if otp.Code != code {
		return nil, errors.New("incorrect OTP code")
	}

	// Mark as used
	if err := s.repo.MarkAsUsed(otp.ID); err != nil {
		return nil, err
	}

	return otp, nil
}

func (s *otpService) generateRandomCode(length int) (string, error) {
	const digits = "0123456789"
	result := make([]byte, length)
	for i := 0; i < length; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(digits))))
		if err != nil {
			return "", err
		}
		result[i] = digits[num.Int64()]
	}
	return string(result), nil
}
