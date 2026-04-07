package controllers

import (
	"net/http"

	"github.com/TriStrac/Scarrow-Go-API/internal/service"
	"github.com/gin-gonic/gin"
)

type ActivityLogController struct {
	activityLogService service.ActivityLogService
}

func NewActivityLogController(activityLogService service.ActivityLogService) *ActivityLogController {
	return &ActivityLogController{activityLogService: activityLogService}
}

func (c *ActivityLogController) GetMyLogs(ctx *gin.Context) {
	callerID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	logs, err := c.activityLogService.GetLogsByUserID(callerID.(string))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"logs": logs})
}

func (c *ActivityLogController) GetAllLogs(ctx *gin.Context) {
	// Usually admin only, but for now we allow anyone with a token
	logs, err := c.activityLogService.GetAllLogs()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"logs": logs})
}
