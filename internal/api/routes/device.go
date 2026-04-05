package routes

import (
	"github.com/TriStrac/Scarrow-Go-API/internal/api/controllers"
	"github.com/TriStrac/Scarrow-Go-API/internal/api/middlewares"
	"github.com/gin-gonic/gin"
)

func RegisterDeviceRoutes(r *gin.RouterGroup, controller *controllers.DeviceController) {
	deviceGroup := r.Group("/device")
	deviceGroup.Use(middlewares.AuthMiddleware())
	{
		deviceGroup.POST("/", controller.CreateDevice)
		deviceGroup.GET("/", controller.GetAllDevices)
		deviceGroup.GET("/my", controller.GetMyDevices)
		deviceGroup.GET("/:deviceId", controller.GetDeviceByID)
		deviceGroup.PATCH("/:deviceId", controller.UpdateDevice)
		deviceGroup.DELETE("/:deviceId", controller.SoftDeleteDevice)

		// Ownership
		deviceGroup.POST("/:deviceId/owner", controller.AddOwner)
		deviceGroup.DELETE("/:deviceId/owner", controller.RemoveOwner)
		deviceGroup.GET("/:deviceId/owners", controller.GetOwners)

		// Logs
		deviceGroup.POST("/:deviceId/logs", controller.CreateLog)
		deviceGroup.GET("/:deviceId/logs", controller.GetLogs)
	}
}
