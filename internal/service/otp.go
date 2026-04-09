package service

import (
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/TriStrac/Scarrow-Go-API/internal/models"
	"github.com/TriStrac/Scarrow-Go-API/internal/repository"
	"github.com/TriStrac/Scarrow-Go-API/pkg/utils"
	"github.com/google/uuid"
)

type OTPService interface {
	GenerateAndSendOTP(identifier string, purpose models.OTPPurpose) (string, error)
	VerifyOTP(identifier string, code string, purpose models.OTPPurpose) (bool, error)
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

func (s *otpService) GenerateAndSendOTP(identifier string, purpose models.OTPPurpose) (string, error) {
	// 1. Rate Limiting: Max 3 OTPs per 10 minutes
	count, err := s.repo.CountRecentOTPs(identifier, 10*time.Minute)
	if err != nil {
		return "", err
	}
	if count >= 3 {
		return "", errors.New("too many OTP requests. please try again in 10 minutes")
	}

	// 2. Generate 6-digit code
	code, err := s.generateRandomCode(6)
	if err != nil {
		return "", err
	}

	// 3. Create OTP record
	otp := &models.OTPCode{
		ID:         uuid.New().String(),
		Identifier: identifier,
		Code:       code,
		Purpose:    purpose,
		ExpiresAt:  time.Now().Add(5 * time.Minute), // 5 minutes expiration
	}

	if err := s.repo.CreateOTP(otp); err != nil {
		return "", err
	}

	// 4. Send SMS
	message := fmt.Sprintf("Your Scarrow verification code is: %s. Valid for 5 minutes.", code)
	_ = s.smsService.SendSMS(identifier, message)

	return code, nil
}

func (s *otpService) VerifyOTP(identifier string, code string, purpose models.OTPPurpose) (bool, error) {
	otp, err := s.repo.GetLatestOTP(identifier, purpose)
	if err != nil {
		return false, errors.New("invalid or expired OTP")
	}

	if otp.ExpiresAt.Before(time.Now()) {
		return false, errors.New("OTP has expired")
	}

	if otp.Code != code {
		return false, errors.New("incorrect OTP code")
	}

	// Mark as used
	if err := s.repo.MarkAsUsed(otp.ID); err != nil {
		return false, err
	}

	return true, nil
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
