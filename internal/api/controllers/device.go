package controllers

import (
	"net/http"

	"github.com/TriStrac/Scarrow-Go-API/internal/models"
	"github.com/TriStrac/Scarrow-Go-API/internal/service"
	"github.com/gin-gonic/gin"
)

type DeviceController struct {
	deviceService service.DeviceService
}

func NewDeviceController(deviceService service.DeviceService) *DeviceController {
	return &DeviceController{deviceService: deviceService}
}

type CreateDeviceReq struct {
	Name      string `json:"name" binding:"required"`
	OwnerType string `json:"owner_type" binding:"required,oneof=USER GROUP"`
}

type UpdateDeviceReq struct {
	Name   string `json:"name"`
	Status string `json:"status"`
}

type AddOwnerReq struct {
	OwnerID   string `json:"owner_id" binding:"required"`
	OwnerType string `json:"owner_type" binding:"required,oneof=USER GROUP"`
}

type RemoveOwnerReq struct {
	OwnerID   string `json:"owner_id" binding:"required"`
	OwnerType string `json:"owner_type" binding:"required,oneof=USER GROUP"`
}

type CreateDeviceLogReq struct {
	LogType string `json:"log_type" binding:"required"`
	Payload string `json:"payload" binding:"required"`
}

func (c *DeviceController) CreateDevice(ctx *gin.Context) {
	var req CreateDeviceReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	callerID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	ownerID := callerID.(string)
	// If owner_type is GROUP, user MUST specify the group_id (though not in my simplified req)
	// For now, let's just assume the user is the owner if USER, and we'd need another field if they want it for a group.
	// But let's keep it simple: the creator can specify the owner if they have permissions.
	// However, for strict IDOR, creator should be the initial owner.

	device, err := c.deviceService.CreateDevice(req.Name, ownerID, req.OwnerType)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{"message": "Device created successfully", "device": device})
}

func (c *DeviceController) GetAllDevices(ctx *gin.Context) {
	devices, err := c.deviceService.GetAllDevices()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"devices": devices})
}

func (c *DeviceController) GetDeviceByID(ctx *gin.Context) {
	deviceID := ctx.Param("deviceId")
	device, err := c.deviceService.GetDeviceByID(deviceID)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"device": device})
}

func (c *DeviceController) UpdateDevice(ctx *gin.Context) {
	deviceID := ctx.Param("deviceId")
	callerID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Authorization Check: Is caller an owner?
	isOwner, err := c.deviceService.IsOwner(deviceID, callerID.(string))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if !isOwner {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "Forbidden: You are not an owner of this device"})
		return
	}

	var req UpdateDeviceReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = c.deviceService.UpdateDevice(deviceID, req.Name, req.Status)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Device updated successfully"})
}

func (c *DeviceController) SoftDeleteDevice(ctx *gin.Context) {
	deviceID := ctx.Param("deviceId")
	callerID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	isOwner, err := c.deviceService.IsOwner(deviceID, callerID.(string))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if !isOwner {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "Forbidden: You are not an owner of this device"})
		return
	}

	err = c.deviceService.SoftDelete(deviceID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Device soft deleted successfully"})
}

// Ownership Endpoints
func (c *DeviceController) AddOwner(ctx *gin.Context) {
	deviceID := ctx.Param("deviceId")
	callerID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	isOwner, err := c.deviceService.IsOwner(deviceID, callerID.(string))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if !isOwner {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "Forbidden: You are not an owner of this device"})
		return
	}

	var req AddOwnerReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = c.deviceService.AddOwner(deviceID, req.OwnerID, req.OwnerType)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Owner added successfully"})
}

func (c *DeviceController) RemoveOwner(ctx *gin.Context) {
	deviceID := ctx.Param("deviceId")
	callerID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	isOwner, err := c.deviceService.IsOwner(deviceID, callerID.(string))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if !isOwner {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "Forbidden: You are not an owner of this device"})
		return
	}

	var req RemoveOwnerReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = c.deviceService.RemoveOwner(deviceID, req.OwnerID, req.OwnerType)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Owner removed successfully"})
}

func (c *DeviceController) GetOwners(ctx *gin.Context) {
	deviceID := ctx.Param("deviceId")
	owners, err := c.deviceService.GetOwnersByDeviceID(deviceID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"owners": owners})
}

// Device Logging Endpoints
func (c *DeviceController) CreateLog(ctx *gin.Context) {
	deviceID := ctx.Param("deviceId")
	// Maybe we should check ownership here too?
	// Or maybe devices themselves log without being "logged in" as a user?
	// For now let's keep ownership check.
	callerID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	isOwner, err := c.deviceService.IsOwner(deviceID, callerID.(string))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if !isOwner {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "Forbidden: You are not an owner of this device"})
		return
	}

	var req CreateDeviceLogReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = c.deviceService.CreateLog(deviceID, req.LogType, req.Payload)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{"message": "Device log created successfully"})
}

func (c *DeviceController) GetLogs(ctx *gin.Context) {
	deviceID := ctx.Param("deviceId")
	logs, err := c.deviceService.GetLogsByDeviceID(deviceID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"logs": logs})
}

func (c *DeviceController) GetMyDevices(ctx *gin.Context) {
	callerID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	devices, err := c.deviceService.GetDevicesByOwner(callerID.(string), "USER")
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"devices": devices})
}
