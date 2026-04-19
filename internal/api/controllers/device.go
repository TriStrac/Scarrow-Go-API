package controllers

import (
	"net/http"
	"strconv"

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

type RegisterHubReq struct {
	Name        string   `json:"name" binding:"required"`
	LocationLat *float64 `json:"location_lat"`
	LocationLng *float64 `json:"location_lng"`
}

type RegisterNodeReq struct {
	HubID    string `json:"hub_id" binding:"required"`
	NodeType string `json:"node_type" binding:"required"`
	Label    string `json:"label" binding:"required"`
}

type CreateDeviceReq struct {
	Name       string            `json:"name" binding:"required"`
	OwnerType  string            `json:"owner_type" binding:"required,oneof=USER GROUP"`
	DeviceType models.DeviceType `json:"device_type" binding:"required,oneof=CENTRAL NODE"`
	ParentID   *string           `json:"parent_id"`
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
	LogType         string  `json:"log_type" binding:"required"`
	Payload         string  `json:"payload" binding:"required"`
	PestType        string  `json:"pest_type"`
	FrequencyHz     float64 `json:"frequency_hz"`
	DurationSeconds int     `json:"duration_seconds"`
}

func (c *DeviceController) RegisterHub(ctx *gin.Context) {
	var req RegisterHubReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	callerID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	hub, err := c.deviceService.RegisterHub(req.Name, callerID.(string), req.LocationLat, req.LocationLng)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"hub_id": hub.ID,
		"secret": hub.Secret,
		"status": hub.Status,
	})
}

func (c *DeviceController) RegisterNode(ctx *gin.Context) {
	var req RegisterNodeReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	callerID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	node, err := c.deviceService.RegisterNode(req.Label, callerID.(string), req.HubID, req.NodeType)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"node_id":     node.ID,
		"node_secret": node.Secret,
		"hub_filter":  req.HubID,
		"status":      node.Status,
	})
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

	device, err := c.deviceService.CreateDevice(req.Name, callerID.(string), req.OwnerType, req.DeviceType, req.ParentID)
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

	err = c.deviceService.CreateLog(deviceID, req.LogType, req.Payload, req.PestType, req.FrequencyHz, req.DurationSeconds)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{"message": "Device log created successfully"})
}

func (c *DeviceController) GetLogs(ctx *gin.Context) {
	deviceID := ctx.Param("deviceId")
	
	// Default to 50 logs per page
	limit := 50
	offset := 0

	if limitStr := ctx.Query("limit"); limitStr != "" {
		if parsed, err := strconv.Atoi(limitStr); err == nil && parsed > 0 {
			limit = parsed
		}
	}
	
	if offsetStr := ctx.Query("offset"); offsetStr != "" {
		if parsed, err := strconv.Atoi(offsetStr); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	logs, err := c.deviceService.GetLogsByDeviceID(deviceID, limit, offset)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"logs": logs, "limit": limit, "offset": offset})
}

func (c *DeviceController) GetMyDevices(ctx *gin.Context) {
	callerID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	devices, err := c.deviceService.GetMyDevices(callerID.(string))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"devices": devices})
}
