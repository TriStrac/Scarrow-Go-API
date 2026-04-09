package controllers

import (
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
	FirstName string `json:"first_name" binding:"required"`
	LastName  string `json:"last_name" binding:"required"`
	Username  string `json:"username" binding:"required"`
	Password  string `json:"password" binding:"required,min=6"`
	Number    string `json:"number" binding:"required"` // Phone number for OTP
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

type ResetPasswordReq struct {
	Username    string `json:"username" binding:"required"`
	OTP         string `json:"otp" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

type ChangePasswordReq struct {
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

func (c *UserController) Register(ctx *gin.Context) {
	var req RegisterReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user := &models.User{
		Username: req.Username,
		Password: req.Password,
		Profile: &models.UserProfile{
			FirstName:   req.FirstName,
			LastName:    req.LastName,
			PhoneNumber: req.Number,
		},
		IsVerified: false,
	}

	_, err := c.userService.Register(user)
	if err != nil {
		ctx.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}

	// Send OTP for registration
	otp, err := c.otpService.GenerateAndSendOTP(req.Number, models.OTPPurposeRegistration)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send OTP: " + err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"message":    "Registration initiated. Please verify with the OTP sent to your number.",
		"identifier": req.Username,
		"otp":        otp, // Included for testing purposes
	})
}

func (c *UserController) VerifyRegistration(ctx *gin.Context) {
	var req VerifyOTPReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 1. Get user to find their phone number
	user, err := c.userService.FindByUsername(req.Identifier)
	if err != nil || user == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// 2. Verify OTP
	success, err := c.otpService.VerifyOTP(user.Profile.PhoneNumber, req.Code, models.OTPPurposeRegistration)
	if err != nil || !success {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired OTP"})
		return
	}

	// 3. Mark user as verified
	err = c.userService.VerifyUser(req.Identifier)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "User verified successfully. You can now login."})
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
			// Resend OTP if not verified?
			// For now just return error as per usual flow
			ctx.JSON(http.StatusForbidden, gin.H{"error": "User is not verified. Please verify your account first."})
			return
		}
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	// Send OTP for login
	otp, err := c.otpService.GenerateAndSendOTP(user.Profile.PhoneNumber, models.OTPPurposeLogin)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send OTP: " + err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message":    "OTP sent for login verification",
		"identifier": user.Username,
		"otp":        otp, // Included for testing purposes
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

	success, err := c.otpService.VerifyOTP(user.Profile.PhoneNumber, req.Code, models.OTPPurposeLogin)
	if err != nil || !success {
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
		// Don't reveal if user exists for security, but user specifically asked for forgot password flow
		// Let's reveal it for now as per requirement
		ctx.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	otp, err := c.otpService.GenerateAndSendOTP(user.Profile.PhoneNumber, models.OTPPurposeForgotPassword)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send OTP"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "OTP sent for password reset",
		"otp":     otp, // Included for testing purposes
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

	success, err := c.otpService.VerifyOTP(user.Profile.PhoneNumber, req.OTP, models.OTPPurposeForgotPassword)
	if err != nil || !success {
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
