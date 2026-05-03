package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/TriStrac/Scarrow-Go-API/internal/models"
	"github.com/TriStrac/Scarrow-Go-API/internal/service"
	"github.com/gin-gonic/gin"
)

type UserController struct {
	userService service.UserService
	otpService  service.OTPService
}

func NewUserController(userService service.UserService, otpService service.OTPService) *UserController {
	return &UserController{userService: userService, otpService: otpService}
}

// Request Data Transfer Objects (DTOs) for validation
type RegisterReq struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required,min=6"`
	Number   string `json:"number" binding:"required"` // Phone number for OTP
}

// Internal Payload DTO for OTP storage (bypasses json:"-" on user.Password)
type RegistrationPayload struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Number   string `json:"number"`
}

type VerifyOTPReq struct {
	Identifier string `json:"identifier" binding:"required"`
	Code       string `json:"code" binding:"required"`
}

type LoginReq struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type ForgotPasswordReq struct {
	Username string `json:"username" binding:"required"`
}

type ResendOTPReq struct {
	Identifier string            `json:"identifier" binding:"required"` // Username
	Purpose    models.OTPPurpose `json:"purpose" binding:"required,oneof=REGISTRATION LOGIN FORGOT_PASSWORD"`
}

type ResetPasswordReq struct {
	Username    string `json:"username" binding:"required"`
	OTP         string `json:"otp" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

type ChangePasswordReq struct {
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

type PatchUserAddressReq struct {
	StreetName string `json:"street_name"`
	Baranggay  string `json:"baranggay"`
	Town       string `json:"town"`
	Province   string `json:"province"`
	ZipCode    string `json:"zip_code"`
}

func (c *UserController) PatchUserAddress(ctx *gin.Context) {
	userId := ctx.Param("userId")
	callerID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	if callerID.(string) != userId {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "Forbidden: You can only modify your own address"})
		return
	}

	var req PatchUserAddressReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	address := &models.UserAddress{
		UserID:      userId,
		StreetName:  req.StreetName,
		Baranggay:    req.Baranggay,
		Town:         req.Town,
		Province:    req.Province,
		ZipCode:      req.ZipCode,
	}

	err := c.userService.UpdateUserAddress(userId, address)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Address updated successfully"})
}

func (c *UserController) Register(ctx *gin.Context) {
	var req RegisterReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 1. Check if username exists
	exists, err := c.userService.UsernameExists(req.Username)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if exists {
		ctx.JSON(http.StatusConflict, gin.H{"error": "username already exists"})
		return
	}

	// 2. Check if phone number is already registered
	users, err := c.userService.FindByPhoneNumber(req.Number)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if len(users) > 0 {
		ctx.JSON(http.StatusConflict, gin.H{"error": "phone number is already registered"})
		return
	}

	// 3. Prepare user data to be stored in OTP payload
	payloadData := RegistrationPayload{
		Username: req.Username,
		Password: req.Password,
		Number:   req.Number,
	}

	payloadBytes, err := json.Marshal(payloadData)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process registration data"})
		return
	}

	// 4. Send OTP for registration and store payload
	_, err = c.otpService.GenerateAndSendOTP(req.Username, req.Number, models.OTPPurposeRegistration, string(payloadBytes))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send OTP: " + err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"message":    "Registration initiated. Please verify with the OTP sent to your number.",
		"identifier": req.Username,
	})
}

func (c *UserController) VerifyRegistration(ctx *gin.Context) {
	var req VerifyOTPReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 1. Verify OTP using Username as identifier
	otp, err := c.otpService.VerifyOTP(req.Identifier, req.Code, models.OTPPurposeRegistration)
	if err != nil || otp == nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired OTP"})
		return
	}

	// 2. Extract user data from payload
	var payload RegistrationPayload
	if err := json.Unmarshal([]byte(otp.Payload), &payload); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse registration data"})
		return
	}

	user := &models.User{
		Username: payload.Username,
		Password: payload.Password,
		Profile: &models.UserProfile{
			PhoneNumber: payload.Number,
		},
		IsVerified: true, // we can immediately mark as verified here before saving
	}

	// 3. Register the user into the database
	savedUser, err := c.userService.Register(user)
	if err != nil {
		ctx.JSON(http.StatusConflict, gin.H{"error": "Registration failed: " + err.Error()})
		return
	}

	// Ensure they are fully verified
	err = c.userService.VerifyUser(savedUser.Username)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "User created but verification status update failed"})
		return
	}

	// 5. Auto-Login: Generate and return session token
	token, err := c.userService.Login(savedUser.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Verified, but failed to generate session token: " + err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "User verified successfully.",
		"token":   token,
		"user": gin.H{
			"id":                  savedUser.ID,
			"username":            savedUser.Username,
			"phone_number":        savedUser.Profile.PhoneNumber,
			"subscription_status": savedUser.SubscriptionStatus,
		},
	})
}

func (c *UserController) Login(ctx *gin.Context) {
	var req LoginReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := c.userService.ValidateCredentials(req.Username, req.Password)
	if err != nil {
		if err.Error() == "user is not verified" {
			ctx.JSON(http.StatusForbidden, gin.H{"error": "User is not verified. Please verify your account first."})
			return
		}
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	// Send OTP for login
	_, err = c.otpService.GenerateAndSendOTP(user.Username, user.Profile.PhoneNumber, models.OTPPurposeLogin, "")
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send OTP: " + err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message":    "OTP sent for login verification",
		"identifier": user.Username,
	})
}

func (c *UserController) VerifyLogin(ctx *gin.Context) {
	var req VerifyOTPReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := c.userService.FindByUsername(req.Identifier)
	if err != nil || user == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	_, err = c.otpService.VerifyOTP(req.Identifier, req.Code, models.OTPPurposeLogin)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired OTP"})
		return
	}

	token, err := c.userService.Login(user.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"token": token})
}

func (c *UserController) ForgotPassword(ctx *gin.Context) {
	var req ForgotPasswordReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := c.userService.FindByUsername(req.Username)
	if err != nil || user == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	_, err = c.otpService.GenerateAndSendOTP(user.Username, user.Profile.PhoneNumber, models.OTPPurposeForgotPassword, "")
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send OTP"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "OTP sent for password reset",
	})
}

func (c *UserController) ResendOTP(ctx *gin.Context) {
	var req ResendOTPReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var destination string

	if req.Purpose == models.OTPPurposeRegistration {
		// User is not in the DB yet. Fetch from the latest OTP record.
		otp, err := c.otpService.GetLatestOTP(req.Identifier, req.Purpose)
		if err != nil || otp == nil {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "No pending registration found for this username"})
			return
		}
		destination = otp.Destination
	} else {
		// For Login or ForgotPassword, user must be in the DB
		user, err := c.userService.FindByUsername(req.Identifier)
		if err != nil || user == nil {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}
		destination = user.Profile.PhoneNumber
	}

	// Resend OTP
	_, err := c.otpService.GenerateAndSendOTP(req.Identifier, destination, req.Purpose, "")
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to resend OTP: " + err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "OTP resent successfully",
	})
}

func (c *UserController) ResetPassword(ctx *gin.Context) {
	var req ResetPasswordReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := c.userService.FindByUsername(req.Username)
	if err != nil || user == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	_, err = c.otpService.VerifyOTP(req.Username, req.OTP, models.OTPPurposeForgotPassword)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired OTP"})
		return
	}

	err = c.userService.ChangePassword(user.ID, req.NewPassword)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Password reset successfully"})
}

func (c *UserController) GetAllUsers(ctx *gin.Context) {
	users, err := c.userService.GetAllUsers()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"users": users})
}

func (c *UserController) GetMe(ctx *gin.Context) {
	userId, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	sessionInfo, err := c.userService.GetMeSession(userId.(string))
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, sessionInfo)
}

func (c *UserController) GetUserByID(ctx *gin.Context) {
	userId := ctx.Param("userId")
	profile, err := c.userService.GetUserFullProfile(userId)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, profile)
}

func (c *UserController) UpdateUser(ctx *gin.Context) {
	userId := ctx.Param("userId")
	callerID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Authorization Check: Users can only update themselves
	if callerID.(string) != userId {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "Forbidden: You can only modify your own data"})
		return
	}

	// We'll use the user model directly here for partial updates
	var inputData models.User
	if err := ctx.ShouldBindJSON(&inputData); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := c.userService.UpdateUser(userId, &inputData)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "User updated successfully"})
}

func (c *UserController) ChangePassword(ctx *gin.Context) {
	// The User ID should ideally be extracted from the JWT token via middleware
	userId, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req ChangePasswordReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := c.userService.ChangePassword(userId.(string), req.NewPassword)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Password changed successfully"})
}

func (c *UserController) SoftDeleteUser(ctx *gin.Context) {
	userId := ctx.Param("userId")
	callerID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Authorization Check: Users can only delete themselves
	if callerID.(string) != userId {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "Forbidden: You can only delete your own account"})
		return
	}

	err := c.userService.SoftDelete(userId)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "User soft deleted successfully"})
}

func (c *UserController) HardDeleteUser(ctx *gin.Context) {
	userId := ctx.Param("userId")
	callerID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Authorization Check: Users can only delete themselves (even for hard delete)
	if callerID.(string) != userId {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "Forbidden: You can only delete your own account"})
		return
	}

	err := c.userService.HardDelete(userId)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "User and all related data completely wiped successfully"})
}

func (c *UserController) CheckUsernameExists(ctx *gin.Context) {
	username := ctx.Query("username")
	if username == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "username query parameter is required"})
		return
	}

	exists, err := c.userService.UsernameExists(username)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"exists": exists})
}
